package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/anglesson/simple-web-server/pkg/storage"
)

type FileService interface {
	UploadFile(file *multipart.FileHeader, name, description string, creatorID uint) (*models.File, error)
	GetFilesByCreator(creatorID uint) ([]*models.File, error)
	GetFilesByCreatorPaginated(creatorID uint, query repository.FileQuery) ([]*models.File, int64, error)
	GetActiveByCreator(creatorID uint) ([]*models.File, error)
	GetFileByID(id uint) (*models.File, error)
	UpdateFile(id uint, name, description string) error
	DeleteFile(id uint) error
	GetFilesByType(creatorID uint, fileType string) ([]*models.File, error)
	ValidateFile(file *multipart.FileHeader) error
	GetFileType(ext string) string
}

type fileService struct {
	fileRepository repository.FileRepository
	s3Storage      storage.S3Storage
}

func NewFileService(fileRepository repository.FileRepository, s3Storage storage.S3Storage) FileService {
	return &fileService{
		fileRepository: fileRepository,
		s3Storage:      s3Storage,
	}
}

func (s *fileService) UploadFile(file *multipart.FileHeader, name, description string, creatorID uint) (*models.File, error) {
	// Validar arquivo
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// Gerar nome único para o arquivo
	originalName := file.Filename
	fileExt := filepath.Ext(originalName)
	uniqueID := s.generateUniqueID()
	fileName := fmt.Sprintf("%s-%s%s",
		strings.TrimSuffix(originalName, fileExt),
		uniqueID,
		fileExt,
	)

	// Determinar tipo do arquivo
	fileType := s.getFileType(fileExt)

	// Upload para S3
	const fileCache = "private, no-cache, no-store, must-revalidate"
	s3Key := fmt.Sprintf("files/%d/%s", creatorID, fileName)
	s3URL, err := s.s3Storage.UploadFile(file, s3Key, fileCache)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer upload para S3: %w", err)
	}

	if strings.TrimSpace(name) != "" {
		fileName = name
	}

	// Criar registro no banco
	fileModel := models.NewFile(
		fileName,
		originalName,
		description,
		fileType,
		s3Key,
		s3URL,
		file.Size,
		creatorID,
	)

	if err := s.fileRepository.Create(fileModel); err != nil {
		// Se falhar, tentar deletar do S3
		s.s3Storage.DeleteFile(s3Key)
		return nil, fmt.Errorf("erro ao salvar arquivo no banco: %w", err)
	}

	return fileModel, nil
}

func (s *fileService) GetFilesByCreator(creatorID uint) ([]*models.File, error) {
	files, err := s.fileRepository.FindByCreator(creatorID)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		expirationTime := 5 * time.Minute
		file.S3URL = s.s3Storage.GenerateDownloadLinkWithExpiration(file.S3Key, int(expirationTime.Seconds()))
	}
	return files, nil
}

func (s *fileService) GetFilesByCreatorPaginated(creatorID uint, query repository.FileQuery) ([]*models.File, int64, error) {
	query.CreatorID = creatorID
	files, total, err := s.fileRepository.FindByCreatorPaginated(query)
	if err != nil {
		return nil, 0, err
	}
	for _, file := range files {
		expirationTime := 5 * time.Minute
		if file.FileType == "pdf" {
			file.S3URL = s.s3Storage.GeneratePreviewLinkWithExpiration(file.S3Key, "application/pdf", int(expirationTime.Seconds()))
		} else {
			file.S3URL = s.s3Storage.GeneratePreviewLinkWithExpiration(file.S3Key, file.FileType, int(expirationTime.Seconds()))
		}
	}
	return files, total, nil
}

func (s *fileService) GetActiveByCreator(creatorID uint) ([]*models.File, error) {
	return s.fileRepository.FindActiveByCreator(creatorID)
}

func (s *fileService) GetFileByID(id uint) (*models.File, error) {
	return s.fileRepository.FindByID(id)
}

func (s *fileService) UpdateFile(id uint, name, description string) error {
	file, err := s.fileRepository.FindByID(id)
	if err != nil {
		return err
	}

	file.SetName(name)
	file.SetDescription(description)

	return s.fileRepository.Update(file)
}

func (s *fileService) DeleteFile(id uint) error {
	file, err := s.fileRepository.FindByID(id)
	if err != nil {
		return err
	}

	if file.InUse() {
		slog.Info("Arquivo em uso.")
		return errors.New("você não pode excluir um arquivo usado em um ebook")
	}

	// Deletar do S3
	if err := s.s3Storage.DeleteFile(file.S3Key); err != nil {
		return fmt.Errorf("erro ao deletar arquivo do S3: %w", err)
	}

	// Deletar do banco
	return s.fileRepository.Delete(id)
}

func (s *fileService) GetFilesByType(creatorID uint, fileType string) ([]*models.File, error) {
	return s.fileRepository.FindByType(creatorID, fileType)
}

func (s *fileService) ValidateFile(file *multipart.FileHeader) error {
	return s.validateFile(file)
}

func (s *fileService) GetFileType(ext string) string {
	return s.getFileType(ext)
}

func (s *fileService) validateFile(file *multipart.FileHeader) error {
	// Verificar tamanho (máximo 50MB)
	const maxSize = 50 * 1024 * 1024 // 50MB
	if file.Size > maxSize {
		return fmt.Errorf("arquivo muito grande. Tamanho máximo: 50MB")
	}

	// Verificar extensão
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".pdf", ".doc", ".docx", ".jpg", ".jpeg", ".png", ".gif"}

	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("tipo de arquivo não permitido. Tipos aceitos: %v", allowedExts)
	}

	// Verificar MIME type
	if err := s.validateMimeType(file); err != nil {
		return err
	}

	return nil
}

// validateMimeType validates the actual MIME type of the file
func (s *fileService) validateMimeType(file *multipart.FileHeader) error {
	// Open file to check MIME type
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer src.Close()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer)

	// Allowed MIME types
	allowedMimeTypes := map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}

	if !allowedMimeTypes[mimeType] {
		return fmt.Errorf("tipo MIME não permitido: %s", mimeType)
	}

	return nil
}

func (s *fileService) getFileType(ext string) string {
	ext = strings.ToLower(ext)

	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "document"
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image"
	default:
		return "other"
	}
}

func (s *fileService) generateUniqueID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
