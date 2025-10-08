package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type CreatorHandler struct {
	creatorService       service.CreatorService
	stripeConnectService service.StripeConnectService
	sessionService       service.SessionService
	templateRenderer     template.TemplateRenderer
}

func NewCreatorHandler(
	creatorService service.CreatorService,
	stripeConnectService service.StripeConnectService,
	sessionService service.SessionService,
	templateRenderer template.TemplateRenderer,
) *CreatorHandler {
	return &CreatorHandler{
		creatorService:       creatorService,
		stripeConnectService: stripeConnectService,
		sessionService:       sessionService,
		templateRenderer:     templateRenderer,
	}
}

func (ch *CreatorHandler) RegisterView(w http.ResponseWriter, r *http.Request) {
	csrfToken, err := ch.sessionService.RegenerateCSRFToken(r, w)
	if err != nil {
		http.Error(w, "Unable to generate CSRF token", http.StatusInternalServerError)
		return
	}

	var form models.InputCreateCreator
	formBytes := ch.sessionService.Get(r, "form")
	if formBytes != nil {
		if data, ok := formBytes.([]byte); ok {
			json.Unmarshal(data, &form)
		}
	}

	errors := ch.sessionService.GetFlashes(w, r, "error")

	data := map[string]any{
		"csrf_token": csrfToken,
		"Form":       form,
		"Errors":     errors,
	}

	ch.templateRenderer.View(w, r, "creator/register", data, "guest")
}

func (ch *CreatorHandler) RegisterCreatorSSR(w http.ResponseWriter, r *http.Request) {
	_, err := ch.sessionService.RegenerateCSRFToken(r, w)
	if err != nil {
		http.Error(w, "Unable to generate CSRF token", http.StatusInternalServerError)
		return
	}

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
		fmt.Printf("[ERROR]: %v\n", err)
		ch.sessionService.AddFlash(w, r, err.Error(), "error")
		formData, _ := json.Marshal(input)
		ch.sessionService.Set(r, w, "form", formData)
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	// Criar automaticamente conta Stripe Connect com os dados fornecidos
	log.Printf("Criando conta Stripe Connect para creator: %s", creator.Email)
	stripeAccountID, err := ch.stripeConnectService.CreateConnectAccount(creator)
	if err != nil {
		log.Printf("Erro ao criar conta Stripe Connect: %v", err)
		// Não falhar o registro por conta disso, mas logar o erro
	} else {
		// Atualizar creator com ID da conta Stripe
		creator.StripeConnectAccountID = stripeAccountID
		err = ch.creatorService.UpdateCreator(creator)
		if err != nil {
			log.Printf("Erro ao atualizar creator com Stripe Account ID: %v", err)
		} else {
			log.Printf("Conta Stripe Connect criada com sucesso: %s", stripeAccountID)
		}
	}

	ch.sessionService.InitSession(w, r, creator.UserID, creator.Email)

	// Redirecionar para página de boas-vindas do onboarding
	http.Redirect(w, r, "/stripe-connect/welcome", http.StatusSeeOther)
}
