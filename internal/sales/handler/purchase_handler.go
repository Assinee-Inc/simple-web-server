package handler

import (
	"log"
	"net/http"
	"strconv"

	librarysvc "github.com/anglesson/simple-web-server/internal/library/service"
	handlerweb "github.com/anglesson/simple-web-server/internal/shared/web"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesrepogorm "github.com/anglesson/simple-web-server/internal/sales/repository/gorm"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/internal/config"
	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
	"github.com/anglesson/simple-web-server/pkg/mail"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

type PurchaseHandler struct {
	templateRenderer template.TemplateRenderer
	ebookService     librarysvc.EbookService
}

func NewPurchaseHandler(templateRenderer template.TemplateRenderer, ebookService librarysvc.EbookService) *PurchaseHandler {
	return &PurchaseHandler{
		templateRenderer: templateRenderer,
		ebookService:     ebookService,
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
		http.Error(w, "Erro to parse form", http.StatusBadRequest)
		return
	}

	ebookPublicID := chi.URLParam(r, "id")

	ebook, err := h.ebookService.FindByPublicID(ebookPublicID)
	if err != nil || ebook == nil {
		log.Printf("Invalid ebook PublicID: %v", ebookPublicID)
		handlerweb.RedirectBackWithErrors(w, r, "Invalid EbookID")
		return
	}

	var clients []uint
	clientRepo := salesrepogorm.NewClientGormRepository()
	for _, clientPublicID := range r.Form["clients[]"] {
		client, err := clientRepo.FindByPublicID(clientPublicID)
		if err != nil {
			log.Printf("Invalid client PublicID: %v", clientPublicID)
			continue
		}
		clients = append(clients, client.ID)
		log.Printf("ClientID: %v (PublicID: %v)", client.ID, clientPublicID)
	}

	if len(clients) == 0 {
		cookies.NotifyError(w, "Informe os clientes que receberão o e-book")
		return
	}

	err = purchaseServiceFactory().CreatePurchase(ebook.ID, clients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	cookies.NotifySuccess(w, "Envio realizado!")
	http.Redirect(w, r, "/ebook/view/"+ebookPublicID, http.StatusSeeOther)
}
