package handler

import (
	"encoding/csv"
	"log"
	"net/http"
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
}

func NewClientHandler(clientService service.ClientService, creatorService service.CreatorService, flashMessageFactory web.FlashMessageFactory) *ClientHandler {
	return &ClientHandler{
		clientService:       clientService,
		creatorService:      creatorService,
		flashMessageFactory: flashMessageFactory,
	}
}

func (ch *ClientHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	loggedUser := middleware.Auth(r)
	if loggedUser.ID == 0 {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusInternalServerError)
		return
	}

	template.View(w, r, "client/create", nil, "admin")
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

	template.View(w, r, "client/update", map[string]interface{}{"Client": client}, "admin")
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
	}

	// Set total count for pagination
	if clients != nil {
		pagination.SetTotal(int64(len(*clients)))
	}

	template.View(w, r, "client", map[string]any{
		"Clients":    clients,
		"Pagination": pagination,
	}, "admin")
}

func (ch *ClientHandler) ClientCreateSubmit(w http.ResponseWriter, r *http.Request) {
	flashMessage := ch.flashMessageFactory(w, r)

	user_email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		flashMessage.Error("Unauthorized. Invalid user email")
		http.Error(w, "Invalid user email", http.StatusUnauthorized)
		return
	}

	input := service.CreateClientInput{
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
		flashMessage.Error(err.Error())
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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

	input := service.UpdateClientInput{
		ID:           uint(id),
		Email:        r.FormValue("email"),
		Phone:        r.FormValue("phone"),
		EmailCreator: user_email,
	}

	_, err := ch.clientService.Update(input)
	if err != nil {
		flashMessage.Error(err.Error())
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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
