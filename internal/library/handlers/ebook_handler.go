package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/library/models"
	"github.com/anglesson/simple-web-server/internal/library/ports"
)

type EbookHandler struct {
	ebookSvc ports.EbookService
}

func NewEbookHandler(ebookSvc ports.EbookService) *EbookHandler {
	return &EbookHandler{
		ebookSvc: ebookSvc,
	}
}

func (h *EbookHandler) CreateEbook(w http.ResponseWriter, r *http.Request) {
	var input models.Ebook
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdEbook, err := h.ebookSvc.CreateEbook(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdEbook)
}
