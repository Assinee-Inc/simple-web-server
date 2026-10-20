package handler

import (
	"net/http"

	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type ErrorHandler struct{}

func NewErrorHandler(_ ...any) *ErrorHandler {
	return &ErrorHandler{}
}

func (h *ErrorHandler) ErrorView(w http.ResponseWriter, _ *http.Request, code int) {
	switch code {
	case http.StatusNotFound:
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Recurso não encontrado")
	case http.StatusMethodNotAllowed:
		web.RespondError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método HTTP não permitido")
	default:
		if code < 400 {
			code = http.StatusInternalServerError
		}
		web.RespondError(w, code, "INTERNAL_SERVER_ERROR", http.StatusText(code))
	}
}

func (h *ErrorHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.ErrorView(w, r, http.StatusNotFound)
}

func (h *ErrorHandler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	h.ErrorView(w, r, http.StatusMethodNotAllowed)
}
