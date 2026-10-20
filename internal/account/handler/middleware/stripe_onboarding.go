package middleware

import (
	"log"
	"net/http"
	"strings"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/config"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

// StripeOnboardingMiddleware verifica se o criador completou o onboarding do Stripe Connect.
// Retorna 403 Forbidden em JSON com código STRIPE_ONBOARDING_REQUIRED quando pendente.
func StripeOnboardingMiddleware(
	creatorService accountsvc.CreatorService,
	stripeConnectService accountsvc.StripeConnectService,
	deps ...any,
) func(http.Handler) http.Handler {
	var sessionService authsvc.SessionService
	for _, dep := range deps {
		if s, ok := dep.(authsvc.SessionService); ok {
			sessionService = s
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripeConfig := config.GetStripeOnboardingConfig()

			for _, path := range stripeConfig.ExcludedPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			creator := CreatorFromContext(r.Context())
			if creator == nil {
				var err error
				if userID, ok := authmw.GetAuthenticatedUserID(r.Context()); ok && userID > 0 {
					creator, err = creatorService.FindCreatorByUserID(userID)
				} else if email := authmw.GetAuthenticatedUserEmail(r.Context()); email != "" {
					creator, err = creatorService.FindCreatorByEmail(email)
				} else if sessionService != nil {
					if userEmail, sErr := sessionService.GetUserEmailFromSession(r); sErr == nil && userEmail != "" {
						creator, err = creatorService.FindCreatorByEmail(userEmail)
					} else {
						web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
						return
					}
				} else {
					web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
					return
				}

				if err != nil || creator == nil {
					web.RespondError(w, http.StatusForbidden, "CREATOR_ONBOARDING_REQUIRED", "Conclua o cadastro de criador para acessar este recurso", nil)
					return
				}
			}

			needsOnboarding := creator.StripeConnectAccountID == "" || !creator.OnboardingCompleted || !creator.ChargesEnabled || !creator.PayoutsEnabled

			if needsOnboarding {
				log.Printf("Creator %s precisa completar onboarding do Stripe", creator.Email)

				if creator.StripeConnectAccountID == "" {
					respondStripeOnboardingRequired(w, creator, "not_started")
					return
				}

				if stripeConnectService != nil {
					account, err := stripeConnectService.GetAccountDetails(creator.StripeConnectAccountID)
					if err != nil {
						log.Printf("Erro ao verificar conta Stripe: %v", err)
						respondStripeOnboardingRequired(w, creator, "pending")
						return
					}

					if err := stripeConnectService.UpdateCreatorFromAccount(creator, account); err != nil {
						log.Printf("Erro ao atualizar status do creator: %v", err)
					}

					if !account.DetailsSubmitted || !account.ChargesEnabled || !account.PayoutsEnabled {
						respondStripeOnboardingRequired(w, creator, "pending")
						return
					}
				} else {
					respondStripeOnboardingRequired(w, creator, "pending")
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(WithCreator(r.Context(), creator)))
		})
	}
}

func respondStripeOnboardingRequired(w http.ResponseWriter, creator *accountmodel.Creator, onboardingStatus string) {
	web.RespondJSON(w, http.StatusForbidden, map[string]any{
		"success":           false,
		"code":              "STRIPE_ONBOARDING_REQUIRED",
		"error":             "Complete o onboarding da conta de recebimento",
		"onboarding_status": onboardingStatus,
		"details": map[string]any{
			"has_stripe_account":   creator.StripeConnectAccountID != "",
			"onboarding_completed": creator.OnboardingCompleted,
			"charges_enabled":      creator.ChargesEnabled,
			"payouts_enabled":      creator.PayoutsEnabled,
			"remediation_link":     creator.OnboardingRefreshURL,
		},
	})
}
