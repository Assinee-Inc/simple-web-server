package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	librarysvc "github.com/Assinee-Inc/docffy/internal/library/service"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	salesvc "github.com/Assinee-Inc/docffy/internal/sales/service"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type PurchaseSalesHandler struct {
	purchaseService           salesvc.PurchaseService
	sessionService            authsvc.SessionService
	creatorService            accountsvc.CreatorService
	ebookService              librarysvc.EbookService
	resendDownloadLinkService salesvc.ResendDownloadLinkServiceInterface
	transactionService        salesvc.TransactionService
}

func NewPurchaseSalesHandler(deps ...any) *PurchaseSalesHandler {
	h := &PurchaseSalesHandler{}
	for _, dep := range deps {
		switch v := dep.(type) {
		case salesvc.PurchaseService:
			h.purchaseService = v
		case authsvc.SessionService:
			h.sessionService = v
		case accountsvc.CreatorService:
			h.creatorService = v
		case librarysvc.EbookService:
			h.ebookService = v
		case salesvc.ResendDownloadLinkServiceInterface:
			h.resendDownloadLinkService = v
		case salesvc.TransactionService:
			h.transactionService = v
		}
	}
	return h
}

func (h *PurchaseSalesHandler) resolveCreator(r *http.Request) (*accountmodel.Creator, int, string) {
	if creator := accountmw.CreatorFromContext(r.Context()); creator != nil {
		return creator, http.StatusOK, ""
	}

	userEmail := authmw.GetAuthenticatedUserEmail(r.Context())
	if userEmail == "" && h.sessionService != nil {
		var err error
		userEmail, err = h.sessionService.GetUserEmailFromSession(r)
		if err != nil || userEmail == "" {
			return nil, http.StatusUnauthorized, "Sessão inválida"
		}
	}
	if userEmail == "" {
		return nil, http.StatusUnauthorized, "Sessão inválida"
	}

	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil || creator == nil {
		return nil, http.StatusNotFound, "Criador não encontrado"
	}
	return creator, http.StatusOK, ""
}

func (h *PurchaseSalesHandler) extractPurchaseID(r *http.Request) string {
	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			PurchaseID string `json:"purchase_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.PurchaseID != "" {
			return body.PurchaseID
		}
	}
	return r.FormValue("purchase_id")
}

func (h *PurchaseSalesHandler) PurchaseSalesList(w http.ResponseWriter, r *http.Request) {
	creator, status, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, status, "AUTH_ERROR", errMsg)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit := 10

	ebookPublicID := r.URL.Query().Get("ebook_id")
	clientName := r.URL.Query().Get("client_name")
	clientEmail := r.URL.Query().Get("client_email")

	var ebookID uint
	if ebookPublicID != "" {
		ebook, err := h.ebookService.FindByPublicID(ebookPublicID)
		if err == nil && ebook != nil {
			ebookID = ebook.ID
		}
	}

	var ebookIDPtr *uint
	if ebookID > 0 {
		ebookIDPtr = &ebookID
	}

	purchases, total, err := h.purchaseService.GetPurchasesByCreatorIDWithFilters(
		creator.ID, ebookIDPtr, clientName, clientEmail, page, limit,
	)
	if err != nil {
		slog.Error("Erro ao buscar vendas", "error", err)
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar vendas")
		return
	}

	purchaseTransactionMap := make(map[uint]*salesmodel.Transaction)
	for _, purchase := range purchases {
		transaction, err := h.transactionService.FindTransactionByPurchaseID(purchase.ID)
		if err == nil && transaction != nil {
			purchaseTransactionMap[purchase.ID] = transaction
		}
	}

	ebooks, err := h.ebookService.GetEbooksByCreatorID(creator.ID)
	if err != nil {
		slog.Error("Erro ao buscar ebooks para filtro", "error", err)
		ebooks = []*librarymodel.Ebook{}
	}

	pagination := salesmodel.NewPagination(page, limit)
	pagination.SetTotal(total)

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":                  true,
		"creator":                  creator,
		"purchases":                purchases,
		"purchase_transaction_map": purchaseTransactionMap,
		"pagination":               pagination,
		"ebooks":                   ebooks,
		"ebook_id":                 ebookPublicID,
		"client_name":              clientName,
		"client_email":             clientEmail,
		"record_type":              "vendas",
		"filters": map[string]any{
			"client_name":  clientName,
			"client_email": clientEmail,
			"ebook_id":     ebookPublicID,
		},
	})
}

func (h *PurchaseSalesHandler) BlockDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		web.RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método não permitido")
		return
	}

	purchasePublicID := h.extractPurchaseID(r)
	if purchasePublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_PURCHASE_ID", "ID de purchase inválido")
		return
	}

	creator, status, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, status, "AUTH_ERROR", errMsg)
		return
	}

	purchase, err := h.purchaseService.GetPurchaseByPublicID(purchasePublicID)
	if err != nil || purchase == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Venda não encontrada")
		return
	}

	if purchase.Ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	err = h.purchaseService.BlockDownload(purchase.ID, creator.ID, true)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno do servidor")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":          true,
		"download_blocked": true,
		"purchase_id":      purchasePublicID,
	})
}

func (h *PurchaseSalesHandler) UnblockDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		web.RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método não permitido")
		return
	}

	purchasePublicID := h.extractPurchaseID(r)
	if purchasePublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_PURCHASE_ID", "ID de purchase inválido")
		return
	}

	creator, status, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, status, "AUTH_ERROR", errMsg)
		return
	}

	purchase, err := h.purchaseService.GetPurchaseByPublicID(purchasePublicID)
	if err != nil || purchase == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Venda não encontrada")
		return
	}

	if purchase.Ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	err = h.purchaseService.BlockDownload(purchase.ID, creator.ID, false)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno do servidor")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":          true,
		"download_blocked": false,
		"purchase_id":      purchasePublicID,
	})
}

func (h *PurchaseSalesHandler) ResendDownloadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		web.RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método não permitido")
		return
	}

	purchasePublicID := h.extractPurchaseID(r)
	if purchasePublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_PURCHASE_ID", "ID de purchase inválido")
		return
	}

	var newEmail string

	creator, status, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, status, "AUTH_ERROR", errMsg)
		return
	}

	purchase, err := h.purchaseService.GetPurchaseByPublicID(purchasePublicID)
	if err != nil || purchase == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Venda não encontrada")
		return
	}

	if purchase.Ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	err = h.resendDownloadLinkService.ResendDownloadLinkByPurchaseID(purchase.ID, newEmail)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno do servidor")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"message":     fmt.Sprintf("Link de download reenviado com sucesso para a compra %s", purchasePublicID),
		"purchase_id": purchasePublicID,
	})
}
