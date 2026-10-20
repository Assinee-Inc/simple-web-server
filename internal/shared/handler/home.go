package handler

import (
	"net/http"

	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type HomeHandler struct {
	errorHandler *ErrorHandler
}

func NewHomeHandler(deps ...any) *HomeHandler {
	h := &HomeHandler{}
	for _, dep := range deps {
		if eh, ok := dep.(*ErrorHandler); ok {
			h.errorHandler = eh
		}
	}
	return h
}

func (h *HomeHandler) HomeView(w http.ResponseWriter, _ *http.Request) {
	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"status":  "ok",
		"service": "docffy-api",
	})
}
