package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type TransactionHandler struct {
	transactionService        service.TransactionService
	sessionService            service.SessionService
	creatorService            service.CreatorService
	resendDownloadLinkService *service.ResendDownloadLinkService
	templateRenderer          template.TemplateRenderer
}

func NewTransactionHandler(
	transactionService service.TransactionService,
	sessionService service.SessionService,
	creatorService service.CreatorService,
	resendDownloadLinkService *service.ResendDownloadLinkService,
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
	// Obter usuário da sessão
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		slog.Error("Erro ao obter email da sessão", "error", err)
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Buscar criador
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	// Obter parâmetros de paginação e filtros
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit := 10

	// Obter parâmetros de busca e filtros
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	// Buscar transações com filtros
	var transactions []*models.Transaction
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

	// Calcular paginação usando o modelo padrão do projeto
	pagination := models.NewPagination(page, limit)
	pagination.SetTotal(total)
	pagination.SearchTerm = search

	// Renderizar template
	h.templateRenderer.View(w, r, "transactions/list", map[string]interface{}{
		"Creator":      creator,
		"Transactions": transactions,
		"Pagination":   pagination,
		"Search":       search,
		"Status":       status,
	}, "admin")
}

// TransactionDetail exibe detalhes de uma transação
func (h *TransactionHandler) TransactionDetail(w http.ResponseWriter, r *http.Request) {
	// Obter ID da transação
	transactionID, err := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)
	if err != nil {
		slog.Error("ID de transação inválido", "error", err)
		http.Error(w, "ID de transação inválido", http.StatusBadRequest)
		return
	}

	// Obter usuário da sessão
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		slog.Error("Erro ao obter email da sessão", "error", err)
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Buscar criador
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	// Buscar transação
	transaction, err := h.transactionService.GetTransactionByID(uint(transactionID))
	if err != nil {
		slog.Error("Erro ao buscar transação", "error", err)
		http.Error(w, "Transação não encontrada", http.StatusNotFound)
		return
	}

	// Verificar se a transação pertence ao criador
	if transaction.CreatorID != creator.ID {
		slog.Warn("Tentativa de acesso não autorizado a transação",
			"transactionID", transactionID,
			"creatorID", creator.ID,
			"ownerID", transaction.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	// Renderizar template
	h.templateRenderer.View(w, r, "transactions/detail", map[string]interface{}{
		"Creator":     creator,
		"Transaction": transaction,
	}, "admin")
}

// ResendDownloadLink reenvia o link de download para o cliente
func (h *TransactionHandler) ResendDownloadLink(w http.ResponseWriter, r *http.Request) {
	// Verificar método HTTP
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obter ID da transação
	transactionID, err := strconv.ParseUint(r.FormValue("transaction_id"), 10, 32)
	if err != nil {
		slog.Error("ID de transação inválido", "error", err)
		http.Error(w, "ID de transação inválido", http.StatusBadRequest)
		return
	}

	// Obter usuário da sessão
	userEmail, err := h.sessionService.GetUserEmailFromSession(r)
	if err != nil {
		slog.Error("Erro ao obter email da sessão", "error", err)
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Buscar criador
	creator, err := h.creatorService.FindCreatorByEmail(userEmail)
	if err != nil {
		slog.Error("Erro ao buscar criador", "error", err)
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	// Buscar transação para verificar permissões
	transaction, err := h.transactionService.GetTransactionByID(uint(transactionID))
	if err != nil {
		slog.Error("Erro ao buscar transação", "error", err)
		http.Error(w, "Transação não encontrada", http.StatusNotFound)
		return
	}

	// Verificar se a transação pertence ao criador
	if transaction.CreatorID != creator.ID {
		slog.Warn("Tentativa de reenvio não autorizado",
			"transactionID", transactionID,
			"creatorID", creator.ID,
			"ownerID", transaction.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	// Verificar se a transação está completa
	if transaction.Status != models.TransactionStatusCompleted {
		slog.Error("Tentativa de reenvio para transação não completada",
			"transactionID", transactionID,
			"status", transaction.Status)
		http.Error(w, "Não é possível reenviar link para transação não completada", http.StatusBadRequest)
		return
	}

	// Reenviar o link
	err = h.resendDownloadLinkService.ResendDownloadLinkByTransactionID(uint(transactionID))
	if err != nil {
		slog.Error("Erro ao reenviar link de download", "error", err, "transactionID", transactionID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Link de download reenviado com sucesso",
		"transactionID", transactionID,
		"creatorID", creator.ID,
		"creatorEmail", creator.Email)

	// Redirecionar de volta com sucesso
	http.Redirect(w, r, fmt.Sprintf("/transactions?success=download_link_resent&transaction_id=%d", transactionID), http.StatusSeeOther)
}
