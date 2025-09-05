package handler

import (
	"net/http"
	"strconv"

	"log/slog"

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

	// Obter parâmetros de paginação
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit := 10

	// Buscar transações
	transactions, total, err := h.transactionService.GetTransactionsByCreatorID(creator.ID, page, limit)
	if err != nil {
		slog.Error("Erro ao buscar transações", "error", err)
		http.Error(w, "Erro ao buscar transações", http.StatusInternalServerError)
		return
	}

	// Calcular paginação
	totalPages := (int(total) + limit - 1) / limit
	pagination := map[string]interface{}{
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"TotalItems":  total,
		"HasPrev":     page > 1,
		"HasNext":     page < totalPages,
		"PrevPage":    page - 1,
		"NextPage":    page + 1,
	}

	// Renderizar template
	h.templateRenderer.View(w, r, "transactions/list", map[string]interface{}{
		"Creator":      creator,
		"Transactions": transactions,
		"Pagination":   pagination,
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
