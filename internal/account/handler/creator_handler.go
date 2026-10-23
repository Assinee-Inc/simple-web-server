package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type CreatorHandler struct {
	creatorService       accountsvc.CreatorService
	stripeConnectService accountsvc.StripeConnectService
}

func NewCreatorHandler(
	creatorService accountsvc.CreatorService,
	stripeConnectService accountsvc.StripeConnectService,
	_ ...any,
) *CreatorHandler {
	return &CreatorHandler{
		creatorService:       creatorService,
		stripeConnectService: stripeConnectService,
	}
}

// OnboardCreator executa a Etapa 2 do onboarding de criador (`POST /api/account/onboarding`).
func (ch *CreatorHandler) OnboardCreator(w http.ResponseWriter, r *http.Request) {
	userID, ok := authmw.GetAuthenticatedUserID(r.Context())
	if !ok || userID == 0 {
		if user := authmw.Auth(r); user != nil && user.ID > 0 {
			userID = user.ID
		} else {
			web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
			return
		}
	}

	var input accountsvc.InputOnboardingCreator
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_JSON", "Corpo da requisição JSON inválido", nil)
		return
	}

	creator, err := ch.creatorService.OnboardCreator(userID, input)
	if err != nil {
		if errors.Is(err, accountsvc.ErrCreatorAlreadyExists) {
			web.RespondError(w, http.StatusConflict, "CREATOR_ALREADY_EXISTS", err.Error(), nil)
			return
		}
		web.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	if ch.stripeConnectService != nil {
		stripeAccountID, err := ch.stripeConnectService.CreateConnectAccount(creator)
		if err != nil {
			log.Printf("Erro ao criar conta Stripe Connect: %v", err)
		} else if stripeAccountID != "" {
			creator.StripeConnectAccountID = stripeAccountID
			if err := ch.creatorService.UpdateCreator(creator); err != nil {
				log.Printf("Erro ao atualizar creator com Stripe Account ID: %v", err)
			}
		}
	}

	web.RespondJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"creator": accountmodel.CreatorDTO{
			PublicID:               creator.PublicID,
			Name:                   creator.Name,
			SocialName:             creator.SocialName,
			Email:                  creator.Email,
			CPF:                    creator.CPF,
			Phone:                  creator.Phone,
			StripeConnectAccountID: creator.StripeConnectAccountID,
			OnboardingCompleted:    creator.OnboardingCompleted,
			ChargesEnabled:         creator.ChargesEnabled,
			PayoutsEnabled:         creator.PayoutsEnabled,
			NeedsStripeOnboarding:  creator.NeedsOnboarding(),
		},
	})
}
