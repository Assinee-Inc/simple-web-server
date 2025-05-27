package middlewares

import (
	"net/http"
)

func TrialMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := Auth(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Skip trial check for these paths
		excludedPaths := map[string]bool{
			"/settings": true,
			"/logout":   true,
		}

		if excludedPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		if !user.IsInTrialPeriod() {
			http.Redirect(w, r, "/settings", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
