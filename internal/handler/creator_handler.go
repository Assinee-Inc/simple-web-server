package handler

import (
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

	ch.sessionService.InitSession(w, creator.Email)

	// Redirecionar para página de boas-vindas do onboarding
	http.Redirect(w, r, "/stripe-connect/welcome", http.StatusSeeOther)
}
