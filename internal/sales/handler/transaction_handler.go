package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	salesvc "github.com/Assinee-Inc/docffy/internal/sales/service"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type TransactionHandler struct {
	transactionService        salesvc.TransactionService
	sessionService            authsvc.SessionService
	creatorService            accountsvc.CreatorService
	resendDownloadLinkService salesvc.ResendDownloadLinkServiceInterface
}

func NewTransactionHandler(
	transactionService salesvc.TransactionService,
	sessionService authsvc.SessionService,
	creatorService accountsvc.CreatorService,
	resendDownloadLinkService salesvc.ResendDownloadLinkServiceInterface,
	_ ...any,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService:        transactionService,
		sessionService:            sessionService,
		creatorService:            creatorService,
		resendDownloadLinkService: resendDownloadLinkService,
	}
}

func (h *TransactionHandler) resolveCreator(r *http.Request) (*accountmodel.Creator, int, string) {
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

func (h *TransactionHandler) TransactionList(w http.ResponseWriter, r *http.Request) {
	creator, statusCode, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, statusCode, "AUTH_ERROR", errMsg)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit := 10

	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	var transactions []*salesmodel.Transaction
	var total int64

	if search != "" || (status != "" && status != "Todos") {
		transactions, total, err = h.transactionService.GetTransactionsByCreatorIDWithFilters(creator.ID, page, limit, search, status)
	} else {
		transactions, total, err = h.transactionService.GetTransactionsByCreatorID(creator.ID, page, limit)
	}

	if err != nil {
		slog.Error("Erro ao buscar transações", "error", err)
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao buscar transações")
		return
	}

	pagination := salesmodel.NewPagination(page, limit)
	pagination.SetTotal(total)
	pagination.SearchTerm = search

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":      true,
		"creator":      creator,
		"transactions": transactions,
		"pagination":   pagination,
		"search":       search,
		"status":       status,
		"filters": map[string]any{
			"search": search,
			"status": status,
		},
	})
}

func (h *TransactionHandler) TransactionDetail(w http.ResponseWriter, r *http.Request) {
	transactionPublicID := r.URL.Query().Get("id")
	if transactionPublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "ID de transação inválido")
		return
	}

	creator, statusCode, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, statusCode, "AUTH_ERROR", errMsg)
		return
	}

	transaction, err := h.transactionService.GetTransactionByPublicID(transactionPublicID)
	if err != nil || transaction == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Transação não encontrada")
		return
	}

	if transaction.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"creator":     creator,
		"transaction": transaction,
	})
}

func (h *TransactionHandler) ResendDownloadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		web.RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método não permitido")
		return
	}

	var transactionPublicID string
	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			TransactionID string `json:"transaction_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		transactionPublicID = body.TransactionID
	} else {
		transactionPublicID = r.FormValue("transaction_id")
	}

	if transactionPublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "ID de transação inválido")
		return
	}

	creator, statusCode, errMsg := h.resolveCreator(r)
	if creator == nil {
		web.RespondError(w, statusCode, "AUTH_ERROR", errMsg)
		return
	}

	transaction, err := h.transactionService.GetTransactionByPublicID(transactionPublicID)
	if err != nil || transaction == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Transação não encontrada")
		return
	}

	if transaction.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	if transaction.Status != salesmodel.TransactionStatusCompleted {
		web.RespondError(w, http.StatusBadRequest, "TRANSACTION_NOT_COMPLETED", "Não é possível reenviar link para transação não completada")
		return
	}

	err = h.resendDownloadLinkService.ResendDownloadLinkByTransactionID(transaction.ID)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno do servidor")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":        true,
		"message":        "Link de download reenviado com sucesso",
		"transaction_id": transactionPublicID,
	})
}
