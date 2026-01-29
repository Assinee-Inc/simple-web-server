package ebook

import (
	"encoding/json"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/platform/router"
	"github.com/anglesson/simple-web-server/internal/platform/validator"
)

type ErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type CreateEbookRequest struct {
	Title       string `json:"title" validate:"required,max=120"`
	Description string `json:"description" validate:"max=120"`
	// preços em centavos
	Price            int64    `json:"price" validate:"required,gt=0"`
	PromotionalPrice int64    `json:"promotional_price" validate:"omitempty,gte=0,ltfield=Price"`
	CoverImage       string   `json:"cover_image" validate:"omitempty,url"`
	SalesDescription string   `json:"sales_description" validate:"max=255"`
	InfoProducerID   string   `json:"info_producer_id" validate:"required"`
	FileIDs          []string `json:"file_ids" validate:"dive,uuid"`
}

func (r *CreateEbookRequest) ToDomain() *Ebook {
	return &Ebook{
		Title:            r.Title,
		Description:      r.Description,
		SalesDescription: r.SalesDescription,
		Price:            r.Price,
		PromotionalPrice: r.PromotionalPrice,
		CoverImage:       r.CoverImage,
		InfoProducerID:   r.InfoProducerID,
		FileIDs:          r.FileIDs,
	}
}

type APIHandler struct {
	manager ManagerPort
}

func NewHandler(manager ManagerPort) *APIHandler {
	return &APIHandler{
		manager: manager,
	}
}

func (h *APIHandler) CreateEbook(w http.ResponseWriter, r *http.Request) {
	var input CreateEbookRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "verifique o corpo da requisição", nil)
		return
	}

	v := validator.NewValidator()
	if err := v.Struct(input); err != nil {
		msgs := validator.Translate(err, input)
		h.respondError(w, http.StatusUnprocessableEntity, "erro na validação", msgs)
		return
	}

	ebook := input.ToDomain()

	createdEbook, err := h.manager.CreateEbook(r.Context(), ebook)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.respondSuccess(w, http.StatusCreated, createdEbook)
}

func (h *APIHandler) respondError(w http.ResponseWriter, code int, msg string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := ErrorResponse{
		Message: msg,
		Fields:  fields,
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) respondSuccess(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(data)
}

func (h *APIHandler) RegisterRoutes(r router.Router) {
	r.Handle("POST", "/ebooks", h.CreateEbook)
}
