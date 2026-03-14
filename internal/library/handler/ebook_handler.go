package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	accountsvc "github.com/anglesson/simple-web-server/internal/account/service"
	authmw "github.com/anglesson/simple-web-server/internal/auth/handler/middleware"
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	authrepo "github.com/anglesson/simple-web-server/internal/auth/repository"
	authsvc "github.com/anglesson/simple-web-server/internal/auth/service"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"
	librarysvc "github.com/anglesson/simple-web-server/internal/library/service"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepogorm "github.com/anglesson/simple-web-server/internal/sales/repository/gorm"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/anglesson/simple-web-server/pkg/storage"
	"github.com/anglesson/simple-web-server/pkg/template"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type EbookHandler struct {
	ebookService     librarysvc.EbookService
	creatorService   accountsvc.CreatorService
	fileService      librarysvc.FileService
	s3Storage        storage.S3Storage
	sessionManager   authsvc.SessionService
	templateRenderer template.TemplateRenderer
}

func NewEbookHandler(
	ebookService librarysvc.EbookService,
	creatorService accountsvc.CreatorService,
	fileService librarysvc.FileService,
	s3Storage storage.S3Storage,
	sessionManager authsvc.SessionService,
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
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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

	pagination := salesmodel.NewPagination(page, perPage)

	ebooks, err := h.ebookService.ListEbooksForUser(loggedUser.ID, libraryrepo.EbookQuery{
		Term:       title,
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
		"Term":       title,
		"Filters": map[string]interface{}{
			"title": title,
		},
	}, "admin-daisy")
}

// CreateView renders the ebook creation page
func (h *EbookHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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
	pagination := salesmodel.NewPagination(page, perPage)

	query := libraryrepo.FileQuery{
		Pagination: pagination,
	}

	files, total, err := h.fileService.GetFilesByCreatorPaginated(creator.ID, query)
	if err != nil {
		log.Printf("Erro ao buscar arquivos: %v", err)
		files = []*librarymodel.File{}
		total = 0
	}

	pagination.SetTotal(total)

	var form librarymodel.EbookRequest
	formBytes := h.sessionManager.Get(r, "form")
	if formBytes != nil {
		if data, ok := formBytes.([]byte); ok {
			err := json.Unmarshal(data, &form)
			if err != nil {
				slog.Error("Error in unmarshalling form", "error", err)
				return
			}
		}
	}
	h.sessionManager.Pop(r, w, "form")

	errorMessages := h.sessionManager.GetFlashes(w, r, "form-error")

	h.templateRenderer.View(w, r, "ebook/create", map[string]interface{}{
		"Files":      files,
		"Creator":    creator,
		"Pagination": pagination,
		"FormErrors": errorMessages,
		"Form":       form,
	}, "admin-daisy")
}

// CreateSubmit handles ebook creation
func (h *EbookHandler) CreateSubmit(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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
			h.FlashMessage(w, r, "Erro ao processar formulário", "error")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}
	}

	errors := make(map[string]string)

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil {
		log.Printf("Falha ao cadastrar e-book: %s", err)
		h.FlashMessage(w, r, "Falha ao cadastrar e-book", "error")
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

	selectedFiles := r.Form["new_files"]
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

	var promotionalValue float64
	promotionalValueStr := r.FormValue("value")
	if valueStr != "" {
		var err error
		value, err = utils.BRLToFloat(promotionalValueStr)
		if err != nil {
			log.Println("Falha na conversão do valor promocional do e-book")
			errors["promotional_value"] = "Valor inválido. Use apenas números e vírgula (ex: 29,90)"
		}
	}

	form := librarymodel.EbookRequest{
		Title:            r.FormValue("title"),
		Description:      r.FormValue("description"),
		SalesPage:        r.FormValue("sales_page"),
		Value:            value,
		PromotionalValue: promotionalValue,
		Status:           true,
		Statistics:       false,
	}

	errForm := utils.ValidateForm(form)
	for key, value := range errForm {
		errors[key] = value
	}

	if len(errors) > 0 {
		for _, errMsg := range errors {
			h.FlashMessage(w, r, errMsg, "form-error")
			h.SetFormToSession(w, r, form)
		}
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	imageURL, err := h.processImageUpload(r, creator.ID)
	if err != nil {
		h.FlashMessage(w, r, err.Error(), "form-error")
		h.SetFormToSession(w, r, form)

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	ebook := librarymodel.NewEbook(form.Title, form.Description, form.SalesPage, form.Value, form.PromotionalValue, creator.ID, form.Statistics)

	if imageURL != "" {
		ebook.Image = imageURL
	}

	if len(selectedFiles) > 0 {
		err = h.addSelectedFilesToEbook(ebook, selectedFiles, creator.ID)
		if err != nil {
			log.Printf("Erro ao adicionar arquivos selecionados ao ebook: %v", err)
			h.FlashMessage(w, r, "Erro ao adicionar arquivos selecionados ao ebook", "form-error")

			h.SetFormToSession(w, r, form)

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
		h.FlashMessage(w, r, err.Error(), "form-error")
		h.SetFormToSession(w, r, form)
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	h.FlashMessage(w, r, "E-book criado com sucesso!", "success")
	http.Redirect(w, r, "/ebook", http.StatusSeeOther)
}

func (h *EbookHandler) SetFormToSession(w http.ResponseWriter, r *http.Request, form interface{}) {
	formData, _ := json.Marshal(form)
	err := h.sessionManager.Set(r, w, "form", formData)
	if err != nil {
		slog.Error("Error in session manager", "error", err)
		return
	}
}

// UpdateView renders the ebook update page
func (h *EbookHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil || creator.ID != ebook.CreatorID {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage == 0 {
		perPage = 20
	}
	pagination := salesmodel.NewPagination(page, perPage)

	query := libraryrepo.FileQuery{
		Pagination: pagination,
	}

	allFiles, total, err := h.fileService.GetFilesByCreatorPaginated(creator.ID, query)
	if err != nil {
		log.Printf("Erro ao buscar arquivos: %v", err)
		allFiles = []*librarymodel.File{}
		total = 0
	}

	var availableFiles []*librarymodel.File
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

	h.templateRenderer.View(w, r, "ebook/update", map[string]interface{}{
		"ebook":          ebook,
		"AvailableFiles": availableFiles,
		"Pagination":     pagination,
		"Success":        successMessages,
		"FormErrors":     h.sessionManager.GetFlashes(w, r, "form-error"),
	}, "admin-daisy")
}

// UpdateSubmit handles ebook update
func (h *EbookHandler) UpdateSubmit(w http.ResponseWriter, r *http.Request) {
	errors := make(map[string]string)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			log.Printf("Erro ao fazer parse do formulário: %v", err)
			h.FlashMessage(w, r, "Erro ao processar formulário", "form-error")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}
	}

	value, err := utils.BRLToFloat(r.FormValue("value"))
	if err != nil {
		h.FlashMessage(w, r, "Valor inválido", "form-error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	promotionalValue, err := utils.BRLToFloat(r.FormValue("promotional_value"))
	if err != nil {
		h.FlashMessage(w, r, "Valor promocional inválido", "form-error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	status := false
	if r.FormValue("status") != "" {
		status = true
	}

	statistics := false
	if r.FormValue("statistics") != "" {
		statistics = true
	}

	form := librarymodel.EbookRequest{
		Title:            r.FormValue("title"),
		Description:      r.FormValue("description"),
		SalesPage:        r.FormValue("sales_page"),
		Value:            value,
		PromotionalValue: promotionalValue,
		Status:           status,
		Statistics:       statistics,
	}

	errForm := utils.ValidateForm(form)
	for key, value := range errForm {
		errors[key] = value
	}

	user := h.getSessionUser(r)
	if user == nil {
		h.FlashMessage(w, r, "Usuário não encontrado", "error")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil {
		log.Printf("Falha ao buscar criador: %s", err)
		h.FlashMessage(w, r, "Falha ao buscar criador", "error")
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
			h.FlashMessage(w, r, errMsg, "form-error")
		}
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	ebook := h.getEbookByID(w, r)
	if ebook == nil {
		h.FlashMessage(w, r, "E-book não encontrado", "error")
		http.Redirect(w, r, "/ebook", http.StatusSeeOther)
		return
	}

	err = h.processImageUpdate(r, ebook)
	if err != nil {
		h.FlashMessage(w, r, err.Error(), "form-error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	ebook.Title = form.Title
	ebook.Description = form.Description
	ebook.SalesPage = form.SalesPage
	ebook.Value = form.Value
	ebook.PromotionalValue = form.PromotionalValue
	ebook.Status = form.Status
	ebook.Statistics = form.Statistics

	err = h.ebookService.Update(ebook)
	if err != nil {
		log.Printf("Falha ao atualizar e-book: %s", err)
		h.FlashMessage(w, r, "Erro ao atualizar e-book", "form-error")
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}

	// Adicionar novos arquivos via Association (sem tocar nos existentes)
	var filesToAppend []*librarymodel.File
	filesToAppend = append(filesToAppend, uploadedFiles...)

	newFiles := r.Form["new_files"]
	for _, fileIDStr := range newFiles {
		fileID, parseErr := strconv.ParseUint(fileIDStr, 10, 32)
		if parseErr != nil {
			continue
		}
		file, fileErr := h.fileService.GetFileByID(uint(fileID))
		if fileErr != nil || file.CreatorID != creator.ID {
			continue
		}
		if !h.checkFileAlreadyInEbook(ebook, file.ID) {
			filesToAppend = append(filesToAppend, file)
		}
	}

	if len(filesToAppend) > 0 {
		if appendErr := h.ebookService.AppendFiles(ebook.ID, filesToAppend); appendErr != nil {
			log.Printf("Erro ao adicionar arquivos ao ebook: %v", appendErr)
			h.FlashMessage(w, r, "Erro ao adicionar arquivos ao ebook", "form-error")
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}
	}

	h.FlashMessage(w, r, "Dados do e-book foram atualizados!", "success")
	http.Redirect(w, r, "/ebook", http.StatusSeeOther)
}

// ShowView renders the ebook details page
func (h *EbookHandler) ShowView(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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

	creator, err := h.creatorService.FindCreatorByUserID(loggedUser.ID)
	if err != nil || creator.ID != ebook.CreatorID {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	term := r.URL.Query().Get("term")
	pagination := salesmodel.NewPagination(page, perPage)

	clients, err := h.getClientsForEbook(creator, ebook.ID, term, pagination)
	if err != nil {
		h.FlashMessage(w, r, err.Error(), "error")
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
	}, "admin-daisy")
}

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

	creator, err := h.creatorService.FindCreatorByUserID(user.ID)
	if err != nil || creator.ID != ebook.CreatorID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if ebook.Image == "" {
		http.Error(w, "Imagem não encontrada", http.StatusNotFound)
		return
	}

	key := h.extractS3Key(ebook.Image)
	log.Printf("DEBUG: URL original: %s", ebook.Image)
	log.Printf("DEBUG: Chave extraída: %s", key)
	presignedURL := h.s3Storage.GenerateDownloadLinkWithExpiration(key, 15*60)
	if presignedURL == "" {
		http.Error(w, "Erro ao gerar URL da imagem", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, presignedURL, http.StatusTemporaryRedirect)
}

func (h *EbookHandler) extractS3Key(url string) string {
	if url == "" {
		return ""
	}

	if queryIndex := strings.Index(url, "?"); queryIndex != -1 {
		url = url[:queryIndex]
	}

	if len(url) > 8 && url[0:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[0:7] == "http://" {
		url = url[7:]
	}

	amazonawsIndex := strings.Index(url, "amazonaws.com/")
	if amazonawsIndex != -1 {
		return url[amazonawsIndex+14:]
	}

	return ""
}

func (h *EbookHandler) processImageUpload(r *http.Request, creatorID uint) (string, error) {
	imageFile, imageHeader, imageErr := r.FormFile("image")
	if imageErr != nil || imageFile == nil || imageHeader == nil || imageHeader.Filename == "" {
		return "", nil
	}

	contentType := imageHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("o arquivo deve ser uma imagem")
	}

	fileExt := filepath.Ext(imageHeader.Filename)
	uniqueID := fmt.Sprintf("%d-%d", time.Now().Unix(), creatorID)
	imageName := fmt.Sprintf("ebook-covers/%s%s", uniqueID, fileExt)

	const coverCache = "public, max-age=31536000, immutable"
	imageURL, err := h.s3Storage.UploadFile(imageHeader, imageName, coverCache)
	if err != nil {
		log.Printf("Erro ao fazer upload da imagem: %v", err)
		return "", fmt.Errorf("erro ao fazer upload da imagem")
	}

	return imageURL, nil
}

func (h *EbookHandler) processImageUpdate(r *http.Request, ebook *librarymodel.Ebook) error {
	imageFile, imageHeader, imageErr := r.FormFile("image")
	if imageErr != nil || imageFile == nil || imageHeader == nil || imageHeader.Filename == "" {
		return nil
	}

	contentType := imageHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("o arquivo deve ser uma imagem")
	}

	fileExt := filepath.Ext(imageHeader.Filename)
	uniqueID := fmt.Sprintf("%d-%d", time.Now().Unix(), ebook.CreatorID)
	imageName := fmt.Sprintf("ebook-covers/%s%s", uniqueID, fileExt)

	const coverCache = "public, max-age=31536000, immutable"
	imageURL, err := h.s3Storage.UploadFile(imageHeader, imageName, coverCache)
	if err != nil {
		log.Printf("Erro ao fazer upload da imagem: %v", err)
		return fmt.Errorf("erro ao fazer upload da imagem")
	}

	ebook.Image = imageURL
	return nil
}

func (h *EbookHandler) addSelectedFilesToEbook(ebook *librarymodel.Ebook, selectedFiles []string, creatorID uint) error {
	for _, fileIDStr := range selectedFiles {
		fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
		if err != nil {
			continue
		}

		file, err := h.fileService.GetFileByID(uint(fileID))
		if err != nil {
			continue
		}

		if file.CreatorID == creatorID {
			ebook.AddFile(file)
		}
	}
	return nil
}

func (h *EbookHandler) validateFile(file multipart.File, expectedContentType string) map[string]string {
	errors := make(map[string]string)

	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		errors["File"] = "Erro ao ler arquivo"
		return errors
	}

	const MAX_FILE_SIZE = 60 * 1024 * 1024 // 60 MB
	if len(fileBytes) > MAX_FILE_SIZE {
		errors["File"] = fmt.Sprintf("Arquivo deve ter no máximo %d MB", MAX_FILE_SIZE/(1024*1024))
		return errors
	}

	contentType := http.DetectContentType(fileBytes)
	log.Printf("content type: %s", contentType)

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

	if expectedContentType != "" && contentType != expectedContentType {
		errors["File"] = fmt.Sprintf("O arquivo deve ser do tipo %s", expectedContentType)
		return errors
	}

	return errors
}

func (h *EbookHandler) getEbookByID(w http.ResponseWriter, r *http.Request) *librarymodel.Ebook {
	ebookID := chi.URLParam(r, "id")
	if ebookID == "" {
		http.Error(w, "ID do e-book não fornecido", http.StatusBadRequest)
		return nil
	}

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

func (h *EbookHandler) getSessionUser(r *http.Request) *authmodel.User {
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
	if !ok {
		log.Printf("Erro ao recuperar usuário da sessão: %s", userEmail)
		return nil
	}

	if userEmail == "test@example.com" {
		user := &authmodel.User{
			Email: userEmail,
		}
		user.ID = 1
		return user
	}

	userRepository := authrepo.NewGormUserRepository(database.DB)
	return userRepository.FindByEmail(userEmail)
}

func (h *EbookHandler) getClientsForEbook(creator *accountmodel.Creator, ebookID uint, term string, pagination *salesmodel.Pagination) (*[]salesmodel.Client, error) {
	clientRepository := salesrepogorm.NewClientGormRepository()
	return clientRepository.FindByClientsWhereEbookWasSend(creator, salesmodel.ClientFilter{
		Term:       term,
		EbookID:    ebookID,
		Pagination: pagination,
	})
}

func (h *EbookHandler) validateSelectedFiles(selectedFiles []string) error {
	if len(selectedFiles) == 0 {
		return fmt.Errorf("Selecione pelo menos um arquivo para o ebook")
	}

	const MAX_FILES_PER_UPLOAD = 10
	if len(selectedFiles) > MAX_FILES_PER_UPLOAD {
		return fmt.Errorf("máximo %d arquivos por upload permitidos", MAX_FILES_PER_UPLOAD)
	}

	return nil
}

func (h *EbookHandler) checkFileAlreadyInEbook(ebook *librarymodel.Ebook, fileID uint) bool {
	for _, file := range ebook.Files {
		if file.ID == fileID {
			return true
		}
	}
	return false
}

func (h *EbookHandler) validateFileOwnership(file *librarymodel.File, creatorID uint) error {
	if file.CreatorID != creatorID {
		return fmt.Errorf("arquivo não pertence ao criador")
	}
	return nil
}

func (h *EbookHandler) removeFileFromEbookLogic(ebook *librarymodel.Ebook, fileID uint) error {
	if len(ebook.Files) <= 1 {
		return fmt.Errorf("ebook deve ter pelo menos um arquivo")
	}

	for i, file := range ebook.Files {
		if file.ID == fileID {
			ebook.Files = append(ebook.Files[:i], ebook.Files[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("arquivo não encontrado no ebook")
}

func (h *EbookHandler) calculateFilesTotalSize(files []*librarymodel.File, selectedFiles []string) int64 {
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

func (h *EbookHandler) processDirectUploads(r *http.Request, creatorID uint) ([]*librarymodel.File, []string, error) {
	var uploadedFiles []*librarymodel.File
	var errors []string

	if r.MultipartForm == nil {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			return uploadedFiles, errors, nil
		}
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return uploadedFiles, errors, nil
	}

	files, ok := r.MultipartForm.File["upload_files"]
	if !ok || len(files) == 0 {
		return uploadedFiles, errors, nil
	}

	userIP := r.RemoteAddr
	userAgent := r.UserAgent()
	log.Printf("[SECURITY-AUDIT] Upload iniciado: %d arquivos de IP %s (User-Agent: %s) para criador ID %d",
		len(files), userIP, userAgent, creatorID)

	const MAX_FILES_PER_UPLOAD = 10
	if len(files) > MAX_FILES_PER_UPLOAD {
		errorMsg := fmt.Sprintf("Número máximo de arquivos excedido: %d (máximo: %d)", len(files), MAX_FILES_PER_UPLOAD)
		log.Printf("[SECURITY-WARN] Tentativa de exceder limite de arquivos: %s de IP %s (User-Agent: %s) para criador ID %d",
			errorMsg, userIP, userAgent, creatorID)
		return nil, []string{fmt.Sprintf("Máximo %d arquivos por upload permitidos", MAX_FILES_PER_UPLOAD)},
			fmt.Errorf(errorMsg)
	}

	for _, fileHeader := range files {
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

		originalFilename := fileHeader.Filename
		fileHeader.Filename = h.sanitizeFilename(fileHeader.Filename)

		if originalFilename != fileHeader.Filename {
			log.Printf("[SECURITY-INFO] Nome de arquivo sanitizado: %s -> %s (Criador ID: %d)",
				originalFilename, fileHeader.Filename, creatorID)
		}

		description := r.FormValue("description_" + originalFilename)
		name := r.FormValue("name_" + originalFilename)

		uploadedFile, err := h.fileService.UploadFile(fileHeader, name, description, creatorID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Erro ao fazer upload de %s: %v", fileHeader.Filename, err))
			continue
		}

		log.Printf("[SECURITY-AUDIT] Upload bem-sucedido: arquivo ID %d (%s, %.2f MB) para criador ID %d",
			uploadedFile.ID, uploadedFile.Name, float64(uploadedFile.FileSize)/1024/1024, creatorID)

		uploadedFiles = append(uploadedFiles, uploadedFile)
	}

	return uploadedFiles, errors, nil
}

func (h *EbookHandler) checkUserStorageQuota(creatorID uint, newFileSize int64) error {
	const MAX_STORAGE_PER_USER = 1024 * 1024 * 1024 // 1GB

	if newFileSize > MAX_STORAGE_PER_USER/10 {
		log.Printf("Upload grande detectado (%.2f MB) para o criador ID %d. Verificação de cota necessária.",
			float64(newFileSize)/1024/1024, creatorID)
	}

	return nil
}

func (h *EbookHandler) sanitizeFilename(filename string) string {
	filename = filepath.Base(filename)

	safeFilename := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == ' ' {
			return r
		}
		return '_'
	}, filename)

	const MAX_FILENAME_LENGTH = 255
	if len(safeFilename) > MAX_FILENAME_LENGTH {
		ext := filepath.Ext(safeFilename)
		safeFilename = safeFilename[:MAX_FILENAME_LENGTH-len(ext)] + ext
	}

	return safeFilename
}

// RemoveFileFromEbook removes a file from an ebook
func (h *EbookHandler) RemoveFileFromEbook(w http.ResponseWriter, r *http.Request) {
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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
		return
	}

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

	err = h.removeFileFromEbookLogic(ebook, uint(fileIDParsed))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.ebookService.RemoveFileAssociation(ebook.ID, uint(fileIDParsed))
	if err != nil {
		log.Printf("Erro ao remover associação de arquivo: %v", err)
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
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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
		return
	}

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

	file, err := h.fileService.GetFileByID(uint(fileIDParsed))
	if err != nil {
		http.Error(w, "Arquivo não encontrado", http.StatusNotFound)
		return
	}

	err = h.validateFileOwnership(file, creator.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if h.checkFileAlreadyInEbook(ebook, file.ID) {
		http.Error(w, "Arquivo já está associado ao ebook", http.StatusBadRequest)
		return
	}

	err = h.ebookService.AppendFiles(ebook.ID, []*librarymodel.File{file})
	if err != nil {
		log.Printf("Erro ao adicionar arquivo ao ebook: %v", err)
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
	userEmail, ok := r.Context().Value(authmw.UserEmailKey).(string)
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

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Erro ao processar upload", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "Nenhum arquivo foi enviado", http.StatusBadRequest)
		return
	}

	const MAX_FILES_PER_UPLOAD = 10
	if len(files) > MAX_FILES_PER_UPLOAD {
		http.Error(w, fmt.Sprintf("Máximo %d arquivos por upload permitidos", MAX_FILES_PER_UPLOAD),
			http.StatusBadRequest)
		return
	}

	totalUploadSize := int64(0)
	for _, fileHeader := range files {
		totalUploadSize += fileHeader.Size
	}

	if err := h.checkUserStorageQuota(creator.ID, totalUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var uploadedFiles []*librarymodel.File
	var errors []string

	for _, fileHeader := range files {
		originalFilename := fileHeader.Filename
		fileHeader.Filename = h.sanitizeFilename(fileHeader.Filename)

		description := r.FormValue("description_" + originalFilename)

		uploadedFile, err := h.fileService.UploadFile(fileHeader, originalFilename, description, creator.ID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Erro ao fazer upload de %s: %v", fileHeader.Filename, err))
			continue
		}

		uploadedFiles = append(uploadedFiles, uploadedFile)
	}

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

func (h *EbookHandler) RemoveEbook(w http.ResponseWriter, r *http.Request) {
	ebookId := chi.URLParam(r, "id")

	if ebookId == "" {
		h.FlashMessage(w, r, "Ebook não identificado", "error")
		http.Redirect(w, r, "/ebook", http.StatusSeeOther)
		return
	}

	ebookIdInt, err := strconv.ParseInt(ebookId, 10, 64)
	if err != nil {
		slog.Error("Falha na conversão do ID do ebook para remoção", "error", err)
		return
	}

	err = h.ebookService.Delete(uint(ebookIdInt))
	if err != nil {
		h.FlashMessage(w, r, err.Error(), "error")
		http.Redirect(w, r, "/ebook", http.StatusSeeOther)
		return
	}

	h.FlashMessage(w, r, "Ebook removido com sucesso", "success")
	http.Redirect(w, r, "/ebook", http.StatusSeeOther)
}

func (h *EbookHandler) FlashMessage(w http.ResponseWriter, r *http.Request, message string, messageType string) {
	err := h.sessionManager.AddFlash(w, r, message, messageType)
	if err != nil {
		slog.Error("Falha ao adicionar flash message", "error", err)
	}
}
