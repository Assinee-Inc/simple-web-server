package middleware

import (
	"net/http"

	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

func TrialMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := authmw.Auth(r)
		if user == nil {
			web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
			return
		}

		// Skip trial check for these paths
		excludedPaths := map[string]bool{
			"/settings":        true,
			"/logout":          true,
			"/api/auth/logout": true,
			"/api/auth/me":     true,
		}

		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		if !user.IsInTrialPeriod() && !user.IsSubscribed() {
			web.RespondError(w, http.StatusForbidden, "SUBSCRIPTION_REQUIRED", "Assinatura ativa ou período de avaliação necessário para acessar este recurso", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
