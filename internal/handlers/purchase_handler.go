package handlers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/mail"
	"github.com/anglesson/simple-web-server/internal/repositories"
	"github.com/anglesson/simple-web-server/internal/services"
	cookies "github.com/anglesson/simple-web-server/internal/shared/cookie"
	"github.com/go-chi/chi/v5"
)

func purchaseServiceFactory() *services.PurchaseService {
	mailPort, _ := strconv.Atoi(config.AppConfig.MailPort)
	ms := mail.NewEmailService(mail.NewGoMailer(
		config.AppConfig.MailHost,
		mailPort,
		config.AppConfig.MailUsername,
		config.AppConfig.MailPassword))
	pr := repositories.NewPurchaseRepository()
	return services.NewPurchaseService(pr, ms)
}

func PurchaseCreateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro to parse form", http.StatusBadRequest)
		return
	}

	var clients []uint
	ebookIdStr := chi.URLParam(r, "id")

	ebookId, err := strconv.Atoi(ebookIdStr)
	if err != nil {
		log.Printf("Invalid client ID: %v", ebookIdStr)
		redirectBackWithErrors(w, r, "Invalid EbookID")
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

	purchaseServiceFactory().CreatePurchase(uint(ebookId), clients)

	cookies.NotifySuccess(w, "Envio realizado!")
	http.Redirect(w, r, "/ebook/view/"+ebookIdStr, http.StatusSeeOther)
}

func PurchaseDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID Purchase
	idStrPurchase := chi.URLParam(r, "id")

	purchaseID, _ := strconv.Atoi(idStrPurchase)

	outputPath, err := purchaseServiceFactory().GetEbookFile(purchaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer os.Remove(outputPath)

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(outputPath))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, outputPath)
}
