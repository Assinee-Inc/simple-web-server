package handler

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	deliverysvc "github.com/Assinee-Inc/docffy/internal/delivery/service"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
	"github.com/go-chi/chi/v5"
)

type DownloadHandler struct {
	downloadService deliverysvc.DownloadService
}

func NewDownloadHandler(downloadService deliverysvc.DownloadService, _ ...any) *DownloadHandler {
	return &DownloadHandler{
		downloadService: downloadService,
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
	if err != nil || purchase == nil {
		web.RespondError(w, http.StatusNotFound, "PURCHASE_NOT_FOUND", "Compra não encontrada")
		return
	}

	if !purchase.IsPaymentConfirmed() {
		h.showPaymentPendingPage(w, r, purchase)
		return
	}

	if purchase.IsExpired() {
		h.showExpiredDownloadPage(w, r, purchase)
		return
	}

	if !purchase.AvailableDownloads() {
		h.showLimitExceededPage(w, r, purchase)
		return
	}

	outputPath, err := h.downloadService.GetEbookFile(hashID, fileIDStr)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "DOWNLOAD_ERROR", err.Error())
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
	purchase, err := h.downloadService.FindPurchaseByHash(hashID)
	if err != nil || purchase == nil {
		web.RespondError(w, http.StatusNotFound, "PURCHASE_NOT_FOUND", "Compra não encontrada")
		return
	}

	if !purchase.IsPaymentConfirmed() {
		h.showPaymentPendingPage(w, r, purchase)
		return
	}

	if purchase.IsExpired() {
		h.showExpiredDownloadPage(w, r, purchase)
		return
	}

	if !purchase.AvailableDownloads() {
		h.showLimitExceededPage(w, r, purchase)
		return
	}

	files, err := h.downloadService.GetEbookFiles(int(purchase.ID))
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "FILES_ERROR", err.Error())
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":        true,
		"ebook_title":    purchase.Ebook.Title,
		"expires_at":     purchase.ExpiresAt,
		"downloads_used": purchase.DownloadsUsed,
		"download_limit": purchase.DownloadLimit,
		"purchase":       purchase,
		"files":          files,
	})
}

func (h *DownloadHandler) showPaymentPendingPage(w http.ResponseWriter, _ *http.Request, _ *salesmodel.Purchase) {
	web.RespondError(w, http.StatusForbidden, "PAYMENT_PENDING", "Pagamento em processamento. O download será liberado assim que o pagamento for confirmado.")
}

func (h *DownloadHandler) showLimitExceededPage(w http.ResponseWriter, _ *http.Request, _ *salesmodel.Purchase) {
	web.RespondError(w, http.StatusForbidden, "DOWNLOAD_LIMIT_EXCEEDED", "Limite de downloads atingido para esta compra.")
}

func (h *DownloadHandler) showExpiredDownloadPage(w http.ResponseWriter, _ *http.Request, purchase *salesmodel.Purchase) {
	daysExpired := int(time.Since(purchase.ExpiresAt).Hours() / 24)
	web.RespondJSON(w, http.StatusGone, map[string]any{
		"success":      false,
		"code":         "DOWNLOAD_EXPIRED",
		"error":        "O prazo para download deste e-book expirou.",
		"days_expired": daysExpired,
	})
}
