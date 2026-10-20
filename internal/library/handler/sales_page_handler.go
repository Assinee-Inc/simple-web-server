package handler

import (
	"log/slog"
	"net/http"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	librarysvc "github.com/Assinee-Inc/docffy/internal/library/service"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/go-chi/chi/v5"
)

type SalesPageCreatorLookup interface {
	FindByID(id uint) (*accountmodel.Creator, error)
	FindCreatorByUserID(userID uint) (*accountmodel.Creator, error)
}

type SalesPageHandler struct {
	ebookService   librarysvc.EbookService
	creatorService SalesPageCreatorLookup
}

func NewSalesPageHandler(
	ebookService librarysvc.EbookService,
	creatorService SalesPageCreatorLookup,
	_ ...any,
) *SalesPageHandler {
	return &SalesPageHandler{
		ebookService:   ebookService,
		creatorService: creatorService,
	}
}

func (h *SalesPageHandler) SalesPageView(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Página não encontrada")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil || !ebook.Status {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "E-book não encontrado ou indisponível")
		return
	}

	creator, err := h.creatorService.FindByID(ebook.CreatorID)
	if err != nil || creator == nil {
		web.RespondError(w, http.StatusInternalServerError, "CREATOR_NOT_FOUND", "Erro ao carregar informações do autor")
		return
	}

	if ebook.AuthorName == "" {
		ebook.AuthorName = creator.GetDisplayName()
	}

	viewCookieName := "ebook_view_" + publicID
	if _, err := r.Cookie(viewCookieName); err != nil {
		ebook.IncrementViews()
		if err := h.ebookService.Update(ebook); err != nil {
			slog.Warn("Falha ao atualizar contagem de visualizações", "ebookID", ebook.ID, "error", err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     viewCookieName,
			Value:    "1",
			Path:     "/",
			MaxAge:   86400,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":          true,
		"ebook":            ebook,
		"creator":          creator,
		"original_price":   ebook.Value,
		"savings":          ebook.Value * 0.25,
		"formatted_price":  ebook.GetValue(),
		"checkout_url":     "/checkout/" + ebook.PublicID,
		"has_stripe_setup": creator.StripeConnectAccountID != "" && creator.OnboardingCompleted,
	})
}

func (h *SalesPageHandler) SalesPagePreviewView(w http.ResponseWriter, r *http.Request) {
	currentUser := h.getCurrentUser(r)
	if currentUser == nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Acesso não autorizado")
		return
	}

	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID do e-book inválido")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "E-book não encontrado")
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(currentUser.ID)
	if err != nil || creator == nil || ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Acesso não autorizado")
		return
	}

	if ebook.AuthorName == "" {
		ebook.AuthorName = creator.GetDisplayName()
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":          true,
		"is_preview":       true,
		"ebook":            ebook,
		"creator":          creator,
		"original_price":   ebook.Value * 1.33,
		"savings":          ebook.Value * 0.33,
		"formatted_price":  ebook.GetValue(),
		"checkout_url":     "#",
		"has_stripe_setup": creator.StripeConnectAccountID != "" && creator.OnboardingCompleted,
	})
}

func (h *SalesPageHandler) getCurrentUser(r *http.Request) *authmodel.User {
	if user := authmw.GetAuthenticatedUser(r.Context()); user != nil {
		return user
	}
	userID, _ := authmw.GetAuthenticatedUserID(r.Context())
	email := authmw.GetAuthenticatedUserEmail(r.Context())
	if userID > 0 {
		u := &authmodel.User{Email: email}
		u.ID = userID
		return u
	}
	if email != "" && database.DB != nil {
		userRepository := authrepo.NewGormUserRepository(database.DB)
		return userRepository.FindByEmail(email)
	}
	return nil
}
