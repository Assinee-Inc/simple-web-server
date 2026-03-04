package middleware

import (
	"log"
	"net/http"
	"strings"

	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	"github.com/anglesson/simple-web-server/internal/config"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
)

// StripeOnboardingMiddleware verifica se o usuário completou o onboarding do Stripe
// antes de permitir acesso aos recursos protegidos
func StripeOnboardingMiddleware(
	creatorService accountsvc.CreatorService,
	stripeConnectService accountsvc.StripeConnectService,
	sessionService authsvc.SessionService,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripeConfig := config.GetStripeOnboardingConfig()

			for _, path := range stripeConfig.ExcludedPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			userEmail, err := sessionService.GetUserEmailFromSession(r)
			if err != nil {
				log.Printf("Erro ao obter email da sessão: %v", err)
				http.Error(w, "Sessão inválida", http.StatusUnauthorized)
				return
			}

			creator, err := creatorService.FindCreatorByEmail(userEmail)
			if err != nil {
				log.Printf("Erro ao buscar creator: %v", err)
				http.Error(w, "Creator não encontrado", http.StatusNotFound)
				return
			}

			needsOnboarding := creator.StripeConnectAccountID == "" || !creator.OnboardingCompleted || !creator.ChargesEnabled || !creator.PayoutsEnabled

			if needsOnboarding {
				log.Printf("Creator %s precisa completar onboarding do Stripe", creator.Email)

				if creator.StripeConnectAccountID == "" {
					http.Redirect(w, r, "/stripe-connect/welcome", http.StatusSeeOther)
					return
				}

				account, err := stripeConnectService.GetAccountDetails(creator.StripeConnectAccountID)
				if err != nil {
					log.Printf("Erro ao verificar conta Stripe: %v", err)
					http.Redirect(w, r, "/stripe-connect/status", http.StatusSeeOther)
					return
				}

				err = stripeConnectService.UpdateCreatorFromAccount(creator, account)
				if err != nil {
					log.Printf("Erro ao atualizar status do creator: %v", err)
				}

				if !account.DetailsSubmitted || !account.ChargesEnabled || !account.PayoutsEnabled {
					http.Redirect(w, r, "/stripe-connect/status", http.StatusSeeOther)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
