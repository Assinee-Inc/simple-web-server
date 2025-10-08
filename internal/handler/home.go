package handler

import (
	"net/http"

	"github.com/anglesson/simple-web-server/pkg/template"
)

type HomeHandler struct {
	templateRenderer template.TemplateRenderer
	errorHandler     *ErrorHandler
}

func NewHomeHandler(templateRenderer template.TemplateRenderer, errorHandler *ErrorHandler) *HomeHandler {
	return &HomeHandler{
		templateRenderer: templateRenderer,
		errorHandler:     errorHandler,
	}
}

func (h *HomeHandler) HomeView(w http.ResponseWriter, r *http.Request) {
	h.templateRenderer.View(w, r, "home", nil, "guest")
}
