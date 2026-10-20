package handler

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	salesrepogorm "github.com/Assinee-Inc/docffy/internal/sales/repository/gorm"
	salesvc "github.com/Assinee-Inc/docffy/internal/sales/service"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
	"github.com/go-chi/chi/v5"
)

type ClientHandler struct {
	clientService  salesvc.ClientService
	creatorService accountsvc.CreatorService
}

func NewClientHandler(
	clientService salesvc.ClientService,
	creatorService accountsvc.CreatorService,
	_ ...any,
) *ClientHandler {
	return &ClientHandler{
		clientService:  clientService,
		creatorService: creatorService,
	}
}

func (ch *ClientHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	loggedUser := authmw.Auth(r)
	if loggedUser.ID == 0 {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
		return
	}

	clientPublicID := chi.URLParam(r, "id")
	client, err := ch.clientService.FindClientByPublicID(clientPublicID)
	if err != nil || client == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Cliente não encontrado")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"client":  client,
	})
}

func (ch *ClientHandler) ClientIndexView(w http.ResponseWriter, r *http.Request) {
	creator := accountmw.CreatorFromContext(r.Context())
	if creator == nil {
		loggedUser := authmw.Auth(r)
		if loggedUser.ID == 0 {
			web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
			return
		}
		var err error
		creator, err = ch.creatorService.FindCreatorByUserID(loggedUser.ID)
		if err != nil || creator == nil {
			web.RespondError(w, http.StatusNotFound, "CREATOR_NOT_FOUND", "Criador não encontrado")
			return
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	term := r.URL.Query().Get("term")

	pagination := salesmodel.NewPagination(page, perPage)

	clients, err := salesrepogorm.NewClientGormRepository().FindClientsByCreator(creator, salesmodel.ClientFilter{
		Term:       term,
		Pagination: pagination,
	})
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
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

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":     true,
		"clients":     clients,
		"pagination":  pagination,
		"search_term": term,
		"has_clients": hasClients,
		"filters": map[string]any{
			"term": term,
		},
	})
}

func (ch *ClientHandler) ClientUpdateSubmit(w http.ResponseWriter, r *http.Request) {
	userEmail := authmw.GetAuthenticatedUserEmail(r.Context())
	if userEmail == "" {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user email")
		return
	}

	clientPublicID := chi.URLParam(r, "id")
	client, err := ch.clientService.FindClientByPublicID(clientPublicID)
	if err != nil || client == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Cliente não encontrado")
		return
	}

	var email, phone string
	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		email = body.Email
		phone = body.Phone
	} else {
		email = r.FormValue("email")
		phone = r.FormValue("phone")
	}

	input := salesmodel.UpdateClientInput{
		ID:           client.ID,
		Email:        email,
		Phone:        phone,
		EmailCreator: userEmail,
	}

	updatedClient, err := ch.clientService.Update(input)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "UPDATE_ERROR", err.Error())
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Cliente foi atualizado!",
		"client":  updatedClient,
	})
}

func (ch *ClientHandler) ClientExportCSV(w http.ResponseWriter, r *http.Request) {
	userEmail := authmw.GetAuthenticatedUserEmail(r.Context())
	if userEmail == "" {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user email")
		return
	}

	clients, err := ch.clientService.ExportClients(userEmail)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "EXPORT_ERROR", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=clientes.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{"Nome", "Email", "Telefone"})
	if clients != nil {
		for _, client := range *clients {
			_ = writer.Write([]string{client.Name, client.Email, client.Phone})
		}
	}
}
