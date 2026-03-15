package handler

import (
	"log"
	"net/http"

	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	librarysvc "github.com/anglesson/simple-web-server/internal/library/service"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// SalesPageHandler gerencia as páginas de vendas
type SalesPageHandler struct {
	ebookService     librarysvc.EbookService
	creatorService   accountsvc.CreatorService
	templateRenderer template.TemplateRenderer
}

// NewSalesPageHandler cria uma instância do SalesPageHandler
func NewSalesPageHandler(ebookService librarysvc.EbookService, creatorService accountsvc.CreatorService, templateRenderer template.TemplateRenderer) *SalesPageHandler {
	return &SalesPageHandler{
		ebookService:     ebookService,
		creatorService:   creatorService,
		templateRenderer: templateRenderer,
	}
}

// SalesPageView exibe a página de vendas pública do ebook
func (h *SalesPageHandler) SalesPageView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	ebook, err := h.ebookService.FindByPublicID(id)
	if err != nil {
		log.Printf("Erro ao buscar ebook por id %s: %v", id, err)
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	if !ebook.Status {
		http.Error(w, "Ebook não disponível", http.StatusNotFound)
		return
	}

	creator, err := h.creatorService.FindByID(ebook.CreatorID)
	if err != nil {
		log.Printf("Erro ao buscar criador do ebook: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	if ebook.AuthorName == "" {
		ebook.AuthorName = creator.GetDisplayName()
	}

	cookieName := "ebook_view_" + id
	if _, err := r.Cookie(cookieName); err != nil {
		ebook.IncrementViews()
		if err := h.ebookService.Update(ebook); err != nil {
			log.Printf("Erro ao incrementar visualizações: %v", err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    "1",
			Path:     "/sales/" + id,
			MaxAge:   86400,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	data := map[string]any{
		"Ebook":   ebook,
		"Creator": creator,
	}

	h.templateRenderer.View(w, r, "purchase/sales-page", data, "guest")
}

// SalesPagePreviewView exibe a página de vendas em modo "preview" para o criador
func (h *SalesPageHandler) SalesPagePreviewView(w http.ResponseWriter, r *http.Request) {
	loggedUser := authmw.Auth(r)
	if loggedUser == nil || loggedUser.ID == 0 {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		http.Error(w, "ID do ebook não fornecido", http.StatusBadRequest)
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil {
		log.Printf("Erro ao buscar ebook: %v", err)
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil || creator.ID != ebook.CreatorID {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	if ebook.AuthorName == "" {
		ebook.AuthorName = creator.GetDisplayName()
	}

	originalPrice := ebook.Value
	savings := originalPrice - ebook.PromotionalValue

	data := map[string]any{
		"Ebook":         ebook,
		"OriginalPrice": utils.FloatToBRL(originalPrice),
		"Savings":       utils.FloatToBRL(savings),
		"Creator":       creator,
		"IsPreview":     true,
	}

	h.templateRenderer.View(w, r, "purchase/sales-page", data, "guest")
}
