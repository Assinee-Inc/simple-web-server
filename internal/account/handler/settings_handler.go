package handler

import (
	"log"
	"net/http"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type SettingsHandler struct {
	sessionService   authsvc.SessionService
	creatorService   accountsvc.CreatorService
	templateRenderer template.TemplateRenderer
}

func NewSettingsHandler(
	sessionService authsvc.SessionService,
	creatorService accountsvc.CreatorService,
	templateRenderer template.TemplateRenderer,
) *SettingsHandler {
	return &SettingsHandler{
		sessionService:   sessionService,
		creatorService:   creatorService,
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

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil {
		log.Printf("Erro ao buscar creator nas configurações: %v", err)
	}

	successMessages := h.sessionService.GetFlashes(w, r, "success")
	errorMessages := h.sessionService.GetFlashes(w, r, "error")

	h.templateRenderer.View(w, r, "settings", map[string]interface{}{
		"user":    user,
		"Creator": creator,
		"Success": successMessages,
		"Errors":  errorMessages,
	}, "admin-daisy")
}

func (h *SettingsHandler) UpdateFacebookPixel(w http.ResponseWriter, r *http.Request) {
	user := authmw.Auth(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil {
		h.sessionService.AddFlash(w, r, "Creator não encontrado", "error")
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	pixelID := r.FormValue("facebook_pixel_id")
	if err := accountmodel.ValidateFacebookPixelID(pixelID); err != nil {
		h.sessionService.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	creator.FacebookPixelID = pixelID
	if err := h.creatorService.UpdateCreator(creator); err != nil {
		log.Printf("Erro ao atualizar Facebook Pixel ID: %v", err)
		h.sessionService.AddFlash(w, r, "Erro ao salvar Facebook Pixel ID", "error")
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	h.sessionService.AddFlash(w, r, "Facebook Pixel ID atualizado com sucesso!", "success")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (h *SettingsHandler) UpdateSocialName(w http.ResponseWriter, r *http.Request) {
	user := authmw.Auth(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil {
		h.sessionService.AddFlash(w, r, "Creator não encontrado", "error")
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	creator.SocialName = r.FormValue("social_name")
	if err := h.creatorService.UpdateCreator(creator); err != nil {
		log.Printf("Erro ao atualizar nome social: %v", err)
		h.sessionService.AddFlash(w, r, "Erro ao salvar nome social", "error")
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	h.sessionService.AddFlash(w, r, "Nome social atualizado com sucesso!", "success")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
