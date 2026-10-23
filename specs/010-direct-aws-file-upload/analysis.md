# Análise Técnica Completa: Upload Direto para AWS S3 com URLs Pré-assinadas

**Feature**: `010-direct-aws-file-upload`  
**Data**: 2026-10-23  
**Status**: Análise Prévia Concluída (Analysis Before Modification)

---

## 1. Problema Reportado e Contexto

### Descrição do Problema
Atualmente, no endpoint `POST /file/upload` (`internal/library/handler/file_handler.go`), o upload de arquivos digitais (ex.: PDFs de até 100MB) é enviado integralmente pelo navegador para o servidor backend Go através de requisições `multipart/form-data`. O backend então faz o buffer/leitura em memória/disco temporário e retransmite os dados para o Amazon S3 através do método `s3Storage.UploadFile`.

### Sintomas e Impactos
1. **Sobrecarga Dupla de Rede (Double-hop)**: Os bytes do arquivo trafegam duas vezes pela rede: `Navegador -> Backend Go -> AWS S3`.
2. **Consumo de Memória e CPU**: O parsing de multipart form de 100MB consome threads e memória do processo Go, podendo degradar o atendimento a outras requisições HTTP críticas (como checkout e autenticação).
3. **Risco Crítico de Timeout**: Plataformas de hospedagem (como o Heroku) possuem timeout rígido de 30 segundos para conexões HTTP. O upload de um arquivo de 50MB a 100MB em conexões lentas de usuários estoura frequentemente esse limite, gerando erros 504/H12.
4. **Acoplamento Desnecessário**: O servidor de aplicação atua como mero intermediário de binários pesados, quando sua função primordial deve ser apenas governança de autorização e persistência de metadados.

---

## 2. Mapeamento de Componentes Afetados

| Componente | Arquivo | Responsabilidade | Regra Aplicável |
|---|---|---|---|
| **Storage (Infra)** | `pkg/storage/s3.go` | Interface `S3Storage`: adicionar `GeneratePresignedPutURL` utilizando `PresignPutObject` do AWS SDK v2 | `archictecture-rules.mdc` |
| **Mocks de Infra** | `internal/mocks/mock_s3_storage.go` | Mock de `S3Storage` para testes unitários | `test-rule.mdc` |
| **DTOs de Domínio** | `internal/library/service/dto.go` | `PresignedUploadRequest`, `PresignedUploadResponse`, `CompleteUploadRequest`, `CompleteUploadResponse` | `service-rule.mdc` |
| **Serviço de Arquivos** | `internal/library/service/file_service.go` | Implementação de `GenerateUploadURL` e `CompleteUpload` | `service-rule.mdc` |
| **Testes de Serviço** | `internal/library/service/file_service_test.go` | Testes unitários com validação de namespace, tamanho e formato | `test-rule.mdc` |
| **Mocks de Serviço** | `internal/mocks/mock_file_service.go` | Mock de `FileService` atualizado | `test-rule.mdc` |
| **Handler HTTP** | `internal/library/handler/file_handler.go` | Endpoints `FilePresignUpload` (`POST /file/upload-url`) e `FileCompleteUpload` (`POST /file/complete-upload`) | `handler-rule.mdc` |
| **Testes de Handler** | `internal/library/handler/file_handler_test.go` | Testes cobrindo 200, 201, 400, 401 e 403 em JSON | `test-rule.mdc` |
| **Roteamento** | `cmd/web/main.go` | Registro das novas rotas no roteador Chi | `archictecture-rules.mdc` |
| **Helper Frontend** | `frontend/src/lib/s3-upload.ts` | Função `uploadFileToS3(uploadUrl, file, contentType)` | `ui-design-rules` |
| **Página Frontend** | `frontend/src/pages/files/FileLibraryPage.tsx` | Substituição do `FormData` pelo pipeline em 3 etapas com feedback visual | `ui-design-rules` |

---

## 3. Levantamento de Requisitos e Casos de Uso

### User Story 1 - Obtenção de URL Pré-assinada (P1)
O criador autenticado envia metadados básicos (`file_name`, `file_size`, `content_type`). O backend valida tamanho (<= 100MB), tipo de arquivo (PDF) e gera uma URL S3 pré-assinada com método `PUT` e chave estruturada `files/{creator_id}/{unique_prefix}-{safe_filename}`.

### User Story 2 - Upload Direto do Navegador para a AWS S3 (P1)
O frontend dispara uma requisição `PUT` diretamente para a URL pré-assinada do S3 com o binário no corpo da requisição e o `Content-Type` acordado. A AWS processa o upload e retorna `200 OK`.

### User Story 3 - Confirmação e Registro de Metadados (P1)
O frontend notifica o backend via `POST /file/complete-upload` com a chave do arquivo e metadados. O backend valida obrigatoriamente se a chave inicia com `files/{creator_id}/`, impedindo adulteração de arquivos de terceiros, persiste a entidade `File` no banco e retorna `201 Created`.

### User Story 4 - Experiência Visual no Frontend SPA (P2)
A tela de biblioteca de arquivos substitui o carregamento monolítico por etapas visuais amigáveis ("Gerando link seguro...", "Enviando para a AWS...", "Finalizando...") com tratamento de erros por Toast.

---

## 4. Análise Técnica da Solução (Go + React)

### 4.1 Interface `S3Storage` (`pkg/storage/s3.go`)
```go
type S3Storage interface {
    UploadFile(file *multipart.FileHeader, key, cacheControl string) (string, error)
    DeleteFile(key string) error
    GenerateDownloadLink(key string) string
    GenerateDownloadLinkWithExpiration(key string, expirationSeconds int) string
    GeneratePreviewLinkWithExpiration(key, contentType string, expirationSeconds int) string
    // Novo método para upload direto
    GeneratePresignedPutURL(ctx context.Context, key, contentType string, expirationSeconds int) (string, error)
}
```

### 4.2 DTOs (`internal/library/service/dto.go`)
```go
type PresignedUploadRequest struct {
    FileName    string `json:"file_name" validate:"required"`
    FileSize    int64  `json:"file_size" validate:"required,gt=0,lte=104857600"`
    ContentType string `json:"content_type" validate:"required"`
}

type PresignedUploadResponse struct {
    Success   bool              `json:"success"`
    UploadURL string            `json:"upload_url"`
    Key       string            `json:"key"`
    Method    string            `json:"method"`
    Headers   map[string]string `json:"headers,omitempty"`
    ExpiresIn int               `json:"expires_in"`
}

type CompleteUploadRequest struct {
    Key          string `json:"key" validate:"required"`
    Name         string `json:"name" validate:"required"`
    Description  string `json:"description"`
    OriginalName string `json:"original_name" validate:"required"`
    Size         int64  `json:"size" validate:"required,gt=0"`
    ContentType  string `json:"content_type"`
}
```

### 4.3 Pseudocódigo da Lógica Principal (`FileService`)
```go
func (s *fileService) GenerateUploadURL(ctx context.Context, req PresignedUploadRequest, creatorID uint) (*PresignedUploadResponse, error) {
    // 1. Validar tamanho máximo (100MB)
    if req.FileSize <= 0 || req.FileSize > 100<<20 {
        return nil, errors.New("tamanho de arquivo inválido ou excede o limite de 100MB")
    }

    // 2. Validar extensão/tipo (apenas PDF)
    ext := strings.ToLower(filepath.Ext(req.FileName))
    if ext != ".pdf" {
        return nil, errors.New("apenas arquivos no formato PDF são permitidos")
    }

    // 3. Gerar chave única protegida por namespace do criador
    uniqueID := s.generateUniqueID()
    safeName := sanitizeFileName(req.FileName)
    s3Key := fmt.Sprintf("files/%d/%s-%s", creatorID, uniqueID, safeName)

    // 4. Gerar URL pré-assinada no S3 (15 minutos de validade)
    expiration := 15 * 60 // 900s
    uploadURL, err := s.s3Storage.GeneratePresignedPutURL(ctx, s3Key, req.ContentType, expiration)
    if err != nil {
        return nil, fmt.Errorf("erro ao gerar URL pré-assinada: %w", err)
    }

    return &PresignedUploadResponse{
        Success:   true,
        UploadURL: uploadURL,
        Key:       s3Key,
        Method:    "PUT",
        Headers:   map[string]string{"Content-Type": req.ContentType},
        ExpiresIn: expiration,
    }, nil
}

func (s *fileService) CompleteUpload(ctx context.Context, req CompleteUploadRequest, creatorID uint) (*librarymodel.File, error) {
    // 1. Validação de isolamento de namespace (Segurança Multi-Tenant)
    expectedPrefix := fmt.Sprintf("files/%d/", creatorID)
    if !strings.HasPrefix(req.Key, expectedPrefix) {
        return nil, errors.New("acesso não autorizado: chave do arquivo não pertence ao criador")
    }

    // 2. Construir entidade de domínio
    fileType := s.getFileType(filepath.Ext(req.OriginalName))
    s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", config.AppConfig.S3BucketName, config.AppConfig.S3Region, req.Key)
    
    fileModel := librarymodel.NewFile(
        req.Name,
        req.OriginalName,
        req.Description,
        fileType,
        req.Key,
        s3URL,
        req.Size,
        creatorID,
    )

    // 3. Salvar no repositório
    if err := s.fileRepository.Create(fileModel); err != nil {
        return nil, fmt.Errorf("erro ao persistir arquivo: %w", err)
    }

    return fileModel, nil
}
```

---

## 5. Estratégia de Testes (TDD Rigoroso)

Os testes devem ser criados antes da implementação e executados via `make test`:

1. **Testes Unitários de `FileService`** (`internal/library/service/file_service_test.go`):
   - `TestGenerateUploadURL_Success`: valida retorno de URL e chave com prefixo `files/<creatorID>/`.
   - `TestGenerateUploadURL_SizeExceeded`: valida erro quando tamanho > 100MB.
   - `TestGenerateUploadURL_InvalidExtension`: valida recusa de extensões não permitidas (.exe, .zip, etc.).
   - `TestCompleteUpload_Success`: valida persistência correta e atributos do modelo retornado.
   - `TestCompleteUpload_NamespaceMismatch`: valida recusa imediata quando a chave aponta para outro `creatorID` (ex: `files/999/arquivo.pdf` para `creatorID = 1`).

2. **Testes Unitários de `FileHandler`** (`internal/library/handler/file_handler_test.go`):
   - `TestFilePresignUpload_Success`: status 200 e payload JSON válido.
   - `TestFilePresignUpload_Unauthorized`: status 401 quando sem token ou 403 sem perfil de criador.
   - `TestFileCompleteUpload_Success`: status 201 com entidade `file` criada.
   - `TestFileCompleteUpload_Forbidden`: status 403 com código semântico quando chave violada.

---

## 6. Análise de Segurança

1. **Prevenção de Namespace Hijacking / Path Traversal**:
   - A chave do arquivo no S3 é calculada e verificada com base no `creator.ID` extraído diretamente da sessão autenticada (JWT Bearer validado), impedindo qualquer tentativa de injeção de parâmetros (`files/../outro_criador`).
2. **Tempo de Expiração Curto (Short-lived SigV4)**:
   - A URL pré-assinada possui validade estrita de 15 minutos (900 segundos). Após esse período, qualquer tentativa de upload é rejeitada pela AWS S3.
3. **Restrição de Verbo HTTP e Assinatura Criptográfica**:
   - A URL pré-assinada com `PresignPutObject` só permite o método HTTP `PUT`. Tentativas de `GET`, `POST` ou `DELETE` com essa URL são sumariamente bloqueadas pela AWS.
4. **Validação de Content-Type Vinculado à Assinatura**:
   - O `ContentType` é incluído na assinatura SigV4. Se o cliente tentar enviar um payload com tipo MIME diferente do assinado, o S3 rejeitará a requisição com `403 SignatureDoesNotMatch`.
5. **CORS no Bucket S3**:
   - O bucket S3 deve aceitar preflight `OPTIONS` apenas com os métodos permitidos (`PUT`, `GET`, `HEAD`) e cabeçalhos restritos.
