package template

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/models"
	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
)

// TemplateRenderer interface for template rendering operations
type TemplateRenderer interface {
	View(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}, layout string)
	ViewWithoutLayout(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{})
}

// TemplateRendererImpl implements TemplateRenderer
type TemplateRendererImpl struct {
	templatePath string
	layoutPath   string
	partialPath  string
}

// NewTemplateRenderer creates a new template renderer instance
func NewTemplateRenderer(templatePath, layoutPath, partialPath string) TemplateRenderer {
	return &TemplateRendererImpl{
		templatePath: templatePath,
		layoutPath:   layoutPath,
		partialPath:  partialPath,
	}
}

// DefaultTemplateRenderer creates a template renderer with default paths
func DefaultTemplateRenderer() TemplateRenderer {
	return NewTemplateRenderer("web/pages/", "web/layouts/", "web/partials/")
}

type PageData struct {
	ErrorMessage string
}

// TemplateFunctions returns a map of functions available to templates
func TemplateFunctions(r *http.Request) template.FuncMap {
	return template.FuncMap{
		"appName": func() string {
			return config.AppConfig.AppName
		},
		"user": func() *models.User {
			return middleware.Auth(r)
		},
		"json": func(data any) (template.JS, error) {
			jsonData, err := json.Marshal(data)
			if err != nil {
				return "", err // Or handle error appropriately
			}
			return template.JS(jsonData), nil
		},
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		// Funções para manipulação de números
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"div": func(a, b int64) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		// Função para criar sequências de números para paginação
		"sequence": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		// Função para obter iniciais de um nome
		"getInitials": func(name string) string {
			if name == "" {
				return "?"
			}
			parts := strings.Fields(strings.TrimSpace(name))
			if len(parts) == 0 {
				return "?"
			}
			initials := ""
			for i, part := range parts {
				if i >= 2 { // Máximo 2 iniciais
					break
				}
				if len(part) > 0 {
					initials += strings.ToUpper(string(part[0]))
				}
			}
			if initials == "" {
				return "?"
			}
			return initials
		},
		// Funções para operações matemáticas de paginação
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		// Função para mascarar CPF ocultando os 6 dígitos do meio
		"maskCPF": func(cpf string) string {
			// Remove todos os caracteres não numéricos
			cleanCPF := ""
			for _, char := range cpf {
				if char >= '0' && char <= '9' {
					cleanCPF += string(char)
				}
			}

			// Verifica se tem 11 dígitos
			if len(cleanCPF) != 11 {
				return cpf // Retorna original se não for CPF válido
			}

			// Formata como 000.***-00
			return cleanCPF[:3] + ".***-" + cleanCPF[9:]
		},
	}
}

func (tr *TemplateRendererImpl) enrichData(w http.ResponseWriter, r *http.Request, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}

	// Get form data from cookies if available
	formCookie, err := r.Cookie("form")
	if err == nil {
		formValue, _ := url.QueryUnescape(formCookie.Value)
		var formData map[string]interface{}
		if err := json.Unmarshal([]byte(formValue), &formData); err == nil {
			data["Form"] = formData
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "form",
			MaxAge: -1,
		})
	}

	// Get error data from cookies if available
	errorsCookie, err := r.Cookie("errors")
	if err == nil {
		errorsValue, _ := url.QueryUnescape(errorsCookie.Value)
		var errorsData map[string]string
		if err := json.Unmarshal([]byte(errorsValue), &errorsData); err == nil {
			data["Errors"] = errorsData
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "errors",
			MaxAge: -1,
		})
	}

	// Get flash message from cookies if available
	if c, err := r.Cookie("flash"); err == nil {
		var flash cookies.FlashMessage
		decoded, _ := url.QueryUnescape(c.Value)
		if err := json.Unmarshal([]byte(decoded), &flash); err == nil {
			data["Flash"] = flash
		}
		http.SetCookie(w, &http.Cookie{Name: "flash", MaxAge: -1})
	}

	// Get CSRF token from context
	if csrfToken := middleware.GetCSRFToken(r); csrfToken != "" {
		data["csrf_token"] = csrfToken
	} else {
		log.Printf("CSRF token não encontrado no contexto")
	}

	// Get user from context
	if user := middleware.Auth(r); user != nil {
		data["user"] = user
		if user.CSRFToken != "" {
			data["csrf_token"] = user.CSRFToken
		}
	} else {
		log.Printf("Usuário não encontrado no contexto")
	}

	// Get subscription data from context
	if subscriptionData := middleware.GetSubscriptionData(r); subscriptionData != nil {
		data["SubscriptionStatus"] = subscriptionData.Status
		data["SubscriptionDaysLeft"] = subscriptionData.DaysLeft
	}

	return data
}

func (tr *TemplateRendererImpl) View(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}, layout string) {
	data = tr.enrichData(w, r, data)

	// Parse the template
	tmpl, err := template.New("").Funcs(TemplateFunctions(r)).ParseGlob(tr.layoutPath + "*.html")
	if err != nil {
		log.Printf("Erro ao carregar layouts: %v", err)
		http.Error(w, "Erro ao carregar página", http.StatusInternalServerError)
		return
	}

	// Parse partial templates
	_, err = tmpl.ParseGlob(tr.partialPath + "*.html")
	if err != nil {
		log.Printf("Erro ao carregar parciais: %v", err)
		http.Error(w, "Erro ao carregar página", http.StatusInternalServerError)
		return
	}

	// Parse the page template
	_, err = tmpl.ParseFiles(tr.templatePath + page + ".html")
	if err != nil {
		log.Printf("Erro ao carregar página: %v", err)
		http.Error(w, "Erro ao carregar página", http.StatusInternalServerError)
		return
	}

	// Execute the template
	err = tmpl.ExecuteTemplate(w, layout, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar página", http.StatusInternalServerError)
		return
	}
}

func (tr *TemplateRendererImpl) ViewWithoutLayout(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	data = tr.enrichData(w, r, data)

	// Parse the page template directly
	tmpl, err := template.New("").Funcs(TemplateFunctions(r)).ParseFiles(tr.templatePath + page + ".html")
	if err != nil {
		log.Printf("Erro ao carregar página: %v", err)
		http.Error(w, "Erro ao carregar página", http.StatusInternalServerError)
		return
	}

	// Execute the template - use the page name as template name
	err = tmpl.ExecuteTemplate(w, page, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar página", http.StatusInternalServerError)
		return
	}
}

// Legacy functions for backward compatibility
func View(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}, layout string) {
	renderer := DefaultTemplateRenderer()
	renderer.View(w, r, page, data, layout)
}

func ViewWithoutLayout(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	renderer := DefaultTemplateRenderer()
	renderer.ViewWithoutLayout(w, r, page, data)
}
