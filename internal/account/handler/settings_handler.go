package handler

import (
	"log"
	"net/http"

	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type SettingsHandler struct {
	sessionService   authsvc.SessionService
	templateRenderer template.TemplateRenderer
}

func NewSettingsHandler(
	sessionService authsvc.SessionService,
	templateRenderer template.TemplateRenderer,
) *SettingsHandler {
	return &SettingsHandler{
		sessionService:   sessionService,
		templateRenderer: templateRenderer,
	}
}

func (h *SettingsHandler) SettingsView(w http.ResponseWriter, r *http.Request) {
	user := authmw.Auth(r)
	if user == nil {
		log.Printf("Usuário não autenticado ao acessar configurações")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	csrfToken, _ := h.sessionService.RegenerateCSRFToken(r, w)
	user.CSRFToken = csrfToken

	log.Printf("Renderizando página de configurações para o usuário: %s", user.Email)
	log.Printf("Token CSRF: %s", user.CSRFToken)

	h.templateRenderer.View(w, r, "settings", map[string]interface{}{
		"user": user,
	}, "admin-daisy")
}
