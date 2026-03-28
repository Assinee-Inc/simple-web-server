package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	librarysvc "github.com/anglesson/simple-web-server/internal/library/service"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type PurchaseSalesHandler struct {
	templateRenderer          template.TemplateRenderer
	purchaseService           salesvc.PurchaseService
	sessionService            authsvc.SessionService
	creatorService            accountsvc.CreatorService
	ebookService              librarysvc.EbookService
	resendDownloadLinkService salesvc.ResendDownloadLinkServiceInterface
	transactionService        salesvc.TransactionService
}

func NewPurchaseSalesHandler(
	templateRenderer template.TemplateRenderer,
	purchaseService salesvc.PurchaseService,
	sessionService authsvc.SessionService,
	creatorService accountsvc.CreatorService,
	ebookService librarysvc.EbookService,
	resendDownloadLinkService salesvc.ResendDownloadLinkServiceInterface,
	transactionService salesvc.TransactionService,
) *PurchaseSalesHandler {
	return &PurchaseSalesHandler{
		templateRenderer:          templateRenderer,
		purchaseService:           purchaseService,
		sessionService:            sessionService,
		creatorService:            creatorService,
		ebookService:              ebookService,
		resendDownloadLinkService: resendDownloadLinkService,
		transactionService:        transactionService,
	}
}

// PurchaseSalesList exibe a lista de vendas (purchases) do criador
func (h *PurchaseSalesHandler) PurchaseSalesList(w http.ResponseWriter, r *http.Request) {
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		slog.Error("Erro ao obter email da sessão", "error", err)
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
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
		http.Error(w, "Erro ao buscar vendas", http.StatusInternalServerError)
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

	h.templateRenderer.View(w, r, "purchase/list", map[string]interface{}{
		"Creator":                creator,
		"Purchases":              purchases,
		"PurchaseTransactionMap": purchaseTransactionMap,
		"Pagination":             pagination,
		"Ebooks":                 ebooks,
		"EbookID":                ebookPublicID,
		"ClientName":             clientName,
		"ClientEmail":            clientEmail,
		"RecordType":             "vendas",
		"Filters": map[string]interface{}{
			"client_name":  clientName,
			"client_email": clientEmail,
			"ebook_id":     ebookPublicID,
		},
	}, "admin-daisy")
}

// BlockDownload bloqueia o download de um cliente específico
func (h *PurchaseSalesHandler) BlockDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	purchasePublicID := r.FormValue("purchase_id")
	if purchasePublicID == "" {
		slog.Error("ID de purchase não fornecido")
		http.Error(w, "ID de purchase inválido", http.StatusBadRequest)
		return
	}

	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		slog.Error("Erro ao obter email da sessão", "error", err)
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	purchase, err := h.purchaseService.GetPurchaseByPublicID(purchasePublicID)
	if err != nil {
		slog.Error("Erro ao buscar purchase", "error", err)
		http.Error(w, "Venda não encontrada", http.StatusNotFound)
		return
	}

	if purchase.Ebook.CreatorID != creator.ID {
		slog.Warn("Tentativa de bloqueio não autorizado",
			"purchasePublicID", purchasePublicID,
			"creatorID", creator.ID,
			"ownerID", purchase.Ebook.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	err = h.purchaseService.BlockDownload(purchase.ID, creator.ID, true)
	if err != nil {
		slog.Error("Erro ao bloquear download", "error", err, "purchaseID", purchase.ID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Download bloqueado com sucesso",
		"purchasePublicID", purchasePublicID,
		"creatorID", creator.ID,
		"clientID", purchase.ClientID)

	http.Redirect(w, r, "/purchase/sales?success=download_blocked", http.StatusSeeOther)
}

// UnblockDownload desbloqueia o download de um cliente específico
func (h *PurchaseSalesHandler) UnblockDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	purchasePublicID := r.FormValue("purchase_id")
	if purchasePublicID == "" {
		slog.Error("ID de purchase não fornecido")
		http.Error(w, "ID de purchase inválido", http.StatusBadRequest)
		return
	}

	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		slog.Error("Erro ao obter email da sessão", "error", err)
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	purchase, err := h.purchaseService.GetPurchaseByPublicID(purchasePublicID)
	if err != nil {
		slog.Error("Erro ao buscar purchase", "error", err)
		http.Error(w, "Venda não encontrada", http.StatusNotFound)
		return
	}

	if purchase.Ebook.CreatorID != creator.ID {
		slog.Warn("Tentativa de desbloqueio não autorizado",
			"purchasePublicID", purchasePublicID,
			"creatorID", creator.ID,
			"ownerID", purchase.Ebook.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	err = h.purchaseService.BlockDownload(purchase.ID, creator.ID, false)
	if err != nil {
		slog.Error("Erro ao desbloquear download", "error", err, "purchaseID", purchase.ID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Download desbloqueado com sucesso",
		"purchasePublicID", purchasePublicID,
		"creatorID", creator.ID,
		"clientID", purchase.ClientID)

	http.Redirect(w, r, "/purchase/sales?success=download_unblocked", http.StatusSeeOther)
}

// ResendDownloadLink reenvia o link de download com opção de novo email
func (h *PurchaseSalesHandler) ResendDownloadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	purchasePublicID := r.FormValue("purchase_id")
	if purchasePublicID == "" {
		slog.Error("ID de purchase não fornecido")
		http.Error(w, "ID de purchase inválido", http.StatusBadRequest)
		return
	}

	// newEmail := r.FormValue("new_email")
	var newEmail string // Desabilitado temporariamente

	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		slog.Error("Erro ao obter email da sessão", "error", err)
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	purchase, err := h.purchaseService.GetPurchaseByPublicID(purchasePublicID)
	if err != nil {
		slog.Error("Erro ao buscar purchase", "error", err)
		http.Error(w, "Venda não encontrada", http.StatusNotFound)
		return
	}

	if purchase.Ebook.CreatorID != creator.ID {
		slog.Warn("Tentativa de reenvio não autorizado",
			"purchasePublicID", purchasePublicID,
			"creatorID", creator.ID,
			"ownerID", purchase.Ebook.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	err = h.resendDownloadLinkService.ResendDownloadLinkByPurchaseID(purchase.ID, newEmail)
	if err != nil {
		slog.Error("Erro ao reenviar link de download", "error", err, "purchaseID", purchase.ID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Link de download reenviado com sucesso",
		"purchasePublicID", purchasePublicID,
		"creatorID", creator.ID,
		"newEmail", newEmail)

	http.Redirect(w, r, fmt.Sprintf("/purchase/sales?success=download_link_resent&purchase_id=%s", purchasePublicID), http.StatusSeeOther)
}
