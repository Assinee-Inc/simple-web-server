package handler

import (
	"net/http"

	"github.com/anglesson/simple-web-server/pkg/template"
)

func HomeView(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		ErrorView(w, r, 404)
		return
	}

	template.View(w, r, "home", nil, "guest")
}
