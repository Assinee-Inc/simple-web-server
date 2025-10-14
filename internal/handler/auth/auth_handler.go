package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
)

type AuthHandler struct {
	userService      service.UserService
	sessionService   service.SessionService
	emailService     service.IEmailService
	templateRenderer template.TemplateRenderer
}

func NewAuthHandler(userService service.UserService, sessionService service.SessionService, emailService service.IEmailService, templateRenderer template.TemplateRenderer) *AuthHandler {
	return &AuthHandler{
		userService:      userService,
		sessionService:   sessionService,
		emailService:     emailService,
		templateRenderer: templateRenderer,
	}
}

// LoginView renders the login page with CSRF token
func (h *AuthHandler) LoginView(w http.ResponseWriter, r *http.Request) {
	csrfToken, err := h.sessionService.RegenerateCSRFToken(r, w)
	if err != nil {
		slog.Error("Failed to generate CSRF token", "error", err)
	}

	var form models.InputLogin
	h.ParseFormToData(&form, w, r)

	data := map[string]interface{}{
		"csrf_token": csrfToken,
		"FormErrors": h.sessionService.GetFlashes(w, r, "form-error"),
		"Success":    h.sessionService.GetFlashes(w, r, "success"),
		"Form":       form,
	}

	h.templateRenderer.View(w, r, "auth/login", data, "guest")
}

// LoginSubmit handles user login authentication
func (h *AuthHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	loginInput := models.InputLogin{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	// Authenticate user using UserService
	user, err := h.userService.AuthenticateUser(loginInput)
	if err != nil {
		slog.Error("Authentication failed", "error", err)
		h.sessionService.AddFlash(w, r, service.ErrInvalidCredentials.Error(), "form-error")
		h.SetFormToSession(w, r, loginInput)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Initialize session for an authenticated user
	h.sessionService.Set(r, w, service.UserIDKey, user.ID)
	h.sessionService.Set(r, w, service.UserEmailKey, user.Email)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// LogoutSubmit handles user logout
func (h *AuthHandler) LogoutSubmit(w http.ResponseWriter, r *http.Request) {
	h.sessionService.Destroy(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) ForgetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"FormErrors": h.sessionService.GetFlashes(w, r, "form-error"),
	}
	h.templateRenderer.View(w, r, "auth/forget-password", data, "guest")
}

func (h *AuthHandler) ForgetPasswordSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		h.sessionService.AddFlash(w, r, "Email é obrigatório", "form-error")
		http.Redirect(w, r, "/forget-password", http.StatusSeeOther)
		return
	}

	// Solicitar reset de senha
	err := h.userService.RequestPasswordReset(email)
	if err != nil {
		h.sessionService.AddFlash(w, r, "Erro ao processar solicitação de reset de senha", "form-error")
		http.Redirect(w, r, "/forget-password", http.StatusSeeOther)
		return
	}

	// Buscar usuário para enviar e-mail
	user := h.userService.FindByEmail(email)
	if user != nil {
		// Enviar e-mail de reset
		resetLink := r.Host + "/reset-password?token=" + user.PasswordResetToken
		go h.emailService.SendPasswordResetEmail(user.Username, user.Email, resetLink)
	}

	// Sempre redirecionar para sucesso (não revelar se o email existe ou não)
	http.Redirect(w, r, "/password-reset-success", http.StatusSeeOther)
}

func (h *AuthHandler) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.sessionService.AddFlash(w, r, "Token de reset inválido", "form-error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Token":      token,
		"FormErrors": h.sessionService.GetFlashes(w, r, "form-error"),
	}

	h.templateRenderer.View(w, r, "auth/reset-password", data, "guest")
}

func (h *AuthHandler) ResetPasswordSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	newPassword := r.FormValue("password")
	passwordConfirmation := r.FormValue("password_confirmation")

	if token == "" {
		slog.Warn("Reset password attempt with empty token")
		h.sessionService.AddFlash(w, r, "Token de reset inválido", "form-error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if newPassword == "" || passwordConfirmation == "" {
		slog.Warn("Reset password attempt with empty fields")
		h.sessionService.AddFlash(w, r, "Todos os campos são obrigatórios", "form-error")
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	if newPassword != passwordConfirmation {
		slog.Warn("Reset password attempt with non-matching passwords")
		h.sessionService.AddFlash(w, r, "As senhas não coincidem", "form-error")
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	// Reset the password using the UserService
	err := h.userService.ResetPassword(token, newPassword)
	if err != nil {
		slog.Warn("Failed to reset password", "error", err)
		h.sessionService.AddFlash(w, r, "Erro ao redefinir senha. Token pode estar expirado.", "form-error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Success
	slog.Info("Password reset successfully", "user", token)
	h.sessionService.AddFlash(w, r, "Senha redefinida com sucesso. Faça login com sua nova senha.", "success")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) SetFormToSession(w http.ResponseWriter, r *http.Request, form interface{}) {
	formData, _ := json.Marshal(form)
	err := h.sessionService.Set(r, w, "form", formData)
	if err != nil {
		slog.Error("Error in session manager: ", err)
		return
	}
}

func (h *AuthHandler) ParseFormToData(form interface{}, w http.ResponseWriter, r *http.Request) {
	formBytes := h.sessionService.Get(r, "form")
	if formBytes != nil {
		if data, ok := formBytes.([]byte); ok {
			err := json.Unmarshal(data, &form)
			if err != nil {
				slog.Error("Error in unmarshalling form: ", err)
				return
			}
		}
	}
	h.sessionService.Pop(r, w, "form")
}
