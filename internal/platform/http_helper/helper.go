package http_helper

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Se mudar de framework, você só muda a implementação desta função
func GetParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
	// ou return c.Param(key) se for Gin
}
