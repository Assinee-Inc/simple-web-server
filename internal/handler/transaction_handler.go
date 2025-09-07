package handler

import (
	"net/http"
	"strconv"

	"log/slog"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type TransactionHandler struct {
	transactionService service.TransactionService
	sessionService     service.SessionService
	creatorService     service.CreatorService
	templateRenderer   template.TemplateRenderer
}

func NewTransactionHandler(
	transactionService service.TransactionService,
	sessionService service.SessionService,
	creatorService service.CreatorService,
	templateRenderer template.TemplateRenderer,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		sessionService:     sessionService,
		creatorService:     creatorService,
		templateRenderer:   templateRenderer,
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
