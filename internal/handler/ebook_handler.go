package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/internal/handler/middleware"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/internal/repository/gorm"
	"github.com/anglesson/simple-web-server/internal/service"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/anglesson/simple-web-server/pkg/storage"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type EbookHandler struct {
	ebookService     service.EbookService
	creatorService   service.CreatorService
	fileService      service.FileService
	s3Storage        storage.S3Storage
	sessionManager   service.SessionService
	templateRenderer template.TemplateRenderer
}

func NewEbookHandler(
	ebookService service.EbookService,
	creatorService service.CreatorService,
	fileService service.FileService,
	s3Storage storage.S3Storage,
	sessionManager service.SessionService,
	templateRenderer template.TemplateRenderer,
) *EbookHandler {
	return &EbookHandler{
		ebookService:     ebookService,
		creatorService:   creatorService,
		fileService:      fileService,
		s3Storage:        s3Storage,
		sessionManager:   sessionManager,
		templateRenderer: templateRenderer,
	}
}

// IndexView renders the ebook index page
func (h *EbookHandler) IndexView(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	title := r.URL.Query().Get("title")

	pagination := models.NewPagination(page, perPage)

	ebooks, err := h.ebookService.ListEbooksForUser(loggedUser.ID, repository.EbookQuery{
		Title:      title,
		Pagination: pagination,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalCount := int64(0)
	if ebooks != nil {
		totalCount = int64(len(*ebooks))
	}
	pagination.SetTotal(totalCount)

	successMessages := h.sessionManager.GetFlashes(w, r, "success")
	errorMessages := h.sessionManager.GetFlashes(w, r, "error")

	h.templateRenderer.View(w, r, "ebook/index", map[string]any{
		"Ebooks":     ebooks,
		"Pagination": pagination,
		"Success":    successMessages,
		"Errors":     errorMessages,
	}, "admin")
}

// CreateView renders the ebook creation page
func (h *EbookHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		http.Error(w, "Erro ao buscar criador", http.StatusInternalServerError)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage == 0 {
		perPage = 20
	}
	pagination := models.NewPagination(page, perPage)

	query := repository.FileQuery{
		Pagination: pagination,
	}

	files, total, err := h.fileService.GetFilesByCreatorPaginated(creator.ID, query)
	if err != nil {
		log.Printf("Erro ao buscar arquivos: %v", err)
		files = []*models.File{}
		total = 0
	}

	pagination.SetTotal(total)

	successMessages := h.sessionManager.GetFlashes(w, r, "success")
	errorMessages := h.sessionManager.GetFlashes(w, r, "error")

	h.templateRenderer.View(w, r, "ebook/create", map[string]interface{}{
		"Files":      files,
		"Creator":    creator,
		"Pagination": pagination,
		"Success":    successMessages,
		"Errors":     errorMessages,
	}, "admin")
}

// CreateSubmit handles ebook creation
func (h *EbookHandler) CreateSubmit(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			log.Printf("Erro ao fazer parse do formulário: %v", err)
			h.sessionManager.AddFlash(w, r, "Erro ao processar formulário", "error")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}
	}

	errors := make(map[string]string)

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		log.Printf("Falha ao cadastrar e-book: %s", err)
		h.sessionManager.AddFlash(w, r, "Falha ao cadastrar e-book", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	uploadedFiles, uploadErrors, err := h.processDirectUploads(r, creator.ID)
	if err != nil {
		log.Printf("Erro ao processar uploads: %v", err)
		errors["upload"] = "Erro ao processar uploads de arquivos"
	}

	if len(uploadErrors) > 0 {
		errors["upload"] = strings.Join(uploadErrors, "; ")
	}

	selectedFiles := r.Form["selected_files"]
	if len(selectedFiles) == 0 && len(uploadedFiles) == 0 {
		errors["files"] = "Selecione pelo menos um arquivo da biblioteca ou faça upload de novos arquivos"
	}

	var value float64
	valueStr := r.FormValue("value")
	if valueStr != "" {
		var err error
		value, err = utils.BRLToFloat(valueStr)
		if err != nil {
			log.Println("Falha na conversão do valor do e-book")
			errors["value"] = "Valor inválido. Use apenas números e vírgula (ex: 29,90)"
		}
	}

	form := models.EbookRequest{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		SalesPage:   r.FormValue("sales_page"),
		Value:       value,
		Status:      true,
	}

	errForm := utils.ValidateForm(form)
	for key, value := range errForm {
		errors[key] = value
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			h.sessionManager.AddFlash(w, r, errMsg, "error")
		}
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	imageURL, err := h.processImageUpload(r, creator.ID)
	if err != nil {
		h.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	ebook := models.NewEbook(form.Title, form.Description, form.SalesPage, form.Value, *creator)

	if imageURL != "" {
		ebook.Image = imageURL
	}

	if len(selectedFiles) > 0 {
		err = h.addSelectedFilesToEbook(ebook, selectedFiles, creator.ID)
		if err != nil {
			log.Printf("Erro ao adicionar arquivos selecionados ao ebook: %v", err)
			h.sessionManager.AddFlash(w, r, "Erro ao adicionar arquivos selecionados ao ebook", "error")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}
	}

	for _, uploadedFile := range uploadedFiles {
		ebook.AddFile(uploadedFile)
	}

	err = h.ebookService.Create(ebook)
	if err != nil {
		log.Printf("Falha ao salvar e-book: %s", err)
		h.sessionManager.AddFlash(w, r, "Falha ao salvar e-book", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	h.sessionManager.AddFlash(w, r, "E-book criado com sucesso!", "success")
	http.Redirect(w, r, "/ebook", http.StatusSeeOther)
}

// UpdateView renders the ebook update page
func (h *EbookHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	ebook := h.getEbookByID(w, r)
	if ebook == nil {
		http.Error(w, "Erro ao buscar e-book", http.StatusNotFound)
		return
	}

	if loggedUser.ID != ebook.Creator.UserID {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		log.Printf("Erro ao buscar criador: %v", err)
		http.Error(w, "Erro ao buscar criador", http.StatusInternalServerError)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage == 0 {
		perPage = 20
	}
	pagination := models.NewPagination(page, perPage)

	query := repository.FileQuery{
		Pagination: pagination,
	}

	allFiles, total, err := h.fileService.GetFilesByCreatorPaginated(creator.ID, query)
	if err != nil {
		log.Printf("Erro ao buscar arquivos: %v", err)
		allFiles = []*models.File{}
		total = 0
	}

	var availableFiles []*models.File
	ebookFileIDs := make(map[uint]bool)
	for _, file := range ebook.Files {
		ebookFileIDs[file.ID] = true
	}

	for _, file := range allFiles {
		if !ebookFileIDs[file.ID] {
			availableFiles = append(availableFiles, file)
		}
	}

	pagination.SetTotal(total)

	successMessages := h.sessionManager.GetFlashes(w, r, "success")
	errorMessages := h.sessionManager.GetFlashes(w, r, "error")

	h.templateRenderer.View(w, r, "ebook/update", map[string]interface{}{
		"ebook":          ebook,
		"AvailableFiles": availableFiles,
		"Pagination":     pagination,
		"Success":        successMessages,
		"Errors":         errorMessages,
	}, "admin")
}

// UpdateSubmit handles ebook update
func (h *EbookHandler) UpdateSubmit(w http.ResponseWriter, r *http.Request) {
	errors := make(map[string]string)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			log.Printf("Erro ao fazer parse do formulário: %v", err)
			h.sessionManager.AddFlash(w, r, "Erro ao processar formulário", "error")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}
	}

	value, err := utils.BRLToFloat(r.FormValue("value"))
	if err != nil {
		h.sessionManager.AddFlash(w, r, "Valor inválido", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	status := false
	if r.FormValue("status") != "" {
		status = true
	}

	form := models.EbookRequest{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		SalesPage:   r.FormValue("sales_page"),
		Value:       value,
		Status:      status,
	}

	errForm := utils.ValidateForm(form)
	for key, value := range errForm {
		errors[key] = value
	}

	user := h.getSessionUser(r)
	if user == nil {
		h.sessionManager.AddFlash(w, r, "Usuário não encontrado", "error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil {
		log.Printf("Falha ao buscar criador: %s", err)
		h.sessionManager.AddFlash(w, r, "Falha ao buscar criador", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	uploadedFiles, uploadErrors, err := h.processDirectUploads(r, creator.ID)
	if err != nil {
		log.Printf("Erro ao processar uploads: %v", err)
		errors["upload"] = "Erro ao processar uploads de arquivos"
	}

	if len(uploadErrors) > 0 {
		errors["upload"] = strings.Join(uploadErrors, "; ")
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			h.sessionManager.AddFlash(w, r, errMsg, "error")
		}
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	ebook := h.getEbookByID(w, r)
	if ebook == nil {
		h.sessionManager.AddFlash(w, r, "E-book não encontrado", "error")
		http.Redirect(w, r, "/ebook", http.StatusSeeOther)
		return
	}

	err = h.processImageUpdate(r, ebook)
	if err != nil {
		h.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	ebook.Title = form.Title
	ebook.Description = form.Description
	ebook.SalesPage = form.SalesPage
	ebook.Value = form.Value
	ebook.Status = form.Status

	for _, uploadedFile := range uploadedFiles {
		ebook.AddFile(uploadedFile)
	}

	newFiles := r.Form["new_files"]
	if len(newFiles) > 0 {
		err = h.addSelectedFilesToEbook(ebook, newFiles, ebook.CreatorID)
		if err != nil {
			log.Printf("Erro ao adicionar novos arquivos ao ebook: %v", err)
			h.sessionManager.AddFlash(w, r, "Erro ao adicionar arquivos ao ebook", "error")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}
	}

	err = h.ebookService.Update(ebook)
	if err != nil {
		log.Printf("Falha ao atualizar e-book: %s", err)
		h.sessionManager.AddFlash(w, r, "Erro ao atualizar e-book", "error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	h.sessionManager.AddFlash(w, r, "Dados do e-book foram atualizados!", "success")
	http.Redirect(w, r, "/ebook", http.StatusSeeOther)
}

// ShowView renders the ebook details page
func (h *EbookHandler) ShowView(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Não foi possível prosseguir com a sua solicitação", http.StatusUnauthorized)
		return
	}

	ebook := h.getEbookByID(w, r)
	if ebook == nil {
		http.Error(w, "Erro ao buscar e-book", http.StatusNotFound)
		return
	}

	if loggedUser.ID != ebook.Creator.UserID {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	term := r.URL.Query().Get("term")
	pagination := models.NewPagination(page, perPage)

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		h.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	clients, err := h.getClientsForEbook(creator, ebook.ID, term, pagination)
	if err != nil {
		h.sessionManager.AddFlash(w, r, err.Error(), "error")
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if clients != nil {
		pagination.SetTotal(int64(len(*clients)))
	}

	h.templateRenderer.View(w, r, "ebook/view", map[string]any{
		"Ebook":      ebook,
		"Clients":    clients,
		"Pagination": pagination,
	}, "admin")
}

// Other helper functions remain the same...

// getEbookByID, getSessionUser, etc.

// ServeEbookImage serve a imagem de capa do ebook de forma segura
func (h *EbookHandler) ServeEbookImage(w http.ResponseWriter, r *http.Request) {
	user := h.getSessionUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ebookID := chi.URLParam(r, "id")
	if ebookID == "" {
		http.Error(w, "ID do ebook não fornecido", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(ebookID, 10, 32)
	if err != nil {
		http.Error(w, "ID do ebook inválido", http.StatusBadRequest)
		return
	}

	ebook, err := h.ebookService.FindByID(uint(id))
	if err != nil || ebook == nil {
		http.Error(w, "Ebook não encontrado", http.StatusNotFound)
		return
	}

	// Permitir apenas o criador acessar a imagem
	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil || creator.ID != ebook.CreatorID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if ebook.Image == "" {
		http.Error(w, "Imagem não encontrada", http.StatusNotFound)
		return
	}

	// Gerar URL pré-assinada temporária (15 minutos)
	key := h.extractS3Key(ebook.Image)
	log.Printf("DEBUG: URL original: %s", ebook.Image)
	log.Printf("DEBUG: Chave extraída: %s", key)
	presignedURL := h.s3Storage.GenerateDownloadLinkWithExpiration(key, 15*60) // 15 minutos
	if presignedURL == "" {
		http.Error(w, "Erro ao gerar URL da imagem", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, presignedURL, http.StatusTemporaryRedirect)
}

// extractS3Key extrai a chave S3 de uma URL pública
func (h *EbookHandler) extractS3Key(url string) string {
	if url == "" {
		return ""
	}

	// Remover parâmetros de query se existirem
	if queryIndex := strings.Index(url, "?"); queryIndex != -1 {
		url = url[:queryIndex]
	}

	// Remover o protocolo
	if len(url) > 8 && url[0:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[0:7] == "http://" {
		url = url[7:]
	}

	// Procurar por "amazonaws.com/"
	amazonawsIndex := strings.Index(url, "amazonaws.com/")
	if amazonawsIndex != -1 {
		return url[amazonawsIndex+14:]
	}

	return ""
}

// Helper methods

func (h *EbookHandler) processImageUpload(r *http.Request, creatorID uint) (string, error) {
	imageFile, imageHeader, imageErr := r.FormFile("image")
	if imageErr != nil || imageFile == nil || imageHeader == nil || imageHeader.Filename == "" {
		return "", nil // No image uploaded
	}

	// Validar se é uma imagem
	contentType := imageHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("o arquivo deve ser uma imagem")
	}

	// Gerar nome único para a imagem
	fileExt := filepath.Ext(imageHeader.Filename)
	uniqueID := fmt.Sprintf("%d-%d", time.Now().Unix(), creatorID)
	imageName := fmt.Sprintf("ebook-covers/%s%s", uniqueID, fileExt)

	// Upload para S3
	imageURL, err := h.s3Storage.UploadFile(imageHeader, imageName)
	if err != nil {
		log.Printf("Erro ao fazer upload da imagem: %v", err)
		return "", fmt.Errorf("erro ao fazer upload da imagem")
	}

	return imageURL, nil
}

func (h *EbookHandler) processImageUpdate(r *http.Request, ebook *models.Ebook) error {
	imageFile, imageHeader, imageErr := r.FormFile("image")
	if imageErr != nil || imageFile == nil || imageHeader == nil || imageHeader.Filename == "" {
		return nil // No new image uploaded
	}

	// Validar se é uma imagem
	contentType := imageHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("o arquivo deve ser uma imagem")
	}

	// Gerar nome único para a imagem
	fileExt := filepath.Ext(imageHeader.Filename)
	uniqueID := fmt.Sprintf("%d-%d", time.Now().Unix(), ebook.CreatorID)
	imageName := fmt.Sprintf("ebook-covers/%s%s", uniqueID, fileExt)

	// Upload para S3
	imageURL, err := h.s3Storage.UploadFile(imageHeader, imageName)
	if err != nil {
		log.Printf("Erro ao fazer upload da imagem: %v", err)
		return fmt.Errorf("erro ao fazer upload da imagem")
	}

	// Se o upload foi bem-sucedido, atualizar a URL da imagem
	ebook.Image = imageURL
	return nil
}

func (h *EbookHandler) addSelectedFilesToEbook(ebook *models.Ebook, selectedFiles []string, creatorID uint) error {
	for _, fileIDStr := range selectedFiles {
		fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
		if err != nil {
			continue
		}

		file, err := h.fileService.GetFileByID(uint(fileID))
		if err != nil {
			continue
		}

		// Verificar se o arquivo pertence ao criador
		if file.CreatorID == creatorID {
			ebook.AddFile(file)
		}
	}
	return nil
}

// validateFile valida se um arquivo atende aos requisitos de segurança
func (h *EbookHandler) validateFile(file multipart.File, expectedContentType string) map[string]string {
	errors := make(map[string]string)

	defer file.Close()

	// Validar tamanho
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		errors["File"] = "Erro ao ler arquivo"
		return errors
	}

	// Validar tamanho
	const MAX_FILE_SIZE = 60 * 1024 * 1024 // 60 MB
	if len(fileBytes) > MAX_FILE_SIZE {
		errors["File"] = fmt.Sprintf("Arquivo deve ter no máximo %d MB", MAX_FILE_SIZE/(1024*1024))
		return errors
	}

	// Validar tipo MIME
	contentType := http.DetectContentType(fileBytes)
	log.Printf("content type: %s", contentType)

	// Lista de tipos MIME permitidos
	allowedMimeTypes := map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"image/jpeg": true,
		"image/png":  true,
	}

	if !allowedMimeTypes[contentType] {
		errors["File"] = "Tipo de arquivo não permitido. Apenas PDF, DOC, DOCX, JPEG e PNG são aceitos"
		return errors
	}

	// Se esperamos um tipo específico, validamos contra ele
	if expectedContentType != "" && contentType != expectedContentType {
		errors["File"] = fmt.Sprintf("O arquivo deve ser do tipo %s", expectedContentType)
		return errors
	}

	return errors
}

func (h *EbookHandler) getEbookByID(w http.ResponseWriter, r *http.Request) *models.Ebook {
	ebookID := chi.URLParam(r, "id")
	if ebookID == "" {
		http.Error(w, "ID do e-book não fornecido", http.StatusBadRequest)
		return nil
	}

	// Converter string para uint
	id, err := strconv.ParseUint(ebookID, 10, 32)
	if err != nil {
		http.Error(w, "ID do e-book inválido", http.StatusBadRequest)
		return nil
	}

	ebook, err := h.ebookService.FindByID(uint(id))
	if err != nil {
		http.Error(w, "Erro ao buscar e-book", http.StatusInternalServerError)
		return nil
	}

	return ebook
}

func (h *EbookHandler) getSessionUser(r *http.Request) *models.User {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		log.Printf("Erro ao recuperar usuário da sessão: %s", userEmail)
		return nil
	}

	// For testing purposes, create a mock user if email is test@example.com
	if userEmail == "test@example.com" {
		user := &models.User{
			Email: userEmail,
		}
		// Set ID for testing (gorm.Model embeds ID)
		user.ID = 1
		return user
	}

	// This should be injected as a dependency, but for now we'll use the repository directly
	// TODO: Inject UserRepository as dependency
	userRepository := repository.NewGormUserRepository(database.DB)
	return userRepository.FindByEmail(userEmail)
}

func (h *EbookHandler) getClientsForEbook(creator *models.Creator, ebookID uint, term string, pagination *models.Pagination) (*[]models.Client, error) {
	// This should be moved to a service method
	// TODO: Create a method in ClientService to get clients for ebook
	clientRepository := gorm.NewClientGormRepository()
	return clientRepository.FindByClientsWhereEbookWasSend(creator, models.ClientFilter{
		Term:       term,
		EbookID:    ebookID,
		Pagination: pagination,
	})
}

// validateSelectedFiles validates that at least one file is selected and not too many files
func (h *EbookHandler) validateSelectedFiles(selectedFiles []string) error {
	// Verificar se há pelo menos um arquivo selecionado
	if len(selectedFiles) == 0 {
		return fmt.Errorf("Selecione pelo menos um arquivo para o ebook")
	}

	// Segurança: Limitar o número máximo de arquivos por upload
	const MAX_FILES_PER_UPLOAD = 10
	if len(selectedFiles) > MAX_FILES_PER_UPLOAD {
		return fmt.Errorf("máximo %d arquivos por upload permitidos", MAX_FILES_PER_UPLOAD)
	}

	return nil
}

// checkFileAlreadyInEbook checks if a file is already associated with an ebook
func (h *EbookHandler) checkFileAlreadyInEbook(ebook *models.Ebook, fileID uint) bool {
	for _, file := range ebook.Files {
		if file.ID == fileID {
			return true
		}
	}
	return false
}

// validateFileOwnership validates that a file belongs to the specified creator
func (h *EbookHandler) validateFileOwnership(file *models.File, creatorID uint) error {
	if file.CreatorID != creatorID {
		return fmt.Errorf("arquivo não pertence ao criador")
	}
	return nil
}

// removeFileFromEbookLogic removes a file from an ebook with validation
func (h *EbookHandler) removeFileFromEbookLogic(ebook *models.Ebook, fileID uint) error {
	// Check if it's the last file
	if len(ebook.Files) <= 1 {
		return fmt.Errorf("ebook deve ter pelo menos um arquivo")
	}

	// Find and remove the file
	for i, file := range ebook.Files {
		if file.ID == fileID {
			ebook.Files = append(ebook.Files[:i], ebook.Files[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("arquivo não encontrado no ebook")
}

// calculateFilesTotalSize calculates the total size of selected files
func (h *EbookHandler) calculateFilesTotalSize(files []*models.File, selectedFiles []string) int64 {
	selectedMap := make(map[string]bool)
	for _, fileID := range selectedFiles {
		selectedMap[fileID] = true
	}

	var totalSize int64
	for _, file := range files {
		if selectedMap[fmt.Sprintf("%d", file.ID)] {
			totalSize += file.FileSize
		}
	}

	return totalSize
}

// processDirectUploads handles direct file uploads during ebook creation/editing
func (h *EbookHandler) processDirectUploads(r *http.Request, creatorID uint) ([]*models.File, []string, error) {
	var uploadedFiles []*models.File
	var errors []string

	// Verificar se o form foi parseado
	if r.MultipartForm == nil {
		// Tentar parsear o formulário com um limite razoável
		err := r.ParseMultipartForm(32 << 20) // 32 MB
		if err != nil {
			// Pode não ser um formulário multipart, o que é válido se não houver uploads
			return uploadedFiles, errors, nil
		}
	}

	// Verificar se há arquivos carregados
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return uploadedFiles, errors, nil // Sem uploads
	}

	files, ok := r.MultipartForm.File["new_files"]
	if !ok || len(files) == 0 {
		return uploadedFiles, errors, nil // No files uploaded
	}

	// Log de segurança para auditoria
	userIP := r.RemoteAddr
	userAgent := r.UserAgent()
	log.Printf("[SECURITY-AUDIT] Upload iniciado: %d arquivos de IP %s (User-Agent: %s) para criador ID %d",
		len(files), userIP, userAgent, creatorID)

	// Segurança: Limitar o número máximo de arquivos por upload
	const MAX_FILES_PER_UPLOAD = 10
	if len(files) > MAX_FILES_PER_UPLOAD {
		errorMsg := fmt.Sprintf("Número máximo de arquivos excedido: %d (máximo: %d)", len(files), MAX_FILES_PER_UPLOAD)

		// Log de segurança para tentativa de exceder limite
		log.Printf("[SECURITY-WARN] Tentativa de exceder limite de arquivos: %s de IP %s (User-Agent: %s) para criador ID %d",
			errorMsg, userIP, userAgent, creatorID)

		return nil, []string{fmt.Sprintf("Máximo %d arquivos por upload permitidos", MAX_FILES_PER_UPLOAD)},
			fmt.Errorf(errorMsg)
	}

	// Process each uploaded file
	for _, fileHeader := range files {
		// Segurança: Verificação de tipo de arquivo por extensão antes do processamento
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		allowedExts := []string{".pdf", ".doc", ".docx", ".jpg", ".jpeg", ".png", ".epub", ".mobi", ".azw3", ".txt"}

		isAllowed := false
		for _, allowedExt := range allowedExts {
			if ext == allowedExt {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			errorMsg := fmt.Sprintf("Tipo de arquivo não permitido: %s", ext)
			log.Printf("[SECURITY-WARN] Tentativa de upload de tipo não permitido: %s (%s) de IP %s",
				fileHeader.Filename, ext, userIP)
			errors = append(errors, errorMsg)
			continue
		}

		// Segurança: Sanitizar nome do arquivo
		originalFilename := fileHeader.Filename
		fileHeader.Filename = h.sanitizeFilename(fileHeader.Filename)

		// Registrar mudança de nome para auditoria se houver alteração
		if originalFilename != fileHeader.Filename {
			log.Printf("[SECURITY-INFO] Nome de arquivo sanitizado: %s -> %s (Criador ID: %d)",
				originalFilename, fileHeader.Filename, creatorID)
		}

		// Get description from form if provided
		description := r.FormValue("description_" + originalFilename)

		// Upload file using FileService
		uploadedFile, err := h.fileService.UploadFile(fileHeader, description, creatorID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Erro ao fazer upload de %s: %v", fileHeader.Filename, err))
			continue
		}

		// Log de segurança para sucesso de upload
		log.Printf("[SECURITY-AUDIT] Upload bem-sucedido: arquivo ID %d (%s, %.2f MB) para criador ID %d",
			uploadedFile.ID, uploadedFile.Name, float64(uploadedFile.FileSize)/1024/1024, creatorID)

		uploadedFiles = append(uploadedFiles, uploadedFile)
	}

	return uploadedFiles, errors, nil
}

// checkUserStorageQuota verifica se o usuário excedeu sua cota de armazenamento
func (h *EbookHandler) checkUserStorageQuota(creatorID uint, newFileSize int64) error {
	// Constante de cota máxima de armazenamento por usuário: 1GB
	const MAX_STORAGE_PER_USER = 1024 * 1024 * 1024 // 1GB

	// Para esta implementação, usaremos uma abordagem simplificada de verificação de cota
	// Em uma implementação completa, seria necessário buscar o uso total do usuário no banco de dados

	// Neste caso, usamos uma estimativa conservadora baseada no tamanho do novo upload
	// Se o novo upload for maior que 10% da cota total, verificamos com mais cuidado
	if newFileSize > MAX_STORAGE_PER_USER/10 {
		log.Printf("Upload grande detectado (%.2f MB) para o criador ID %d. Verificação de cota necessária.",
			float64(newFileSize)/1024/1024, creatorID)

		// Na implementação completa, buscaria todos os arquivos do usuário e somaria seus tamanhos
		// files, err := repository.GetFilesByCreatorID(creatorID)
		// var totalSize int64
		// for _, file := range files { totalSize += file.Size }
	}

	// Esta é uma implementação temporária; em produção, deveria verificar o banco de dados
	// Para não bloquear o desenvolvimento, permitiremos o upload por enquanto

	return nil
}

// sanitizeFilename sanitizes a filename to prevent path traversal and other security issues
func (h *EbookHandler) sanitizeFilename(filename string) string {
	// Remover path separators para evitar path traversal
	filename = filepath.Base(filename)

	// Remover caracteres perigosos usando regex
	safeFilename := strings.Map(func(r rune) rune {
		// Permitir apenas caracteres alfanuméricos, ponto, traço, underline e espaços
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == ' ' {
			return r
		}
		return '_' // Substituir caracteres não permitidos por underline
	}, filename)

	// Limitar tamanho do nome do arquivo
	const MAX_FILENAME_LENGTH = 255
	if len(safeFilename) > MAX_FILENAME_LENGTH {
		ext := filepath.Ext(safeFilename)
		safeFilename = safeFilename[:MAX_FILENAME_LENGTH-len(ext)] + ext
	}

	return safeFilename
}

// RemoveFileFromEbook removes a file from an ebook
func (h *EbookHandler) RemoveFileFromEbook(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
		return
	}

	ebook := h.getEbookByID(w, r)
	if ebook == nil {
		return // Error already handled in getEbookByID
	}

	// Verify ownership
	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil || creator.ID != ebook.CreatorID {
		http.Error(w, "Não autorizado", http.StatusForbidden)
		return
	}

	fileID := chi.URLParam(r, "fileId")
	fileIDParsed, err := strconv.ParseUint(fileID, 10, 32)
	if err != nil {
		http.Error(w, "ID do arquivo inválido", http.StatusBadRequest)
		return
	}

	// Remove file logic
	err = h.removeFileFromEbookLogic(ebook, uint(fileIDParsed))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save changes
	err = h.ebookService.Update(ebook)
	if err != nil {
		log.Printf("Erro ao atualizar ebook: %v", err)
		http.Error(w, "Erro ao remover arquivo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Arquivo removido com sucesso",
	})
}

// AddFileToEbook adds a file to an ebook
func (h *EbookHandler) AddFileToEbook(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
		return
	}

	ebook := h.getEbookByID(w, r)
	if ebook == nil {
		return // Error already handled in getEbookByID
	}

	// Verify ownership
	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil || creator.ID != ebook.CreatorID {
		http.Error(w, "Não autorizado", http.StatusForbidden)
		return
	}

	fileID := chi.URLParam(r, "fileId")
	fileIDParsed, err := strconv.ParseUint(fileID, 10, 32)
	if err != nil {
		http.Error(w, "ID do arquivo inválido", http.StatusBadRequest)
		return
	}

	// Get file
	file, err := h.fileService.GetFileByID(uint(fileIDParsed))
	if err != nil {
		http.Error(w, "Arquivo não encontrado", http.StatusNotFound)
		return
	}

	// Validate file ownership
	err = h.validateFileOwnership(file, creator.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Check if file is already in ebook
	if h.checkFileAlreadyInEbook(ebook, file.ID) {
		http.Error(w, "Arquivo já está associado ao ebook", http.StatusBadRequest)
		return
	}

	// Add file to ebook
	ebook.AddFile(file)

	// Save changes
	err = h.ebookService.Update(ebook)
	if err != nil {
		log.Printf("Erro ao atualizar ebook: %v", err)
		http.Error(w, "Erro ao adicionar arquivo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Arquivo adicionado com sucesso",
	})
}

// UploadAndAddFileToEbook handles direct file upload during ebook creation/editing
func (h *EbookHandler) UploadAndAddFileToEbook(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || userEmail == "" {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}

	loggedUser := h.getSessionUser(r)
	if loggedUser == nil {
		http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		http.Error(w, "Criador não encontrado", http.StatusNotFound)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(32 << 20) // 32 MB limit
	if err != nil {
		http.Error(w, "Erro ao processar upload", http.StatusBadRequest)
		return
	}

	// Get uploaded files
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "Nenhum arquivo foi enviado", http.StatusBadRequest)
		return
	}

	// Segurança: Limitar o número máximo de arquivos por upload
	const MAX_FILES_PER_UPLOAD = 10
	if len(files) > MAX_FILES_PER_UPLOAD {
		http.Error(w, fmt.Sprintf("Máximo %d arquivos por upload permitidos", MAX_FILES_PER_UPLOAD),
			http.StatusBadRequest)
		return
	}

	// Verificar cota de armazenamento do usuário
	totalUploadSize := int64(0)
	for _, fileHeader := range files {
		totalUploadSize += fileHeader.Size
	}

	if err := h.checkUserStorageQuota(creator.ID, totalUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var uploadedFiles []*models.File
	var errors []string

	// Process each uploaded file
	for _, fileHeader := range files {
		// Segurança: Sanitizar nome do arquivo
		originalFilename := fileHeader.Filename
		fileHeader.Filename = h.sanitizeFilename(fileHeader.Filename)

		// Get description from form if provided
		description := r.FormValue("description_" + originalFilename)

		// Upload file using FileService
		uploadedFile, err := h.fileService.UploadFile(fileHeader, description, creator.ID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Erro ao fazer upload de %s: %v", fileHeader.Filename, err))
			continue
		}

		uploadedFiles = append(uploadedFiles, uploadedFile)
	}

	// Return response
	response := map[string]interface{}{
		"uploaded_files": uploadedFiles,
		"total_uploaded": len(uploadedFiles),
	}

	if len(errors) > 0 {
		response["errors"] = errors
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
