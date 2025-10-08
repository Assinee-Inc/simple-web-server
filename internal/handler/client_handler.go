package handler

import (
	"encoding/csv"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/anglesson/simple-web-server/internal/repository/gorm"

	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

type ClientHandler struct {
	clientService    service.ClientService
	creatorService   service.CreatorService
	sessionManager   service.SessionService
	templateRenderer template.TemplateRenderer
}

func NewClientHandler(
	clientService service.ClientService,
	creatorService service.CreatorService,
	sessionManager service.SessionService,
	templateRenderer template.TemplateRenderer,
) *ClientHandler {
	return &ClientHandler{
		clientService:    clientService,
		creatorService:   creatorService,
		sessionManager:   sessionManager,
		templateRenderer: templateRenderer,
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

	ch.templateRenderer.View(w, r, "client/update", map[string]any{"Client": client}, "admin")
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
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Buscar clientes
	clients, err := gorm.NewClientGormRepository().FindClientsByCreator(creator, models.ClientFilter{
		Term:       term,
		Pagination: pagination,
	})
	if err != nil {
		log.Printf("Erro ao buscar clientes: %v", err)
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Para uma implementação mais robusta, deveríamos fazer uma query de contagem separada
	// Por enquanto, vamos definir um total baseado na quantidade retornada
	totalCount := int64(0)
	if clients != nil {
		totalCount = int64(len(*clients))
		// Se retornou o máximo por página, provavelmente há mais registros
		if len(*clients) == pagination.Limit {
			totalCount = int64(pagination.Limit * (pagination.Page + 1)) // Estimativa
		}
	}
	pagination.SetTotal(totalCount)

	// Ensure clients is never nil
	if clients == nil {
		clients = &[]models.Client{}
	}

	// Check if there are any clients
	hasClients := clients != nil && len(*clients) > 0

	log.Printf("Encontrados %d clientes para exibição", len(*clients))

	successMessages := ch.sessionManager.GetFlashes(w, r, "success")
	errorMessages := ch.sessionManager.GetFlashes(w, r, "error")

	ch.templateRenderer.View(w, r, "client/list", map[string]any{
		"Clients":    clients,
		"Pagination": pagination,
		"SearchTerm": term,
		"HasClients": hasClients,
		"Success":    successMessages,
		"Errors":     errorMessages,
	}, "admin")
}

func (ch *ClientHandler) ClientCreateSubmit(w http.ResponseWriter, r *http.Request) {
	user_email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		ch.sessionManager.AddFlash(w, r, "Unauthorized. Invalid user email", "error")
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
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}
	ch.sessionManager.AddFlash(w, r, "Cliente foi cadastrado!", "success")

	http.Redirect(w, r, "/client", http.StatusSeeOther)
}

func (ch *ClientHandler) ClientUpdateSubmit(w http.ResponseWriter, r *http.Request) {
	user_email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		ch.sessionManager.AddFlash(w, r, "Invalid user email", "error")
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
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}
	ch.sessionManager.AddFlash(w, r, "Cliente foi atualizado!", "success")

	http.Redirect(w, r, "/client", http.StatusSeeOther)
}

func (ch *ClientHandler) ClientImportSubmit(w http.ResponseWriter, r *http.Request) {
	log.Println("Iniciando processamento de CSV")
	user_email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		ch.sessionManager.AddFlash(w, r, "Invalid user email", "error")
		http.Error(w, "Invalid user email", http.StatusInternalServerError)
		return
	}

	creator, err := ch.creatorService.FindCreatorByEmail(user_email)
	if err != nil {
		log.Println("Nao autorizado")
		ch.sessionManager.AddFlash(w, r, "Não autorizado", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Erro ao processar o formulário", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		ch.sessionManager.AddFlash(w, r, "Erro ao ler o arquivo", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Verifica a extensão do arquivo (opcional)
	if !strings.HasSuffix(handler.Filename, ".csv") {
		ch.sessionManager.AddFlash(w, r, "Arquivo não é CSV", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	log.Println("Arquivo validado!")

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Printf("Erro na leitura do CSV: %s", err.Error())
		ch.sessionManager.AddFlash(w, r, "Erro na leitura do CSV", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
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
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	ch.sessionManager.AddFlash(w, r, "Clientes foram importados!", "success")
	http.Redirect(w, r, "/client", http.StatusSeeOther)
}
