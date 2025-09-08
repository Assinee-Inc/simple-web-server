package handler

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/anglesson/simple-web-server/internal/handler/web"
	"github.com/anglesson/simple-web-server/internal/repository/gorm"

	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	cookies "github.com/anglesson/simple-web-server/pkg/cookie"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

type ClientHandler struct {
	clientService       service.ClientService
	creatorService      service.CreatorService
	flashMessageFactory web.FlashMessageFactory
	templateRenderer    template.TemplateRenderer
}

func NewClientHandler(clientService service.ClientService, creatorService service.CreatorService, flashMessageFactory web.FlashMessageFactory, templateRenderer template.TemplateRenderer) *ClientHandler {
	return &ClientHandler{
		clientService:       clientService,
		creatorService:      creatorService,
		flashMessageFactory: flashMessageFactory,
		templateRenderer:    templateRenderer,
	}
}

func (ch *ClientHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	loggedUser := middleware.Auth(r)
	if loggedUser.ID == 0 {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusInternalServerError)
		return
	}

	ch.templateRenderer.View(w, r, "client/create", nil, "admin")
}

func (ch *ClientHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	loggedUser := middleware.Auth(r)
	if loggedUser.ID == 0 {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusInternalServerError)
		return
	}

	clientID := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(clientID, 10, 32)
	client, err := ch.clientService.FindCreatorsClientByID(uint(id), loggedUser.Email)
	if err != nil {
		http.Redirect(w, r, r.Referer(), http.StatusNotFound)
	}

	ch.templateRenderer.View(w, r, "client/update", map[string]interface{}{"Client": client}, "admin")
}

func (ch *ClientHandler) ClientIndexView(w http.ResponseWriter, r *http.Request) {
	loggedUser := middleware.Auth(r)
	if loggedUser.ID == 0 {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusInternalServerError)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	term := r.URL.Query().Get("term")

	pagination := models.NewPagination(page, perPage)

	log.Printf("User Logado: %v", loggedUser.Email)

	creator, err := ch.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		web.RedirectBackWithErrors(w, r, err.Error())
	}

	clients, err := gorm.NewClientGormRepository().FindClientsByCreator(creator, models.ClientFilter{
		Term:       term,
		Pagination: pagination,
	})
	if err != nil {
		web.RedirectBackWithErrors(w, r, err.Error())
		return
	}

	// Get total count for pagination (this should be a separate query for accurate pagination)
	// For now, we'll use the length of the result, but this should be optimized
	totalCount := int64(0)
	if clients != nil {
		totalCount = int64(len(*clients))
	}
	pagination.SetTotal(totalCount)

	// Ensure clients is never nil
	if clients == nil {
		clients = &[]models.Client{}
	}

	// Check if there are any clients
	hasClients := clients != nil && len(*clients) > 0

	ch.templateRenderer.View(w, r, "client/list", map[string]any{
		"Clients":    clients,
		"Pagination": pagination,
		"SearchTerm": term,
		"HasClients": hasClients,
	}, "admin")
}

// redirectWithFormData is a helper function to redirect with form data and errors
func (ch *ClientHandler) redirectWithFormData(w http.ResponseWriter, r *http.Request, formData map[string]interface{}, errors map[string]string) {
	formJSON, _ := json.Marshal(formData)
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

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

func (ch *ClientHandler) ClientCreateSubmit(w http.ResponseWriter, r *http.Request) {
	flashMessage := ch.flashMessageFactory(w, r)

	user_email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		flashMessage.Error("Unauthorized. Invalid user email")
		http.Error(w, "Invalid user email", http.StatusUnauthorized)
		return
	}

	input := models.CreateClientInput{
		Name:      r.FormValue("name"),
		CPF:       r.FormValue("cpf"),
		BirthDate: r.FormValue("birthdate"),
		Email:     r.FormValue("email"),
		Phone:     r.FormValue("phone"),
	}

	input.EmailCreator = user_email

	// TODO: Validar se o cliente existe
	_, err := ch.clientService.CreateClient(input)
	if err != nil {
		// Salvar dados do formulário em cookies para persistir após erro
		formData := map[string]interface{}{
			"Name":      input.Name,
			"CPF":       input.CPF,
			"Birthdate": input.BirthDate,
			"Email":     input.Email,
			"Phone":     input.Phone,
		}

		errors := map[string]string{
			"general": err.Error(),
		}

		ch.redirectWithFormData(w, r, formData, errors)
		return
	}
	flashMessage.Success("Cliente foi cadastrado!")

	http.Redirect(w, r, "/client", http.StatusSeeOther)
}

func (ch *ClientHandler) ClientUpdateSubmit(w http.ResponseWriter, r *http.Request) {
	flashMessage := ch.flashMessageFactory(w, r)
	user_email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		http.Error(w, "Invalid user email", http.StatusInternalServerError)
		return
	}

	clientID := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(clientID, 10, 32)

	input := models.UpdateClientInput{
		ID:           uint(id),
		Email:        r.FormValue("email"),
		Phone:        r.FormValue("phone"),
		EmailCreator: user_email,
	}

	_, err := ch.clientService.Update(input)
	if err != nil {
		// Salvar dados do formulário em cookies para persistir após erro
		formData := map[string]interface{}{
			"Email": input.Email,
			"Phone": input.Phone,
		}

		errors := map[string]string{
			"general": err.Error(),
		}

		ch.redirectWithFormData(w, r, formData, errors)
		return
	}
	flashMessage.Success("Cliente foi atualizado!")

	http.Redirect(w, r, "/client", http.StatusSeeOther)
}

func (ch *ClientHandler) ClientImportSubmit(w http.ResponseWriter, r *http.Request) {
	log.Println("Iniciando processamento de CSV")
	user_email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		http.Error(w, "Invalid user email", http.StatusInternalServerError)
		return
	}

	creator, err := ch.creatorService.FindCreatorByEmail(user_email)
	if err != nil {
		log.Println("Nao autorizado")
		http.Redirect(w, r, r.Referer(), http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Erro ao processar o formulário", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		web.RedirectBackWithErrors(w, r, "Erro ao ler o arquivo")
	}
	defer file.Close()

	// Verifica a extensão do arquivo (opcional)
	if !strings.HasSuffix(handler.Filename, ".csv") {
		web.RedirectBackWithErrors(w, r, "Arquivo não é CSV")
	}

	log.Println("Arquivo validado!")

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Printf("Erro na leitura do CSV: %s", err.Error())
		web.RedirectBackWithErrors(w, r, "Erro na leitura do CSV")
	}

	// Validate header
	var clients []*models.Client

	for i, linha := range rows {
		log.Printf("linha: %s", linha)
		if i > 0 {
			client := models.NewClient(linha[0], linha[1], linha[2], linha[3], linha[4], creator)
			clients = append(clients, client)
		}
	}

	if err = ch.clientService.CreateBatchClient(clients); err != nil {
		web.RedirectBackWithErrors(w, r, err.Error())
		return
	}

	cookies.NotifySuccess(w, "Clientes foram importados!")
	http.Redirect(w, r, "/client", http.StatusSeeOther)
}
