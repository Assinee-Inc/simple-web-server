package handler

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/handler/web"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/service"
	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
	"github.com/anglesson/simple-web-server/pkg/mail"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

func purchaseServiceFactory() *service.PurchaseService {
	mailPort, _ := strconv.Atoi(config.AppConfig.MailPort)
	ms := mail.NewEmailService(mail.NewGoMailer(
		config.AppConfig.MailHost,
		mailPort,
		config.AppConfig.MailUsername,
		config.AppConfig.MailPassword))
	pr := repository.NewPurchaseRepository()
	return service.NewPurchaseService(pr, ms)
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
		web.RedirectBackWithErrors(w, r, "Invalid EbookID")
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

func PurchaseDownloadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 PurchaseDownloadHandler chamado: %s", r.URL.Path)

	// Get ID Purchase and File ID
	idStrPurchase := chi.URLParam(r, "id")
	fileIDStr := r.URL.Query().Get("file_id")

	log.Printf("📋 Purchase ID: %s, File ID: %s", idStrPurchase, fileIDStr)

	purchaseID, err := strconv.Atoi(idStrPurchase)
	if err != nil {
		log.Printf("❌ Erro ao converter purchase ID: %v", err)
		http.Error(w, "ID da compra inválido", http.StatusBadRequest)
		return
	}

	// Se não especificou arquivo, mostrar lista de arquivos disponíveis
	if fileIDStr == "" {
		log.Printf("📄 Mostrando lista de arquivos para purchase ID: %d", purchaseID)
		showEbookFiles(w, r, purchaseID)
		return
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID do arquivo inválido", http.StatusBadRequest)
		return
	}

	outputPath, err := purchaseServiceFactory().GetEbookFile(purchaseID, uint(fileID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer os.Remove(outputPath)

	// Extrair nome do arquivo do path
	fileName := outputPath
	if idx := strings.LastIndex(outputPath, "/"); idx != -1 {
		fileName = outputPath[idx+1:]
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, outputPath)
}

func showEbookFiles(w http.ResponseWriter, r *http.Request, purchaseID int) {
	log.Printf("🔍 showEbookFiles chamado para purchase ID: %d", purchaseID)

	files, err := purchaseServiceFactory().GetEbookFiles(purchaseID)
	if err != nil {
		log.Printf("❌ Erro ao buscar arquivos: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Arquivos encontrados: %d", len(files))

	// Buscar informações da compra para o template
	purchase, err := repository.NewPurchaseRepository().FindByID(uint(purchaseID))
	if err != nil {
		log.Printf("❌ Erro ao buscar purchase: %v", err)
		http.Error(w, "Compra não encontrada", http.StatusNotFound)
		return
	}

	log.Printf("✅ Purchase carregada: %s", purchase.Ebook.Title)

	data := map[string]interface{}{
		"Data": map[string]interface{}{
			"Purchase": purchase,
			"Files":    files,
		},
		"Title": "Download do Ebook",
	}

	template.ViewWithoutLayout(w, r, "ebook/download", data)
}
