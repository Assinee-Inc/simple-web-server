package handler

import (
	"log"
	"net/http"
	"strconv"

	handlerweb "github.com/anglesson/simple-web-server/internal/shared/web"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/internal/config"
	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
	"github.com/anglesson/simple-web-server/pkg/mail"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

type PurchaseHandler struct {
	templateRenderer template.TemplateRenderer
}

func NewPurchaseHandler(templateRenderer template.TemplateRenderer) *PurchaseHandler {
	return &PurchaseHandler{
		templateRenderer: templateRenderer,
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

	var clients []uint
	ebookIdStr := chi.URLParam(r, "id")

	ebookId, err := strconv.Atoi(ebookIdStr)
	if err != nil {
		log.Printf("Invalid client ID: %v", ebookIdStr)
		handlerweb.RedirectBackWithErrors(w, r, "Invalid EbookID")
	}

	for _, idStr := range r.Form["clients[]"] {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Invalid client ID: %v", idStr)
			continue
		}
		clients = append(clients, uint(id))
		log.Printf("ClientID: %v", id)
	}

	if len(clients) == 0 {
		cookies.NotifyError(w, "Informe os clientes que receberão o e-book")
		return
	}

	err = purchaseServiceFactory().CreatePurchase(uint(ebookId), clients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	cookies.NotifySuccess(w, "Envio realizado!")
	http.Redirect(w, r, "/ebook/view/"+ebookIdStr, http.StatusSeeOther)
}
