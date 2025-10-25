package handler

import (
	"log"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type SettingsHandler struct {
	sessionService   service.SessionService
	templateRenderer template.TemplateRenderer
}

func NewSettingsHandler(
	sessionService service.SessionService,
	templateRenderer template.TemplateRenderer,
) *SettingsHandler {
	return &SettingsHandler{
		sessionService:   sessionService,
		templateRenderer: templateRenderer,
	}
}

func (h *SettingsHandler) SettingsView(w http.ResponseWriter, r *http.Request) {
	user := middleware.Auth(r)
	if user == nil {
		log.Printf("Usuário não autenticado ao acessar configurações")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Gerar novo token CSRF se necessário
	csrfToken, _ := h.sessionService.RegenerateCSRFToken(r, w)
	user.CSRFToken = csrfToken

	log.Printf("Renderizando página de configurações para o usuário: %s", user.Email)
	log.Printf("Token CSRF: %s", user.CSRFToken)

	// Passar apenas o objeto user para o template
	h.templateRenderer.View(w, r, "settings", map[string]interface{}{
		"user": user,
	}, "admin")
}
