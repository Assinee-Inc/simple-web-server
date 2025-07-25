package web

import (
	"net/http"

	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
)

func RedirectBackWithErrors(w http.ResponseWriter, r *http.Request, erroMessage string) {
	cookies.NotifyError(w, erroMessage)
	http.Redirect(w, r, r.Referer(), http.StatusBadRequest)
}
