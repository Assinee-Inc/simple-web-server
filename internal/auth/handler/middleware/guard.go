package middleware

import (
	"net/http"

	"github.com/anglesson/simple-web-server/internal/service"
)

func AuthGuard(sessionService service.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Se o usuário já está autenticado, redireciona para o dashboard
			if _, err := authorizer(r, sessionService); err == nil {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
				return
			}
			// Se não está autenticado, permite acesso (para páginas como login, register, etc.)
			next.ServeHTTP(w, r)
		})
	}
}
