package handler

import (
	"net/http"
	"strconv"

	"github.com/anglesson/simple-web-server/pkg/template"
)

type ErrorHandler struct {
	templateRenderer template.TemplateRenderer
}

func NewErrorHandler(templateRenderer template.TemplateRenderer) *ErrorHandler {
	return &ErrorHandler{
		templateRenderer: templateRenderer,
	}
}

// TODO: Render errors dynamically
func (h *ErrorHandler) ErrorView(w http.ResponseWriter, r *http.Request, code int) {
	codeStr := strconv.Itoa(code)
	h.templateRenderer.View(w, r, codeStr+"-error", nil, "guest")
}
