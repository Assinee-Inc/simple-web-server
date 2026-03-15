package handler

import (
	"encoding/csv"
	"log"
	"net/http"
	"strconv"

	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepogorm "github.com/anglesson/simple-web-server/internal/sales/repository/gorm"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

type ClientHandler struct {
	clientService    salesvc.ClientService
	creatorService   accountsvc.CreatorService
	sessionManager   authsvc.SessionService
	templateRenderer template.TemplateRenderer
}

func NewClientHandler(
	clientService salesvc.ClientService,
	creatorService accountsvc.CreatorService,
	sessionManager authsvc.SessionService,
	templateRenderer template.TemplateRenderer,
) *ClientHandler {
	return &ClientHandler{
		clientService:    clientService,
		creatorService:   creatorService,
		sessionManager:   sessionManager,
		templateRenderer: templateRenderer,
	}
}

func (ch *ClientHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	loggedUser := authmw.Auth(r)
	if loggedUser.ID == 0 {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusInternalServerError)
		return
	}

	clientPublicID := chi.URLParam(r, "id")
	client, err := ch.clientService.FindClientByPublicID(clientPublicID)
	if err != nil {
		http.Redirect(w, r, r.Referer(), http.StatusNotFound)
		return
	}

	ch.templateRenderer.View(w, r, "client/update", map[string]any{"Client": client}, "admin-daisy")
}

func (ch *ClientHandler) ClientIndexView(w http.ResponseWriter, r *http.Request) {
	loggedUser := authmw.Auth(r)
	if loggedUser.ID == 0 {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusInternalServerError)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	term := r.URL.Query().Get("term")

	pagination := salesmodel.NewPagination(page, perPage)

	log.Printf("User Logado: %v", loggedUser.Email)

	creator, err := ch.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	clients, err := salesrepogorm.NewClientGormRepository().FindClientsByCreator(creator, salesmodel.ClientFilter{
		Term:       term,
		Pagination: pagination,
	})
	if err != nil {
		log.Printf("Erro ao buscar clientes: %v", err)
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	totalCount := int64(0)
	if clients != nil {
		totalCount = int64(len(*clients))
		if len(*clients) == pagination.Limit {
			totalCount = int64(pagination.Limit * (pagination.Page + 1))
		}
	}
	pagination.SetTotal(totalCount)

	if clients == nil {
		clients = &[]salesmodel.Client{}
	}

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
		"Filters": map[string]interface{}{
			"term": term,
		},
	}, "admin-daisy")
}

func (ch *ClientHandler) ClientUpdateSubmit(w http.ResponseWriter, r *http.Request) {
	user_email, ok := r.Context().Value(authmw.UserEmailKey).(string)
	if !ok {
		ch.sessionManager.AddFlash(w, r, "Invalid user email", "error")
		http.Error(w, "Invalid user email", http.StatusInternalServerError)
		return
	}

	clientPublicID := chi.URLParam(r, "id")
	client, err := ch.clientService.FindClientByPublicID(clientPublicID)
	if err != nil {
		ch.sessionManager.AddFlash(w, r, "Cliente não encontrado", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	input := salesmodel.UpdateClientInput{
		ID:           client.ID,
		Email:        r.FormValue("email"),
		Phone:        r.FormValue("phone"),
		EmailCreator: user_email,
	}

	_, err = ch.clientService.Update(input)
	if err != nil {
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}
	ch.sessionManager.AddFlash(w, r, "Cliente foi atualizado!", "success")

	http.Redirect(w, r, "/client", http.StatusSeeOther)
}

func (ch *ClientHandler) ClientExportCSV(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
	if !ok {
		http.Error(w, "Invalid user email", http.StatusInternalServerError)
		return
	}

	clients, err := ch.clientService.ExportClients(userEmail)
	if err != nil {
		ch.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/client", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=clientes.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"Nome", "Email", "Telefone", "Data Nascimento"})
	for _, client := range *clients {
		writer.Write([]string{client.Name, client.Email, client.Phone, client.Birthdate})
	}
}

