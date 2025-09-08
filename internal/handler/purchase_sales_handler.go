package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type PurchaseSalesHandler struct {
	templateRenderer          template.TemplateRenderer
	purchaseService           service.PurchaseService
	sessionService            service.SessionService
	creatorService            service.CreatorService
	ebookService              service.EbookService
	resendDownloadLinkService service.ResendDownloadLinkServiceInterface
	transactionService        service.TransactionService
}

func NewPurchaseSalesHandler(
	templateRenderer template.TemplateRenderer,
	purchaseService service.PurchaseService,
	sessionService service.SessionService,
	creatorService service.CreatorService,
	ebookService service.EbookService,
	resendDownloadLinkService service.ResendDownloadLinkServiceInterface,
	transactionService service.TransactionService,
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

	// Obter parâmetros de filtros
	ebookIDStr := r.URL.Query().Get("ebook_id")
	clientName := r.URL.Query().Get("client_name")
	clientEmail := r.URL.Query().Get("client_email")

	var ebookID uint
	if ebookIDStr != "" {
		if id, err := strconv.ParseUint(ebookIDStr, 10, 32); err == nil {
			ebookID = uint(id)
		}
	}

	// Buscar vendas (purchases) com filtros
	// Call service to get purchases with filters
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

	// Buscar transações relacionadas às purchases para incluir informações financeiras
	purchaseTransactionMap := make(map[uint]*models.Transaction)
	for _, purchase := range purchases {
		transaction, err := h.transactionService.FindTransactionByPurchaseID(purchase.ID)
		if err == nil && transaction != nil {
			purchaseTransactionMap[purchase.ID] = transaction
		}
	}

	// Buscar ebooks do criador para o filtro
	ebooks, err := h.ebookService.GetEbooksByCreatorID(creator.ID)
	if err != nil {
		slog.Error("Erro ao buscar ebooks para filtro", "error", err)
		// Não falhar, apenas não mostrar filtro
		ebooks = []*models.Ebook{}
	}

	// Calcular paginação usando o modelo padrão do projeto
	pagination := models.NewPagination(page, limit)
	pagination.SetTotal(total)

	// Renderizar template
	h.templateRenderer.View(w, r, "purchase/list", map[string]interface{}{
		"Creator":                creator,
		"Purchases":              purchases,
		"PurchaseTransactionMap": purchaseTransactionMap,
		"Pagination":             pagination,
		"Ebooks":                 ebooks,
		"EbookID":                ebookID,
		"ClientName":             clientName,
		"ClientEmail":            clientEmail,
		"RecordType":             "vendas",
		"Filters": map[string]interface{}{
			"client_name":  clientName,
			"client_email": clientEmail,
			"ebook_id":     ebookID,
		},
	}, "admin")
}

// BlockDownload bloqueia o download de um cliente específico
func (h *PurchaseSalesHandler) BlockDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obter ID da purchase
	purchaseID, err := strconv.ParseUint(r.FormValue("purchase_id"), 10, 32)
	if err != nil {
		slog.Error("ID de purchase inválido", "error", err)
		http.Error(w, "ID de purchase inválido", http.StatusBadRequest)
		return
	}

	// Verificar permissões do usuário
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

	// Buscar purchase para verificar se pertence ao criador
	purchase, err := h.purchaseService.GetPurchaseByID(uint(purchaseID))
	if err != nil {
		slog.Error("Erro ao buscar purchase", "error", err)
		http.Error(w, "Venda não encontrada", http.StatusNotFound)
		return
	}

	// Verificar se a purchase pertence ao criador
	if purchase.Ebook.CreatorID != creator.ID {
		slog.Warn("Tentativa de bloqueio não autorizado",
			"purchaseID", purchaseID,
			"creatorID", creator.ID,
			"ownerID", purchase.Ebook.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	// Bloquear download (definir limit igual ao usado)
	err = h.purchaseService.BlockDownload(uint(purchaseID), creator.ID, true)
	if err != nil {
		slog.Error("Erro ao bloquear download", "error", err, "purchaseID", purchaseID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Download bloqueado com sucesso",
		"purchaseID", purchaseID,
		"creatorID", creator.ID,
		"clientID", purchase.ClientID)

	// Redirecionar de volta com sucesso
	http.Redirect(w, r, "/purchase/sales?success=download_blocked", http.StatusSeeOther)
}

// UnblockDownload desbloqueia o download de um cliente específico
func (h *PurchaseSalesHandler) UnblockDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obter ID da purchase
	purchaseID, err := strconv.ParseUint(r.FormValue("purchase_id"), 10, 32)
	if err != nil {
		slog.Error("ID de purchase inválido", "error", err)
		http.Error(w, "ID de purchase inválido", http.StatusBadRequest)
		return
	}

	// Verificar permissões do usuário
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

	// Buscar purchase para verificar se pertence ao criador
	purchase, err := h.purchaseService.GetPurchaseByID(uint(purchaseID))
	if err != nil {
		slog.Error("Erro ao buscar purchase", "error", err)
		http.Error(w, "Venda não encontrada", http.StatusNotFound)
		return
	}

	// Verificar se a purchase pertence ao criador
	if purchase.Ebook.CreatorID != creator.ID {
		slog.Warn("Tentativa de desbloqueio não autorizado",
			"purchaseID", purchaseID,
			"creatorID", creator.ID,
			"ownerID", purchase.Ebook.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	// Desbloquear download (usar false para desbloquear)
	err = h.purchaseService.BlockDownload(uint(purchaseID), creator.ID, false)
	if err != nil {
		slog.Error("Erro ao desbloquear download", "error", err, "purchaseID", purchaseID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Download desbloqueado com sucesso",
		"purchaseID", purchaseID,
		"creatorID", creator.ID,
		"clientID", purchase.ClientID)

	// Redirecionar de volta com sucesso
	http.Redirect(w, r, "/purchase/sales?success=download_unblocked", http.StatusSeeOther)
}

// ResendDownloadLink reenvia o link de download com opção de novo email
func (h *PurchaseSalesHandler) ResendDownloadLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obter ID da purchase
	purchaseID, err := strconv.ParseUint(r.FormValue("purchase_id"), 10, 32)
	if err != nil {
		slog.Error("ID de purchase inválido", "error", err)
		http.Error(w, "ID de purchase inválido", http.StatusBadRequest)
		return
	}

	// Obter novo email (opcional)
	newEmail := r.FormValue("new_email")

	// Verificar permissões do usuário
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

	// Buscar purchase para verificar se pertence ao criador
	purchase, err := h.purchaseService.GetPurchaseByID(uint(purchaseID))
	if err != nil {
		slog.Error("Erro ao buscar purchase", "error", err)
		http.Error(w, "Venda não encontrada", http.StatusNotFound)
		return
	}

	// Verificar se a purchase pertence ao criador
	if purchase.Ebook.CreatorID != creator.ID {
		slog.Warn("Tentativa de reenvio não autorizado",
			"purchaseID", purchaseID,
			"creatorID", creator.ID,
			"ownerID", purchase.Ebook.CreatorID)
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	// Reenviar o link
	err = h.resendDownloadLinkService.ResendDownloadLinkByPurchaseID(uint(purchaseID), newEmail)
	if err != nil {
		slog.Error("Erro ao reenviar link de download", "error", err, "purchaseID", purchaseID)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	slog.Info("Link de download reenviado com sucesso",
		"purchaseID", purchaseID,
		"creatorID", creator.ID,
		"newEmail", newEmail)

	// Redirecionar de volta com sucesso
	http.Redirect(w, r, fmt.Sprintf("/purchase/sales?success=download_link_resent&purchase_id=%d", purchaseID), http.StatusSeeOther)
}
