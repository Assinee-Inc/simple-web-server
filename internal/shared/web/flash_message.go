package web

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
	"github.com/gorilla/sessions"
)

var _ FlashMessagePort = (*CookieFlashMessage)(nil)

var SessionStore = sessions.NewCookieStore([]byte("SESSION_KEY"))

type Message struct {
	Type    string // "success", "error", "info", "warning"
	Content string
}

type FlashMessagePort interface {
	Success(message string)
	Error(message string)
}

type FlashMessageFactory func(w http.ResponseWriter, r *http.Request) FlashMessagePort

type CookieFlashMessage struct {
	w http.ResponseWriter
	r *http.Request
}

func NewCookieFlashMessage(w http.ResponseWriter, r *http.Request) FlashMessagePort {
	return &CookieFlashMessage{
		w: w,
		r: r,
	}
}

func (fm *CookieFlashMessage) Success(message string) {
	log.Printf("[SUCCESS FLASH]: " + message)
	b, _ := json.Marshal(cookies.FlashMessage{
		Message: message,
		Type:    "success",
	})
	http.SetCookie(fm.w, &http.Cookie{
		Name:  "flash",
		Value: url.QueryEscape(string(b)),
		Path:  "/",
	})
}

func (fm *CookieFlashMessage) Error(message string) {
	log.Printf("[ERROR FLASH]: " + message)
	b, _ := json.Marshal(cookies.FlashMessage{
		Message: message,
		Type:    "danger",
	})
	http.SetCookie(fm.w, &http.Cookie{
		Name:  "flash",
		Value: url.QueryEscape(string(b)),
		Path:  "/",
	})
}
