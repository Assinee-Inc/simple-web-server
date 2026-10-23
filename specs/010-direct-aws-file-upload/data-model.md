# Phase 1 Data Model: Upload Direto para AWS S3 com URLs Pré-assinadas

**Feature**: `010-direct-aws-file-upload`  
**Date**: 2026-10-23

---

## 1. DTOs (Data Transfer Objects)

### 1.1 `PresignedUploadRequest` (Backend Go)
Objeto recebido no endpoint `POST /file/upload-url`:

```go
type PresignedUploadRequest struct {
    FileName    string `json:"file_name" validate:"required,min=3,max=255"`
    FileSize    int64  `json:"file_size" validate:"required,gt=0,lte=104857600"` // Max 100MB
    ContentType string `json:"content_type" validate:"required"`
}
```

### 1.2 `PresignedUploadResponse` (Backend Go)
Objeto retornado pelo endpoint `POST /file/upload-url`:

```go
type PresignedUploadResponse struct {
    Success   bool              `json:"success"`
    UploadURL string            `json:"upload_url"`
    Key       string            `json:"key"`
    Method    string            `json:"method"` // "PUT"
    Headers   map[string]string `json:"headers,omitempty"`
    ExpiresIn int               `json:"expires_in"` // Segundos (ex.: 900 = 15 minutos)
}
```

### 1.3 `CompleteUploadRequest` (Backend Go)
Objeto recebido no endpoint `POST /file/complete-upload`:

```go
type CompleteUploadRequest struct {
    Key          string `json:"key" validate:"required"`
    Name         string `json:"name" validate:"required,min=1,max=255"`
    Description  string `json:"description"`
    OriginalName string `json:"original_name" validate:"required"`
    Size         int64  `json:"size" validate:"required,gt=0"`
    ContentType  string `json:"content_type"`
}
```

### 1.4 `CompleteUploadResponse` (Backend Go)
Objeto retornado pelo endpoint `POST /file/complete-upload`:

```go
type CompleteUploadResponse struct {
    Success bool              `json:"success"`
    Message string            `json:"message"`
    File    *librarymodel.File `json:"file"`
}
```

---

## 2. Entidade de Domínio (`File`)

A entidade `librarymodel.File` existente no banco de dados (`internal/library/model/file.go`) já possui todos os campos necessários para persistência:

```go
type File struct {
    ID           uint           `gorm:"primaryKey" json:"id"`
    PublicID     string         `gorm:"uniqueIndex;not null" json:"public_id"`
    Name         string         `gorm:"not null" json:"name"`
    OriginalName string         `gorm:"not null" json:"original_name"`
    Description  string         `json:"description"`
    FileType     string         `gorm:"not null" json:"file_type"`
    S3Key        string         `gorm:"not null" json:"s3_key"`
    S3URL        string         `gorm:"not null" json:"s3_url"`
    Size         int64          `gorm:"not null" json:"size"`
    CreatorID    uint           `gorm:"not null" json:"creator_id"`
    IsActive     bool           `gorm:"default:true" json:"is_active"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
```

---

## 3. Máquina de Estados do Upload

```text
       [Início]
          │
          ▼
   Solicitação de URL
  (POST /file/upload-url)
          │
          ├── Validação de tipo/tamanho falhou ──► [400 Bad Request]
          ├── Não autenticado/sem onboarding   ──► [401 / 403 Forbidden]
          │
          ▼
   URL Pré-assinada emitida
  (Key: files/{creatorID}/...)
          │
          ▼
   Upload Direto S3 (PUT)
  (Navegador ──► AWS S3)
          │
          ├── Erro de rede/cancelamento        ──► [Upload Interrompido - Sem resíduo no DB]
          │
          ▼ (200 OK do S3)
   Confirmação de Upload
  (POST /file/complete-upload)
          │
          ├── Key não pertence ao creator      ──► [403 Forbidden]
          ├── Dados inválidos                  ──► [400 Bad Request]
          │
          ▼
   Registro no Banco de Dados
  (File criado com status IsActive: true)
          │
          ▼
      [Concluído]
```
