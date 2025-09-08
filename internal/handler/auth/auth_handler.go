package auth

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
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
	csrfToken := h.sessionService.GenerateCSRFToken()
	h.sessionService.SetCSRFToken(w)

	data := map[string]interface{}{
		"csrf_token": csrfToken,
	}

	// Check for rate limit error
	if r.URL.Query().Get("error") == "rate_limit_exceeded" {
		data["rate_limit_error"] = "Muitas tentativas. Aguarde alguns minutos antes de tentar novamente."
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

	errors := make(map[string]string)

	// Validate required fields
	if loginInput.Email == "" {
		errors["email"] = "Email é obrigatório."
	}
	if loginInput.Password == "" {
		errors["password"] = "Senha é obrigatória."
	}

	// If there are validation errors, redirect back with errors
	if len(errors) > 0 {
		h.redirectWithErrors(w, r, loginInput, errors)
		return
	}

	// Authenticate user using UserService
	user, err := h.userService.AuthenticateUser(loginInput)
	if err != nil {
		errors["password"] = "Email ou senha inválidos"
		h.redirectWithErrors(w, r, loginInput, errors)
		return
	}

	// Initialize session for authenticated user
	h.sessionService.InitSession(w, user.Email)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// LogoutSubmit handles user logout
func (h *AuthHandler) LogoutSubmit(w http.ResponseWriter, r *http.Request) {
	h.sessionService.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) ForgetPasswordView(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}

	// Check for rate limit error
	if r.URL.Query().Get("error") == "rate_limit_exceeded" {
		data["rate_limit_error"] = "Muitas tentativas de recuperação de senha. Aguarde alguns minutos antes de tentar novamente."
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
		cookies.NotifyError(w, "Email é obrigatório")
		http.Redirect(w, r, "/forget-password", http.StatusSeeOther)
		return
	}

	// Solicitar reset de senha
	err := h.userService.RequestPasswordReset(email)
	if err != nil {
		cookies.NotifyError(w, "Erro ao processar solicitação de reset de senha")
		http.Redirect(w, r, "/forget-password", http.StatusSeeOther)
		return
	}

	// Buscar usuário para enviar e-mail
	user := h.userService.FindByEmail(email)
	if user != nil {
		// Enviar e-mail de reset
		resetLink := config.AppConfig.Host + ":" + config.AppConfig.Port + "/reset-password?token=" + user.PasswordResetToken
		go h.emailService.SendPasswordResetEmail(user.Username, user.Email, resetLink)
	}

	// Sempre redirecionar para sucesso (não revelar se o email existe ou não)
	http.Redirect(w, r, "/password-reset-success", http.StatusSeeOther)
}

func (h *AuthHandler) ResetPasswordView(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		cookies.NotifyError(w, "Token de reset inválido")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Token": token,
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
	confirmPassword := r.FormValue("confirm_password")

	if token == "" {
		cookies.NotifyError(w, "Token de reset inválido")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if newPassword == "" || confirmPassword == "" {
		cookies.NotifyError(w, "Todos os campos são obrigatórios")
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	if newPassword != confirmPassword {
		cookies.NotifyError(w, "As senhas não coincidem")
		http.Redirect(w, r, "/reset-password?token="+token, http.StatusSeeOther)
		return
	}

	// Reset the password using the UserService
	err := h.userService.ResetPassword(token, newPassword)
	if err != nil {
		cookies.NotifyError(w, "Erro ao redefinir senha. Token pode estar expirado.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	cookies.NotifySuccess(w, "Senha redefinida com sucesso. Faça login com sua nova senha.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// redirectWithErrors is a helper function to redirect with form data and errors
func (h *AuthHandler) redirectWithErrors(w http.ResponseWriter, r *http.Request, form interface{}, errors map[string]string) {
	formJSON, _ := json.Marshal(form)
	errorsJSON, _ := json.Marshal(errors)

	http.SetCookie(w, &http.Cookie{
		Name:  "form",
		Value: url.QueryEscape(string(formJSON)),
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "errors",
		Value: url.QueryEscape(string(errorsJSON)),
		Path:  "/",
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
