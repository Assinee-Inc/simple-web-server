package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type TransactionHandler struct {
	transactionService        salesvc.TransactionService
	sessionService            authsvc.SessionService
	creatorService            accountsvc.CreatorService
	resendDownloadLinkService salesvc.ResendDownloadLinkServiceInterface
	templateRenderer          template.TemplateRenderer
}

func NewTransactionHandler(
	transactionService salesvc.TransactionService,
	sessionService authsvc.SessionService,
	creatorService accountsvc.CreatorService,
	resendDownloadLinkService salesvc.ResendDownloadLinkServiceInterface,
	templateRenderer template.TemplateRenderer,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService:        transactionService,
		sessionService:            sessionService,
		creatorService:            creatorService,
		resendDownloadLinkService: resendDownloadLinkService,
		templateRenderer:          templateRenderer,
	}
}

// TransactionList exibe a lista de transações do criador
func (h *TransactionHandler) TransactionList(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Erro ao buscar transações", http.StatusInternalServerError)
		return
	}

	pagination := salesmodel.NewPagination(page, limit)
	pagination.SetTotal(total)
	pagination.SearchTerm = search

	h.templateRenderer.View(w, r, "transactions/list", map[string]interface{}{
		"Creator":      creator,
		"Transactions": transactions,
		"Pagination":   pagination,
		"Search":       search,
		"Status":       status,
		"Filters": map[string]interface{}{
			"search": search,
			"status": status,
		},
	}, "admin-daisy")
}

// TransactionDetail exibe detalhes de uma transação
func (h *TransactionHandler) TransactionDetail(w http.ResponseWriter, r *http.Request) {
	transactionPublicID := r.URL.Query().Get("id")
	if transactionPublicID == "" {
		slog.Error("ID de transação não fornecido")
		http.Error(w, "ID de transação inválido", http.StatusBadRequest)
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

	transaction, err := h.transactionService.GetTransactionByPublicID(transactionPublicID)
	if err != nil {
		slog.Error("Erro ao buscar transação", "error", err)
		http.Error(w, "Transação não encontrada", http.StatusNotFound)
		return
	}

	if transaction.CreatorID != creator.ID {
		slog.Warn("Tentativa de acesso não autorizado a transação",
			"transactionPublicID", transactionPublicID,
			"creatorID", creator.ID,
			"ownerID", transaction.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	h.templateRenderer.View(w, r, "transactions/detail", map[string]interface{}{
		"Creator":     creator,
		"Transaction": transaction,
	}, "admin-daisy")
}

// ResendDownloadLink reenvia o link de download para o cliente
func (h *TransactionHandler) ResendDownloadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	transactionPublicID := r.FormValue("transaction_id")
	if transactionPublicID == "" {
		slog.Error("ID de transação não fornecido")
		http.Error(w, "ID de transação inválido", http.StatusBadRequest)
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

	transaction, err := h.transactionService.GetTransactionByPublicID(transactionPublicID)
	if err != nil {
		slog.Error("Erro ao buscar transação", "error", err)
		http.Error(w, "Transação não encontrada", http.StatusNotFound)
		return
	}

	if transaction.CreatorID != creator.ID {
		slog.Warn("Tentativa de reenvio não autorizado",
			"transactionPublicID", transactionPublicID,
			"creatorID", creator.ID,
			"ownerID", transaction.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	if transaction.Status != salesmodel.TransactionStatusCompleted {
		slog.Error("Tentativa de reenvio para transação não completada",
			"transactionPublicID", transactionPublicID,
			"status", transaction.Status)
		http.Error(w, "Não é possível reenviar link para transação não completada", http.StatusBadRequest)
		return
	}

	err = h.resendDownloadLinkService.ResendDownloadLinkByTransactionID(transaction.ID)
	if err != nil {
		slog.Error("Erro ao reenviar link de download", "error", err, "transactionID", transaction.ID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Link de download reenviado com sucesso",
		"transactionPublicID", transactionPublicID,
		"creatorID", creator.ID,
		"creatorEmail", creator.Email)

	http.Redirect(w, r, fmt.Sprintf("/transactions?success=download_link_resent&transaction_id=%s", transactionPublicID), http.StatusSeeOther)
}
