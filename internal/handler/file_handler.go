package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/repository/gorm"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/go-chi/chi/v5"
)

type FileHandler struct {
	fileService      service.FileService
	sessionManager   service.SessionService
	templateRenderer template.TemplateRenderer
}

func NewFileHandler(fileService service.FileService, sessionManager service.SessionService, templateRenderer template.TemplateRenderer) *FileHandler {
	return &FileHandler{
		fileService:      fileService,
		sessionManager:   sessionManager,
		templateRenderer: templateRenderer,
	}
}

// FileIndexView exibe a lista de arquivos do criador
func (h *FileHandler) FileIndexView(w http.ResponseWriter, r *http.Request) {
	creatorID := h.getCreatorIDFromSession(r)
	if creatorID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Obter parâmetros de paginação e busca
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	searchTerm := r.URL.Query().Get("search")
	fileType := r.URL.Query().Get("type")

	// Criar paginação
	pagination := models.NewPagination(page, perPage)

	// Log para debug
	log.Printf("Buscando arquivos para creator ID: %d, página: %d, por página: %d", creatorID, page, perPage)

	// Criar query para busca paginada
	query := repository.FileQuery{
		CreatorID:  creatorID,
		FileType:   fileType,
		SearchTerm: searchTerm,
		Pagination: pagination,
	}

	// Buscar arquivos com paginação
	files, total, err := h.fileService.GetFilesByCreatorPaginated(creatorID, query)
	if err != nil {
		log.Printf("Erro ao buscar arquivos: %v", err)
		h.sessionManager.AddFlash(w, r, "Erro ao carregar arquivos", "error")
		http.Error(w, "Erro ao carregar arquivos", http.StatusInternalServerError)
		return
	}

	// Configurar paginação com total
	pagination.SetTotal(total)

	// Adicionar parâmetros de busca à paginação
	pagination.SearchTerm = searchTerm
	pagination.FileType = fileType

	// Log para debug
	log.Printf("Arquivos encontrados: %d de %d total", len(files), total)

	successMessages := h.sessionManager.GetFlashes(w, r, "success")
	errorMessages := h.sessionManager.GetFlashes(w, r, "error")

	data := map[string]interface{}{
		"Files":      files,
		"Pagination": pagination,
		"Title":      "Minha Biblioteca de Arquivos",
		"Success":    successMessages,
		"Errors":     errorMessages,
	}

	h.templateRenderer.View(w, r, "file/index", data, "admin")
}

// FileUploadView exibe o formulário de upload
func (h *FileHandler) FileUploadView(w http.ResponseWriter, r *http.Request) {
	creatorID := h.getCreatorIDFromSession(r)
	if creatorID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Title": "Upload de Arquivo",
	}

	h.templateRenderer.View(w, r, "file/upload", data, "admin")
}

// FileUploadSubmit processa o upload de arquivo
func (h *FileHandler) FileUploadSubmit(w http.ResponseWriter, r *http.Request) {
	creatorID := h.getCreatorIDFromSession(r)
	if creatorID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse multipart form (máximo 50MB)
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		h.sessionManager.AddFlash(w, r, "Erro ao processar formulário", "error")
		http.Redirect(w, r, "/file/upload", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.sessionManager.AddFlash(w, r, "Arquivo não encontrado", "error")
		http.Redirect(w, r, "/file/upload", http.StatusSeeOther)
		return
	}
	defer file.Close()

	description := r.FormValue("description")

	_, err = h.fileService.UploadFile(header, description, creatorID)
	if err != nil {
		h.sessionManager.AddFlash(w, r, "Erro ao fazer upload: "+err.Error(), "error")
		http.Redirect(w, r, "/file/upload", http.StatusSeeOther)
		return
	}

	// Redirecionar com mensagem de sucesso
	h.sessionManager.AddFlash(w, r, "Arquivo enviado com sucesso!", "success")
	http.Redirect(w, r, "/file", http.StatusSeeOther)
}

// FileDeleteSubmit deleta um arquivo
func (h *FileHandler) FileDeleteSubmit(w http.ResponseWriter, r *http.Request) {
	creatorID := h.getCreatorIDFromSession(r)
	if creatorID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fileIDStr := chi.URLParam(r, "id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		h.sessionManager.AddFlash(w, r, "ID de arquivo inválido", "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	// Verificar se o arquivo pertence ao criador antes de deletar
	file, err := h.fileService.GetFileByID(uint(fileID))
	if err != nil {
		h.sessionManager.AddFlash(w, r, "Arquivo não encontrado", "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	if file.CreatorID != creatorID {
		h.sessionManager.AddFlash(w, r, "Acesso negado", "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	err = h.fileService.DeleteFile(uint(fileID))
	if err != nil {
		h.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	h.sessionManager.AddFlash(w, r, "Arquivo deletado com sucesso!", "success")
	http.Redirect(w, r, "/file", http.StatusSeeOther)
}

// FileUpdateSubmit atualiza nome e descrição do arquivo
func (h *FileHandler) FileUpdateSubmit(w http.ResponseWriter, r *http.Request) {
	creatorID := h.getCreatorIDFromSession(r)
	if creatorID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fileIDStr := chi.URLParam(r, "id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		h.sessionManager.AddFlash(w, r, "ID de arquivo inválido", "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	// Verificar se o arquivo pertence ao criador antes de atualizar
	file, err := h.fileService.GetFileByID(uint(fileID))
	if err != nil {
		h.sessionManager.AddFlash(w, r, "Arquivo não encontrado", "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	if file.CreatorID != creatorID {
		h.sessionManager.AddFlash(w, r, "Acesso negado", "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	// Validar se o nome não está vazio
	if name == "" {
		h.sessionManager.AddFlash(w, r, "Nome do arquivo é obrigatório", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	err = h.fileService.UpdateFile(uint(fileID), name, description)
	if err != nil {
		h.sessionManager.AddFlash(w, r, "Erro ao atualizar arquivo", "error")
		http.Redirect(w, r, "/file", http.StatusSeeOther)
		return
	}

	h.sessionManager.AddFlash(w, r, "Arquivo atualizado com sucesso!", "success")
	http.Redirect(w, r, "/file", http.StatusSeeOther)
}

// getCreatorIDFromSession extrai o ID do criador da sessão usando o SessionService injetado
func (h *FileHandler) getCreatorIDFromSession(r *http.Request) uint {
	// Obter usuário da sessão usando o middleware Auth
	user := middleware.Auth(r)
	if user == nil || user.ID == 0 {
		log.Printf("Usuário não encontrado na sessão")
		return 0
	}

	log.Printf("Usuário encontrado: ID=%d, Email=%s", user.ID, user.Email)

	// Buscar o creator associado ao usuário
	creatorRepository := gorm.NewCreatorRepository(database.DB)
	creator, err := creatorRepository.FindCreatorByUserID(user.ID)
	if err != nil || creator == nil {
		log.Printf("Erro ao buscar creator para usuário %d: %v", user.ID, err)
		return 0
	}

	log.Printf("Creator encontrado: ID=%d, Nome=%s", creator.ID, creator.Name)
	return creator.ID
}
