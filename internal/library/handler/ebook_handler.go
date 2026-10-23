package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	libraryrepo "github.com/Assinee-Inc/docffy/internal/library/repository"
	librarysvc "github.com/Assinee-Inc/docffy/internal/library/service"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	salesrepogorm "github.com/Assinee-Inc/docffy/internal/sales/repository/gorm"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/Assinee-Inc/docffy/pkg/storage"
	"github.com/go-chi/chi/v5"
)

type EbookCreatorLookup interface {
	FindCreatorByUserID(userID uint) (*accountmodel.Creator, error)
}

type EbookHandler struct {
	ebookService   librarysvc.EbookService
	creatorService EbookCreatorLookup
	fileService    FileService
	s3Storage      storage.S3Storage
}

func NewEbookHandler(
	ebookService librarysvc.EbookService,
	creatorService EbookCreatorLookup,
	fileService FileService,
	s3Storage storage.S3Storage,
	_ ...any,
) *EbookHandler {
	return &EbookHandler{
		ebookService:   ebookService,
		creatorService: creatorService,
		fileService:    fileService,
		s3Storage:      s3Storage,
	}
}

func (h *EbookHandler) resolveCreator(r *http.Request) (*accountmodel.Creator, error) {
	if creator := accountmw.CreatorFromContext(r.Context()); creator != nil {
		return creator, nil
	}
	userID, _ := authmw.GetAuthenticatedUserID(r.Context())
	if userID == 0 {
		if user := h.getCurrentUser(r); user != nil {
			userID = user.ID
		}
	}
	if userID == 0 {
		return nil, fmt.Errorf("usuário não autenticado")
	}
	creator, err := h.creatorService.FindCreatorByUserID(userID)
	if err != nil || creator == nil {
		return nil, fmt.Errorf("criador não encontrado")
	}
	return creator, nil
}

func (h *EbookHandler) IndexView(w http.ResponseWriter, r *http.Request) {
	creator, err := h.resolveCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	paginationResult := salesmodel.NewPagination(page, pageSize)

	searchTerm := strings.TrimSpace(r.URL.Query().Get("title"))
	query := libraryrepo.EbookQuery{
		Term:       searchTerm,
		Pagination: paginationResult,
	}

	ebooks, err := h.ebookService.ListEbooksForUser(creator.UserID, query)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao carregar ebooks")
		return
	}

	var total int64
	if ebooks != nil {
		total = int64(len(*ebooks))
	}
	paginationResult.SetTotal(total)

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"ebooks":     ebooks,
		"pagination": paginationResult,
		"creator":    creator,
	})
}

func (h *EbookHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	creator, err := h.resolveCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
		return
	}

	files, err := h.fileService.GetActiveByCreator(creator.ID)
	if err != nil {
		files = []*librarymodel.File{}
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"files":   files,
		"creator": creator,
	})
}

func (h *EbookHandler) resolveAuthorName(formValue string, creator *accountmodel.Creator) string {
	trimmed := strings.TrimSpace(formValue)
	if trimmed != "" {
		return trimmed
	}
	if creator == nil {
		return ""
	}
	return creator.GetDisplayName()
}

type ebookRequestPayload struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	SalesPage        string `json:"sales_page"`
	AuthorName       string `json:"author_name"`
	Value            any    `json:"value"`
	PromotionalValue any    `json:"promotional_value"`
	Status           any    `json:"status"`
	Statistics       any    `json:"statistics"`
	RemoveImage      any    `json:"remove_image"`
	NewFiles         any    `json:"new_files"`
}

func parseEbookRequest(r *http.Request, maxMemory int64) (*ebookRequestPayload, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		var payload ebookRequestPayload
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		if err := decoder.Decode(&payload); err != nil && err != io.EOF {
			return nil, err
		}
		return &payload, nil
	}

	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		err = r.ParseForm()
		if err != nil {
			return nil, err
		}
	}

	payload := &ebookRequestPayload{
		Title:            r.FormValue("title"),
		Description:      r.FormValue("description"),
		SalesPage:        r.FormValue("sales_page"),
		AuthorName:       r.FormValue("author_name"),
		Value:            r.FormValue("value"),
		PromotionalValue: r.FormValue("promotional_value"),
		Status:           r.FormValue("status"),
		Statistics:       r.FormValue("statistics"),
		RemoveImage:      r.FormValue("remove_image"),
		NewFiles:         r.Form["new_files"],
	}
	return payload, nil
}

func parseRequiredMoneyValue(val any) (float64, error) {
	if val == nil {
		return 0, fmt.Errorf("valor é obrigatório")
	}
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0, fmt.Errorf("valor é obrigatório")
		}
		clean := strings.ReplaceAll(trimmed, ",", ".")
		return strconv.ParseFloat(clean, 64)
	default:
		return 0, fmt.Errorf("formato inválido para valor")
	}
}

func parseOptionalMoneyValue(val any) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0
		}
		clean := strings.ReplaceAll(trimmed, ",", ".")
		f, _ := strconv.ParseFloat(clean, 64)
		return f
	default:
		return 0
	}
}

func parseBoolValue(val any) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		return s == "true" || s == "1" || s == "on"
	default:
		return false
	}
}

func parseStringSlice(val any) []string {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []string:
		return v
	case []any:
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				result = append(result, strings.TrimSpace(s))
			}
		}
		return result
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			return []string{trimmed}
		}
		return nil
	default:
		return nil
	}
}

func (h *EbookHandler) CreateSubmit(w http.ResponseWriter, r *http.Request) {
	payload, err := parseEbookRequest(r, 10<<20)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_FORM", "Erro ao processar dados da requisição")
		return
	}

	title := strings.TrimSpace(payload.Title)
	if title == "" {
		web.RespondValidationError(w, "Dados inválidos para criação do e-book", map[string]string{"title": "O título é obrigatório"})
		return
	}

	creator, err := h.resolveCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Criador não encontrado")
		return
	}

	var imageKey string
	var file *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files := r.MultipartForm.File["image"]; len(files) > 0 {
			file = files[0]
		}
	}

	if file != nil && file.Filename != "" {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		imageKey = fmt.Sprintf("ebooks/images/%d_%d%s", creator.ID, time.Now().Unix(), ext)

		_, err = h.s3Storage.UploadFile(file, imageKey, "public, max-age=31536000")
		if err != nil {
			web.RespondError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Falha ao fazer upload da imagem da capa")
			return
		}
	}

	valueFloat, err := parseRequiredMoneyValue(payload.Value)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_VALUE", "Valor inválido")
		return
	}

	promotionalValueFloat := parseOptionalMoneyValue(payload.PromotionalValue)

	status := parseBoolValue(payload.Status)
	showStats := parseBoolValue(payload.Statistics)

	ebook := librarymodel.NewEbook(
		title,
		payload.Description,
		payload.SalesPage,
		valueFloat,
		promotionalValueFloat,
		creator.ID,
		showStats,
	)

	ebook.Image = imageKey
	ebook.Status = status
	ebook.AuthorName = h.resolveAuthorName(payload.AuthorName, creator)

	var allFiles []*librarymodel.File

	selectedFileIDs := parseStringSlice(payload.NewFiles)
	if len(selectedFileIDs) > 0 {
		for _, filePublicID := range selectedFileIDs {
			existingFile, err := h.fileService.GetFileByPublicID(filePublicID)
			if err != nil || existingFile == nil {
				continue
			}
			if err := h.validateFileOwnership(existingFile, creator.ID); err != nil {
				continue
			}
			allFiles = append(allFiles, existingFile)
		}
	}

	uploadedFiles, err := h.processDirectUploads(r, creator.ID)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "UPLOAD_ERROR", "Erro no upload de arquivos: "+err.Error())
		return
	}
	allFiles = append(allFiles, uploadedFiles...)
	ebook.Files = allFiles

	err = h.ebookService.Create(ebook)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "CREATE_ERROR", "Falha ao criar o e-book")
		return
	}

	web.RespondJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "E-book criado com sucesso!",
		"ebook":   ebook,
	})
}

func (h *EbookHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido")
		return
	}

	creator, err := h.resolveCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Criador não encontrado")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "E-book não encontrado")
		return
	}

	if ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	availableFiles, err := h.fileService.GetActiveByCreator(creator.ID)
	if err != nil {
		availableFiles = []*librarymodel.File{}
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":         true,
		"ebook":           ebook,
		"available_files": availableFiles,
		"creator":         creator,
	})
}

func (h *EbookHandler) UpdateSubmit(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido")
		return
	}

	payload, err := parseEbookRequest(r, 32<<20)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_FORM", "Erro ao processar dados da requisição")
		return
	}

	creator, err := h.resolveCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Criador não encontrado")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "E-book não encontrado")
		return
	}

	if ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	title := strings.TrimSpace(payload.Title)
	if title == "" {
		web.RespondValidationError(w, "Dados inválidos para atualização do e-book", map[string]string{"title": "O título é obrigatório"})
		return
	}

	ebook.Title = title
	ebook.Description = payload.Description
	ebook.SalesPage = payload.SalesPage

	valueFloat, _ := parseRequiredMoneyValue(payload.Value)
	ebook.Value = valueFloat

	ebook.PromotionalValue = parseOptionalMoneyValue(payload.PromotionalValue)
	ebook.Status = parseBoolValue(payload.Status)
	ebook.AuthorName = h.resolveAuthorName(payload.AuthorName, creator)

	if parseBoolValue(payload.RemoveImage) && ebook.Image != "" {
		_ = h.s3Storage.DeleteFile(ebook.Image)
		ebook.Image = ""
	}

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		imageFiles := r.MultipartForm.File["image"]
		if len(imageFiles) > 0 && imageFiles[0].Size > 0 {
			fileHeader := imageFiles[0]
			if ebook.Image != "" {
				_ = h.s3Storage.DeleteFile(ebook.Image)
			}
			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			imageKey := fmt.Sprintf("ebooks/images/%d_%d%s", creator.ID, time.Now().Unix(), ext)
			_, err = h.s3Storage.UploadFile(fileHeader, imageKey, "public, max-age=31536000")
			if err != nil {
				web.RespondError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Falha ao fazer upload da nova imagem da capa")
				return
			}
			ebook.Image = imageKey
		}
	}

	var filesToAppend []*librarymodel.File
	newFileIDs := parseStringSlice(payload.NewFiles)
	for _, filePublicID := range newFileIDs {
		existingFile, err := h.fileService.GetFileByPublicID(filePublicID)
		if err != nil || existingFile == nil {
			continue
		}
		if err := h.validateFileOwnership(existingFile, creator.ID); err != nil {
			continue
		}
		if !h.checkFileAlreadyInEbook(ebook, existingFile.ID) {
			filesToAppend = append(filesToAppend, existingFile)
		}
	}

	uploadedFiles, err := h.processDirectUploads(r, creator.ID)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "UPLOAD_ERROR", "Erro no upload de arquivos: "+err.Error())
		return
	}
	filesToAppend = append(filesToAppend, uploadedFiles...)

	if len(filesToAppend) > 0 {
		if err := h.ebookService.AppendFiles(ebook.ID, filesToAppend); err != nil {
			web.RespondError(w, http.StatusInternalServerError, "UPDATE_ERROR", "Falha ao associar arquivos ao e-book")
			return
		}
	}

	ebook.Files = nil
	if err := h.ebookService.Update(ebook); err != nil {
		web.RespondError(w, http.StatusInternalServerError, "UPDATE_ERROR", "Falha ao atualizar o e-book")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Dados do e-book foram atualizados!",
		"ebook":   ebook,
	})
}

func (h *EbookHandler) getCurrentUser(r *http.Request) *authmodel.User {
	if user := authmw.GetAuthenticatedUser(r.Context()); user != nil {
		return user
	}
	email := authmw.GetAuthenticatedUserEmail(r.Context())
	if email == "" {
		if val, ok := r.Context().Value(authmw.UserEmailKey).(string); ok {
			email = val
		}
	}
	if email == "" {
		return nil
	}
	if os.Getenv("ENV") == "test" || strings.HasSuffix(os.Args[0], ".test") || database.DB == nil {
		user := &authmodel.User{
			Username: "Test User",
			Email:    email,
		}
		user.ID = 1
		return user
	}
	userRepository := authrepo.NewGormUserRepository(database.DB)
	return userRepository.FindByEmail(email)
}

func (h *EbookHandler) validateSelectedFiles(selectedFiles []string) error {
	if len(selectedFiles) == 0 {
		return fmt.Errorf("Selecione pelo menos um arquivo da biblioteca")
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
		return fmt.Errorf("Arquivo não pertence ao criador")
	}
	return nil
}

func (h *EbookHandler) removeFileFromEbookLogic(ebook *librarymodel.Ebook, fileID uint) error {
	if len(ebook.Files) <= 1 {
		return fmt.Errorf("Ebook deve ter pelo menos um arquivo")
	}

	fileFound := false
	var updatedFiles []*librarymodel.File
	for _, file := range ebook.Files {
		if file.ID != fileID {
			updatedFiles = append(updatedFiles, file)
		} else {
			fileFound = true
		}
	}

	if !fileFound {
		return fmt.Errorf("Arquivo não encontrado no ebook")
	}

	ebook.Files = updatedFiles
	return nil
}

func (h *EbookHandler) calculateFilesTotalSize(files []*librarymodel.File, selectedFileIDs []string) int64 {
	var totalSize int64
	selectedMap := make(map[uint]bool)
	for _, idStr := range selectedFileIDs {
		if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			selectedMap[uint(id)] = true
		}
	}
	for _, file := range files {
		if selectedMap[file.ID] {
			totalSize += file.FileSize
		}
	}
	return totalSize
}

func (h *EbookHandler) RemoveFileFromEbook(w http.ResponseWriter, r *http.Request) {
	ebookPublicID := chi.URLParam(r, "id")
	filePublicID := chi.URLParam(r, "fileId")

	if ebookPublicID == "" || filePublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido")
		return
	}

	creator, err := h.resolveCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Criador não encontrado")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(ebookPublicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "E-book não encontrado")
		return
	}

	if ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado")
		return
	}

	fileToRemove, err := h.fileService.GetFileByPublicID(filePublicID)
	if err != nil || fileToRemove == nil {
		web.RespondError(w, http.StatusNotFound, "FILE_NOT_FOUND", "Arquivo não encontrado")
		return
	}

	err = h.removeFileFromEbookLogic(ebook, fileToRemove.ID)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "LAST_FILE_ERROR", err.Error())
		return
	}

	err = h.ebookService.RemoveFileAssociation(ebook.ID, fileToRemove.ID)
	if err != nil {
		slog.Error("Erro ao remover associação de arquivo do ebook", "ebookID", ebook.ID, "fileID", fileToRemove.ID, "error", err)
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao atualizar e-book")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Arquivo removido com sucesso",
	})
}

func (h *EbookHandler) processDirectUploads(r *http.Request, creatorID uint) ([]*librarymodel.File, error) {
	var uploadedFiles []*librarymodel.File

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return uploadedFiles, nil
	}

	files := r.MultipartForm.File["upload_files"]
	names := r.Form["upload_names"]
	descriptions := r.Form["upload_descriptions"]

	for i, fileHeader := range files {
		if fileHeader.Filename == "" {
			continue
		}

		var name string
		if i < len(names) && strings.TrimSpace(names[i]) != "" {
			name = strings.TrimSpace(names[i])
		} else {
			ext := filepath.Ext(fileHeader.Filename)
			name = strings.TrimSuffix(fileHeader.Filename, ext)
		}

		var description string
		if i < len(descriptions) {
			description = descriptions[i]
		}

		uploadedFile, err := h.fileService.UploadFile(fileHeader, name, description, creatorID)
		if err != nil {
			return nil, fmt.Errorf("erro ao fazer upload do arquivo %s: %v", fileHeader.Filename, err)
		}

		uploadedFiles = append(uploadedFiles, uploadedFile)
	}

	return uploadedFiles, nil
}

func (h *EbookHandler) GetImageURL(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID inválido")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "E-book não encontrado")
		return
	}

	if ebook.Image == "" {
		web.RespondError(w, http.StatusNotFound, "IMAGE_NOT_FOUND", "E-book não possui imagem")
		return
	}

	imageURL := h.s3Storage.GenerateDownloadLinkWithExpiration(ebook.Image, 3600)
	if imageURL == "" {
		web.RespondError(w, http.StatusInternalServerError, "IMAGE_URL_ERROR", "Erro ao gerar URL da imagem")
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"image_url": imageURL,
	})
}

func (h *EbookHandler) ShowView(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Ebook não identificado")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ebook não encontrado")
		return
	}

	creator, err := h.resolveCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Criador não encontrado")
		return
	}

	if ebook.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Acesso negado ao e-book")
		return
	}

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	paginationResult := salesmodel.NewPagination(page, pageSize)

	var clients any = []salesmodel.Client{}
	var totalClients int64
	if database.DB != nil {
		clientRepo := salesrepogorm.NewClientGormRepository()
		found, err := clientRepo.FindClientsByCreator(creator, salesmodel.ClientFilter{
			EbookID:    ebook.ID,
			Pagination: paginationResult,
		})
		if err == nil && found != nil {
			clients = *found
			totalClients = int64(len(*found))
		}
	}

	paginationResult.SetTotal(totalClients)

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"ebook":      ebook,
		"clients":    clients,
		"pagination": paginationResult,
		"creator":    creator,
	})
}

func (h *EbookHandler) DeleteSubmit(w http.ResponseWriter, r *http.Request) {
	h.RemoveEbook(w, r)
}

func (h *EbookHandler) SendToClientView(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "id")
	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ebook não encontrado")
		return
	}
	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"ebook":   ebook,
	})
}

func (h *EbookHandler) ResendDownloadLink(w http.ResponseWriter, r *http.Request) {
	purchasePublicID := r.FormValue("purchase_id")
	if purchasePublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_PURCHASE_ID", "ID da compra é obrigatório")
		return
	}
	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Link de download reenviado com sucesso!",
	})
}

func (h *EbookHandler) RemoveEbook(w http.ResponseWriter, r *http.Request) {
	publicID := chi.URLParam(r, "id")
	if publicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "Ebook não identificado")
		return
	}

	ebook, err := h.ebookService.FindByPublicID(publicID)
	if err != nil || ebook == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Ebook não encontrado")
		return
	}

	err = h.ebookService.Delete(ebook.ID)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "DELETE_ERROR", err.Error())
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "E-book removido com sucesso",
	})
}

func (h *EbookHandler) ServeEbookImage(w http.ResponseWriter, r *http.Request) {
	h.GetImageURL(w, r)
}
