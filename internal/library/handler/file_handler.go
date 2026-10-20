package handler

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	accountmw "github.com/Assinee-Inc/docffy/internal/account/handler/middleware"
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountrepo "github.com/Assinee-Inc/docffy/internal/account/repository"
	accountservice "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	libraryrepo "github.com/Assinee-Inc/docffy/internal/library/repository"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/go-chi/chi/v5"
)

type FileService interface {
	UploadFile(file *multipart.FileHeader, name, description string, creatorID uint) (*librarymodel.File, error)
	GetFilesByCreator(creatorID uint) ([]*librarymodel.File, error)
	GetActiveByCreator(creatorID uint) ([]*librarymodel.File, error)
	GetFilesByCreatorPaginated(creatorID uint, query libraryrepo.FileQuery) ([]*librarymodel.File, int64, error)
	GetFileByID(id uint) (*librarymodel.File, error)
	GetFileByPublicID(publicID string) (*librarymodel.File, error)
	UpdateFile(id uint, name, description string) error
	DeleteFile(id uint) error
	GetFilesByType(creatorID uint, fileType string) ([]*librarymodel.File, error)
	ValidateFile(file *multipart.FileHeader) error
	GetFileType(ext string) string
}

type FileHandler struct {
	fileService FileService
}

func NewFileHandler(fileService FileService, _ ...any) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

func (h *FileHandler) getCurrentCreator(r *http.Request) (*accountmodel.Creator, error) {
	if creator := accountmw.CreatorFromContext(r.Context()); creator != nil {
		return creator, nil
	}

	userID, _ := authmw.GetAuthenticatedUserID(r.Context())
	if userID == 0 {
		email := authmw.GetAuthenticatedUserEmail(r.Context())
		if email != "" && database.DB != nil {
			userRepo := authrepo.NewGormUserRepository(database.DB)
			if u := userRepo.FindByEmail(email); u != nil {
				userID = u.ID
			}
		}
	}
	if userID == 0 || database.DB == nil {
		return nil, fmt.Errorf("usuário não encontrado")
	}

	creatorService := accountservice.NewCreatorService(accountrepo.NewGormCreatorRepository(database.DB), nil, nil, nil, nil)
	creator, err := creatorService.FindCreatorByUserID(userID)
	if err != nil || creator == nil {
		return nil, fmt.Errorf("criador não encontrado")
	}

	return creator, nil
}

func (h *FileHandler) FileIndexView(w http.ResponseWriter, r *http.Request) {
	creator, err := h.getCurrentCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}
	paginationResult := salesmodel.NewPagination(page, pageSize)

	searchTerm := r.URL.Query().Get("search")
	fileType := r.URL.Query().Get("type")

	query := libraryrepo.FileQuery{
		CreatorID:  creator.ID,
		SearchTerm: searchTerm,
		FileType:   fileType,
		Pagination: paginationResult,
	}

	files, total, err := h.fileService.GetFilesByCreatorPaginated(creator.ID, query)
	if err != nil {
		web.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro ao carregar arquivos")
		return
	}

	paginationResult.SetTotal(total)

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success":    true,
		"files":      files,
		"pagination": paginationResult,
		"creator":    creator,
	})
}

func (h *FileHandler) FileUploadSubmit(w http.ResponseWriter, r *http.Request) {
	creator, err := h.getCurrentCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
		return
	}

	err = r.ParseMultipartForm(100 << 20)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "INVALID_FORM", "Erro ao processar arquivo. Verifique o tamanho máximo permitido (100MB).")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "FILE_REQUIRED", "Nenhum arquivo foi selecionado.")
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	description := r.FormValue("description")

	uploadedFile, err := h.fileService.UploadFile(fileHeader, name, description, creator.ID)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "UPLOAD_ERROR", fmt.Sprintf("Erro ao fazer upload: %s", err.Error()))
		return
	}

	web.RespondJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Arquivo '%s' enviado com sucesso!", uploadedFile.OriginalName),
		"file":    uploadedFile,
	})
}

func (h *FileHandler) FileUpdateSubmit(w http.ResponseWriter, r *http.Request) {
	creator, err := h.getCurrentCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
		return
	}

	filePublicID := chi.URLParam(r, "id")
	if filePublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de arquivo inválido.")
		return
	}

	file, err := h.fileService.GetFileByPublicID(filePublicID)
	if err != nil || file == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Arquivo não encontrado.")
		return
	}

	if file.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Você não tem permissão para editar este arquivo.")
		return
	}

	var name, description string
	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		name = body.Name
		description = body.Description
	} else {
		name = r.FormValue("name")
		description = r.FormValue("description")
	}

	err = h.fileService.UpdateFile(file.ID, name, description)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "UPDATE_ERROR", fmt.Sprintf("Erro ao atualizar arquivo: %s", err.Error()))
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Arquivo atualizado com sucesso!",
	})
}

func (h *FileHandler) FileDeleteSubmit(w http.ResponseWriter, r *http.Request) {
	creator, err := h.getCurrentCreator(r)
	if err != nil {
		web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado")
		return
	}

	filePublicID := chi.URLParam(r, "id")
	if filePublicID == "" {
		web.RespondError(w, http.StatusBadRequest, "INVALID_ID", "ID de arquivo inválido.")
		return
	}

	file, err := h.fileService.GetFileByPublicID(filePublicID)
	if err != nil || file == nil {
		web.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Arquivo não encontrado.")
		return
	}

	if file.CreatorID != creator.ID {
		web.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Você não tem permissão para deletar este arquivo.")
		return
	}

	err = h.fileService.DeleteFile(file.ID)
	if err != nil {
		web.RespondError(w, http.StatusBadRequest, "DELETE_ERROR", fmt.Sprintf("Erro ao deletar arquivo: %s", err.Error()))
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Arquivo deletado com sucesso!",
	})
}
