package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type SettingsHandler struct {
	sessionService authsvc.SessionService
	creatorService accountsvc.CreatorService
}

func NewSettingsHandler(
	deps ...any,
) *SettingsHandler {
	h := &SettingsHandler{}
	for _, dep := range deps {
		switch d := dep.(type) {
		case authsvc.SessionService:
			h.sessionService = d
		case accountsvc.CreatorService:
			h.creatorService = d
		}
	}
	return h
}

func (h *SettingsHandler) SettingsView(w http.ResponseWriter, r *http.Request) {
	user := authmw.Auth(r)
	if user == nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil {
		log.Printf("Erro ao buscar creator nas configurações: %v", err)
	}

	resp := map[string]any{
		"success": true,
		"user":    user,
		"creator": creator,
	}
	if creator != nil {
		resp["social_name"] = creator.SocialName
		resp["facebook_pixel_id"] = creator.FacebookPixelID
	}

	web.RespondJSON(w, http.StatusOK, resp)
}

func (h *SettingsHandler) UpdateFacebookPixel(w http.ResponseWriter, r *http.Request) {
	user := authmw.Auth(r)
	if user == nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
		return
	}

	pixelID, err := extractFieldFromJSONOrForm(r, "facebook_pixel_id")
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Erro ao processar requisição", nil)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil || creator == nil {
		web.RespondError(w, http.StatusNotFound, "CREATOR_NOT_FOUND", "Creator não encontrado", nil)
		return
	}

	if err := accountmodel.ValidateFacebookPixelID(pixelID); err != nil {
		web.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	creator.FacebookPixelID = pixelID
	if err := h.creatorService.UpdateCreator(creator); err != nil {
		log.Printf("Erro ao atualizar Facebook Pixel ID: %v", err)
		web.RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Erro ao salvar Facebook Pixel ID", nil)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":           true,
		"facebook_pixel_id": creator.FacebookPixelID,
	})
}

func (h *SettingsHandler) UpdateSocialName(w http.ResponseWriter, r *http.Request) {
	user := authmw.Auth(r)
	if user == nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
		return
	}

	socialName, err := extractFieldFromJSONOrForm(r, "social_name")
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Erro ao processar requisição", nil)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil || creator == nil {
		web.RespondError(w, http.StatusNotFound, "CREATOR_NOT_FOUND", "Creator não encontrado", nil)
		return
	}

	creator.SocialName = socialName
	if err := h.creatorService.UpdateCreator(creator); err != nil {
		log.Printf("Erro ao atualizar nome social: %v", err)
		web.RespondError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Erro ao salvar nome social", nil)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"social_name": creator.SocialName,
	})
}

func extractFieldFromJSONOrForm(r *http.Request, field string) (string, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			return "", err
		}
		if val, ok := payload[field].(string); ok {
			return val, nil
		}
		return "", nil
	}

	if err := r.ParseForm(); err != nil {
		return "", err
	}
	return r.FormValue(field), nil
}
