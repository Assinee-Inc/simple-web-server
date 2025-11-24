package api

import (
	"encoding/json"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/library/application/services"
	"github.com/anglesson/simple-web-server/internal/library/domain"
)

type EbookStore interface {
	Create(ebook *domain.Ebook)
}

type EbookHandler struct {
	ebookSvc services.EbookServicePort
}

func NewEbookHandler(ebookSvc services.EbookServicePort) *EbookHandler {
	return &EbookHandler{
		ebookSvc: ebookSvc,
	}
}

func (h *EbookHandler) CreateEbook(w http.ResponseWriter, r *http.Request) {
	var ebook domain.Ebook
	ebook.Title = "Any Title"

	h.ebookSvc.CreateEbook(ebook)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ebook)
}
