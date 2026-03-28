package handler

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	deliverysvc "github.com/anglesson/simple-web-server/internal/delivery/service"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

type DownloadHandler struct {
	downloadService  deliverysvc.DownloadService
	templateRenderer template.TemplateRenderer
}

func NewDownloadHandler(downloadService deliverysvc.DownloadService, templateRenderer template.TemplateRenderer) *DownloadHandler {
	return &DownloadHandler{
		downloadService:  downloadService,
		templateRenderer: templateRenderer,
	}
}

func (h *DownloadHandler) PurchaseDownloadHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("PurchaseDownloadHandler chamado", "path", r.URL.Path)

	hashID := chi.URLParam(r, "hash_id")
	fileIDStr := r.URL.Query().Get("file_id")

	if fileIDStr == "" {
		slog.Info("Mostrando lista de arquivos para purchase", "hashID", hashID)
		h.showEbookFiles(w, r, hashID)
		return
	}

	purchase, err := h.downloadService.FindPurchaseByHash(hashID)
	if err != nil {
		http.Error(w, "Compra não encontrada", http.StatusNotFound)
		return
	}

	if !purchase.IsPaymentConfirmed() {
		h.showPaymentPendingPage(w, r, purchase)
		return
	}

	outputPath, err := h.downloadService.GetEbookFile(hashID, fileIDStr)
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

func (h *DownloadHandler) showEbookFiles(w http.ResponseWriter, r *http.Request, hashID string) {
	log.Printf("showEbookFiles chamado para purchase: %s", hashID)

	purchase, err := h.downloadService.FindPurchaseByHash(hashID)
	if err != nil {
		log.Printf("Erro ao buscar purchase: %v", err)
		http.Error(w, "Compra não encontrada", http.StatusNotFound)
		return
	}

	log.Printf("Purchase carregada: %s", purchase.Ebook.Title)

	if !purchase.IsPaymentConfirmed() {
		log.Printf("Pagamento pendente para purchase: %s", hashID)
		h.showPaymentPendingPage(w, r, purchase)
		return
	}

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

	files, err := h.downloadService.GetEbookFiles(int(purchase.ID))
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

func (h *DownloadHandler) showPaymentPendingPage(w http.ResponseWriter, r *http.Request, purchase *salesmodel.Purchase) {
	log.Printf("Mostrando página de pagamento pendente para purchase ID: %d", purchase.ID)

	data := map[string]interface{}{
		"Purchase": purchase,
		"Title":    "Pagamento em Processamento",
	}

	h.templateRenderer.ViewWithoutLayout(w, r, "ebook/payment-pending", data)
}

func (h *DownloadHandler) showLimitExceededPage(w http.ResponseWriter, r *http.Request, purchase *salesmodel.Purchase) {
	log.Printf("Mostrando página de limite excedido para purchase ID: %d", purchase.ID)

	data := map[string]interface{}{
		"Purchase": purchase,
		"Title":    "Limite de Downloads Atingido",
	}

	h.templateRenderer.ViewWithoutLayout(w, r, "ebook/download-limit-exceeded", data)
}

func (h *DownloadHandler) showExpiredDownloadPage(w http.ResponseWriter, r *http.Request, purchase *salesmodel.Purchase) {
	log.Printf("Mostrando página de download expirado para purchase ID: %d", purchase.ID)

	daysExpired := int(time.Since(purchase.ExpiresAt).Hours() / 24)

	data := map[string]interface{}{
		"Purchase":    purchase,
		"DaysExpired": daysExpired,
		"Title":       "Download Expirado",
	}

	h.templateRenderer.ViewWithoutLayout(w, r, "ebook/download-expired", data)
}
