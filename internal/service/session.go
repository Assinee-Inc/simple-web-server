package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	CSRFTokenKey = "csrf_token"
)

// SessionService define a interface unificada para gestão de sessão.
type SessionService interface {
	Set(r *http.Request, w http.ResponseWriter, key string, value interface{}) error
	Get(r *http.Request, key string) interface{}
	Pop(r *http.Request, w http.ResponseWriter, key string) interface{}
	Destroy(r *http.Request, w http.ResponseWriter) error
	AddFlash(w http.ResponseWriter, r *http.Request, message, key string) error
	GetFlashes(w http.ResponseWriter, r *http.Request, key string) []string
	RegenerateCSRFToken(r *http.Request, w http.ResponseWriter) (string, error)
	GetUserEmailFromSession(r *http.Request) (string, error)
	InitSession(w http.ResponseWriter, r *http.Request, userID uint, userEmail string) error
}

// gorillaSessionService é a implementação que usa gorilla/sessions.
type gorillaSessionService struct {
	store       sessions.Store
	sessionName string
}

// NewSessionService cria uma nova instância do nosso gerenciador de sessão.
func NewSessionService(store sessions.Store, sessionName string) SessionService {
	return &gorillaSessionService{
		store:       store,
		sessionName: sessionName,
	}
}

func (s *gorillaSessionService) InitSession(w http.ResponseWriter, r *http.Request, userID uint, userEmail string) error {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one.", "error", err, "path", r.URL.Path)
	}
	session.Values[UserIDKey] = userID
	session.Values[UserEmailKey] = userEmail
	return session.Save(r, w)
}

func (s *gorillaSessionService) GetUserEmailFromSession(r *http.Request) (string, error) {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one.", "error", err, "path", r.URL.Path)
	}
	val := session.Values[UserEmailKey]
	email, ok := val.(string)
	if !ok || email == "" {
		return "", errors.New("user email not found in session")
	}
	return email, nil
}

func (s *gorillaSessionService) Set(r *http.Request, w http.ResponseWriter, key string, value interface{}) error {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one.", "error", err, "path", r.URL.Path)
	}
	session.Values[key] = value
	return session.Save(r, w)
}

func (s *gorillaSessionService) Get(r *http.Request, key string) interface{} {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one.", "error", err, "path", r.URL.Path)
	}
	return session.Values[key]
}

func (s *gorillaSessionService) Pop(r *http.Request, w http.ResponseWriter, key string) interface{} {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one.", "error", err, "path", r.URL.Path)
		return nil
	}
	value := session.Values[key]
	delete(session.Values, key)
	err = session.Save(r, w)
	if err != nil {
		slog.Error("Failed to save session in Pop", "error", err)
		return nil
	}
	return value
}

func (s *gorillaSessionService) Destroy(r *http.Request, w http.ResponseWriter) error {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one to delete.", "error", err, "path", r.URL.Path)
	}
	session.Options.MaxAge = -1 // Deleta o cookie
	return session.Save(r, w)
}

func (s *gorillaSessionService) AddFlash(w http.ResponseWriter, r *http.Request, message, key string) error {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one.", "error", err, "path", r.URL.Path)
	}
	slog.Info("Adding flash message", "message", message, "key", key)
	slog.Info("Session values", "values", session.Values)
	session.AddFlash(message, key)
	err = session.Save(r, w)
	if err != nil {
		slog.Error("Failed to save session in AddFlash", "error", err)
	}
	return err
}

func (s *gorillaSessionService) GetFlashes(w http.ResponseWriter, r *http.Request, key string) []string {
	session, err := s.store.Get(r, s.sessionName)
	if err != nil {
		slog.Warn("Failed to get existing session, creating new one.", "error", err, "path", r.URL.Path)
		return []string{}
	}

	flashes := session.Flashes(key)
	if len(flashes) > 0 {
		if err := session.Save(r, w); err != nil {
			slog.Error("Failed to save session in GetFlashes", "error", err)
			return []string{}
		}
	}

	var messages []string
	for _, f := range flashes {
		if msg, ok := f.(string); ok {
			messages = append(messages, msg)
		}
	}
	return messages
}

func (s *gorillaSessionService) RegenerateCSRFToken(r *http.Request, w http.ResponseWriter) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := base64.StdEncoding.EncodeToString(b)
	err = s.Set(r, w, CSRFTokenKey, token)
	if err != nil {
		slog.Error("Failed to set CSRF token in session", "error", err)
		return "", err
	}
	return token, nil
}
