package handler

import (
	"net/http"
	"strconv"

	"github.com/Assinee-Inc/docffy/internal/config"
	librarysvc "github.com/Assinee-Inc/docffy/internal/library/service"
	salesrepo "github.com/Assinee-Inc/docffy/internal/sales/repository"
	salesrepogorm "github.com/Assinee-Inc/docffy/internal/sales/repository/gorm"
	salesvc "github.com/Assinee-Inc/docffy/internal/sales/service"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
	"github.com/Assinee-Inc/docffy/pkg/mail"
	"github.com/go-chi/chi/v5"
)

type PurchaseHandler struct {
	ebookService librarysvc.EbookService
}

func NewPurchaseHandler(deps ...any) *PurchaseHandler {
	var ebookSvc librarysvc.EbookService
	for _, dep := range deps {
		if svc, ok := dep.(librarysvc.EbookService); ok {
			ebookSvc = svc
		}
	}
	return &PurchaseHandler{
		ebookService: ebookSvc,
	}
}

func purchaseServiceFactory() salesvc.PurchaseService {
	mailPort, _ := strconv.Atoi(config.AppConfig.MailPort)
	ms := salesvc.NewEmailService(mail.NewGoMailer(
		config.AppConfig.MailHost,
		mailPort,
		config.AppConfig.MailUsername,
		config.AppConfig.MailPassword))
	pr := salesrepo.NewPurchaseRepository()
	return salesvc.NewPurchaseService(pr, ms)
}

func (h *PurchaseHandler) PurchaseCreateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_FORM", "Erro ao processar dados do formulário")
		return
	}

	ebookPublicID := chi.URLParam(r, "id")

	ebook, err := h.ebookService.FindByPublicID(ebookPublicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "EBOOK_NOT_FOUND", "E-book não encontrado")
		return
	}

	var clients []uint
	clientRepo := salesrepogorm.NewClientGormRepository()
	for _, clientPublicID := range r.Form["clients[]"] {
		client, err := clientRepo.FindByPublicID(clientPublicID)
		if err != nil {
			continue
		}
		clients = append(clients, client.ID)
	}

	if len(clients) == 0 {
		web.RespondError(w, http.StatusBadRequest, "CLIENTS_REQUIRED", "Informe os clientes que receberão o e-book")
		return
	}

	err = purchaseServiceFactory().CreatePurchase(ebook.ID, clients)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "PURCHASE_ERROR", err.Error())
		return
	}

	web.RespondJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Envio realizado!",
	})
}
