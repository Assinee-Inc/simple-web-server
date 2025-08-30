package handler

import (
	"fmt"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type CreatorHandler struct {
	creatorService   service.CreatorService
	sessionService   service.SessionService
	templateRenderer template.TemplateRenderer
}

func NewCreatorHandler(creatorService service.CreatorService, sessionService service.SessionService, templateRenderer template.TemplateRenderer) *CreatorHandler {
	return &CreatorHandler{
		creatorService:   creatorService,
		sessionService:   sessionService,
		templateRenderer: templateRenderer,
	}
}

func (ch *CreatorHandler) RegisterView(w http.ResponseWriter, r *http.Request) {
	ch.templateRenderer.View(w, r, "creator/register", nil, "guest")
}

func (ch *CreatorHandler) RegisterCreatorSSR(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	input := models.InputCreateCreator{
		Name:                 r.FormValue("name"),
		BirthDate:            r.FormValue("birthdate"),
		PhoneNumber:          r.FormValue("phone"),
		Email:                r.FormValue("email"),
		CPF:                  r.FormValue("cpf"),
		Password:             r.FormValue("password"),
		PasswordConfirmation: r.FormValue("password_confirmation"),
		TermsAccepted:        r.FormValue("terms_accepted"),
	}

	creator, err := ch.creatorService.CreateCreator(input)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err.Error())
		ch.templateRenderer.View(w, r, "creator/register", map[string]interface{}{
			"Error": err.Error(),
			"Form":  input,
		}, "guest")
		return
	}

	ch.sessionService.InitSession(w, creator.Email)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
