package service

import (
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
)

// PresignedUploadRequest representa o payload para solicitar uma URL de upload pré-assinada do S3.
type PresignedUploadRequest struct {
	FileName    string `json:"file_name" validate:"required,min=3,max=255"`
	FileSize    int64  `json:"file_size" validate:"required,gt=0,lte=104857600"` // Max 100MB
	ContentType string `json:"content_type" validate:"required"`
}

// PresignedUploadResponse representa a resposta contendo a URL pré-assinada para upload direto.
type PresignedUploadResponse struct {
	Success   bool              `json:"success"`
	UploadURL string            `json:"upload_url"`
	Key       string            `json:"key"`
	Method    string            `json:"method"` // "PUT"
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresIn int               `json:"expires_in"` // Expiração em segundos (ex.: 60)
}

// CompleteUploadRequest representa os metadados enviados após o upload direto no S3.
type CompleteUploadRequest struct {
	Key          string `json:"key" validate:"required"`
	Name         string `json:"name" validate:"required,min=1,max=255"`
	Description  string `json:"description"`
	OriginalName string `json:"original_name" validate:"required"`
	Size         int64  `json:"size" validate:"required,gt=0"`
	ContentType  string `json:"content_type"`
}

// CompleteUploadResponse representa a resposta após registro dos metadados do arquivo.
type CompleteUploadResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	File    *librarymodel.File `json:"file"`
}
