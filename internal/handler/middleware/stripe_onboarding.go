package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/service"
)

// StripeOnboardingMiddleware verifica se o usuário completou o onboarding do Stripe
// antes de permitir acesso aos recursos protegidos
func StripeOnboardingMiddleware(
	creatorService service.CreatorService,
	stripeConnectService service.StripeConnectService,
	sessionService service.SessionService,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obter configuração de rotas
			stripeConfig := config.GetStripeOnboardingConfig()

			// Verificar se a rota atual está nas exclusões
			for _, path := range stripeConfig.ExcludedPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Obter email do usuário da sessão
			userEmail, err := sessionService.GetUserEmailFromSession(r)
			if err != nil {
				log.Printf("Erro ao obter email da sessão: %v", err)
				http.Error(w, "Sessão inválida", http.StatusUnauthorized)
				return
			}

			// Buscar creator pelo email
			creator, err := creatorService.FindCreatorByEmail(userEmail)
			if err != nil {
				log.Printf("Erro ao buscar creator: %v", err)
				http.Error(w, "Creator não encontrado", http.StatusNotFound)
				return
			}

			// Verificar se o creator tem conta Stripe e se o onboarding está completo
			needsOnboarding := creator.StripeConnectAccountID == "" || !creator.OnboardingCompleted || !creator.ChargesEnabled

			if needsOnboarding {
				log.Printf("Creator %s precisa completar onboarding do Stripe", creator.Email)

				// Se não tem conta Stripe, redirecionar para página de boas-vindas
				if creator.StripeConnectAccountID == "" {
					http.Redirect(w, r, "/stripe-connect/welcome", http.StatusSeeOther)
					return
				}

				// Se tem conta mas onboarding não está completo, verificar status atual
				account, err := stripeConnectService.GetAccountDetails(creator.StripeConnectAccountID)
				if err != nil {
					log.Printf("Erro ao verificar conta Stripe: %v", err)
					http.Redirect(w, r, "/stripe-connect/status", http.StatusSeeOther)
					return
				}

				// Atualizar status do creator com dados mais recentes do Stripe
				err = stripeConnectService.UpdateCreatorFromAccount(creator, account)
				if err != nil {
					log.Printf("Erro ao atualizar status do creator: %v", err)
				}

				// Se ainda não está completo, redirecionar para status/onboarding
				if !account.DetailsSubmitted || !creator.PayoutsEnabled || !creator.ChargesEnabled {
					http.Redirect(w, r, "/stripe-connect/welcome", http.StatusSeeOther)
					return
				}
			}

			// Se chegou até aqui, o onboarding está completo
			next.ServeHTTP(w, r)
		})
	}
}
