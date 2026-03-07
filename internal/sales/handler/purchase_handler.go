package handler

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	handlerweb "github.com/anglesson/simple-web-server/internal/handler/web"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/internal/service"
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
	ms := service.NewEmailService(mail.NewGoMailer(
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

func (h *PurchaseHandler) PurchaseDownloadHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("PurchaseDownloadHandler chamado", "path", r.URL.Path)

	hashID := chi.URLParam(r, "hash_id")
	fileIDStr := r.URL.Query().Get("file_id")

	if fileIDStr == "" {
		slog.Info("Mostrando lista de arquivos para purchase", "hashID", hashID)
		h.showEbookFiles(w, r, hashID)
		return
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID do arquivo inválido", http.StatusBadRequest)
		return
	}

	outputPath, err := purchaseServiceFactory().GetEbookFile(hashID, uint(fileID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer os.Remove(outputPath)

	fileName := outputPath
	if idx := strings.LastIndex(outputPath, "/"); idx != -1 {
		fileName = outputPath[idx+1:]
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, outputPath)
}

func (h *PurchaseHandler) showEbookFiles(w http.ResponseWriter, r *http.Request, hashID string) {
	log.Printf("showEbookFiles chamado para purchase: %s", hashID)

	purchase, err := salesrepo.NewPurchaseRepository().FindEbookByPurchaseHash(hashID)
	if err != nil {
		log.Printf("Erro ao buscar purchase: %v", err)
		http.Error(w, "Compra não encontrada", http.StatusNotFound)
		return
	}

	log.Printf("Purchase carregada: %s", purchase.Ebook.Title)

	if purchase.IsExpired() {
		log.Printf("Download expirado para purchase: %s", hashID)
		h.showExpiredDownloadPage(w, r, purchase)
		return
	}

	if !purchase.AvailableDownloads() {
		log.Printf("Limite de downloads atingido para purchase: %s", hashID)
		h.showLimitExceededPage(w, r, purchase)
		return
	}

	files, err := purchaseServiceFactory().GetEbookFiles(int(purchase.ID))
	if err != nil {
		log.Printf("Erro ao buscar arquivos: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Arquivos encontrados: %d", len(files))

	data := map[string]interface{}{
		"Purchase": purchase,
		"Files":    files,
		"Title":    "Download do Ebook",
	}

	h.templateRenderer.ViewWithoutLayout(w, r, "ebook/download", data)
}

func (h *PurchaseHandler) showLimitExceededPage(w http.ResponseWriter, r *http.Request, purchase *salesmodel.Purchase) {
	log.Printf("Mostrando página de limite excedido para purchase ID: %d", purchase.ID)

	data := map[string]interface{}{
		"Purchase": purchase,
		"Title":    "Limite de Downloads Atingido",
	}

	h.templateRenderer.ViewWithoutLayout(w, r, "ebook/download-limit-exceeded", data)
}

func (h *PurchaseHandler) showExpiredDownloadPage(w http.ResponseWriter, r *http.Request, purchase *salesmodel.Purchase) {
	log.Printf("Mostrando página de download expirado para purchase ID: %d", purchase.ID)

	daysExpired := int(time.Since(purchase.ExpiresAt).Hours() / 24)

	data := map[string]interface{}{
		"Purchase":    purchase,
		"DaysExpired": daysExpired,
		"Title":       "Download Expirado",
	}

	h.templateRenderer.ViewWithoutLayout(w, r, "ebook/download-expired", data)
}
