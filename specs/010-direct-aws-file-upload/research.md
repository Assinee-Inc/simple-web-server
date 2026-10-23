# Phase 0 Research: Upload Direto para AWS S3 com URLs Pré-assinadas

**Feature**: `010-direct-aws-file-upload`  
**Date**: 2026-10-23

---

## 1. Geração de URLs Pré-assinadas com AWS SDK Go v2 (`PresignPutObject`)

- **Decision**: Adicionar o método `GeneratePresignedPutURL(ctx context.Context, key, contentType string, expirationSeconds int) (string, error)` à interface `storage.S3Storage` (`pkg/storage/s3.go`), utilizando `s3.NewPresignClient(s.client).PresignPutObject`.
- **Rationale**:
  - O projeto já utiliza o AWS SDK Go v2 (`github.com/aws/aws-sdk-go-v2/service/s3`).
  - O cliente de pré-assinatura (`s3.NewPresignClient`) já é utilizado atualmente em `GenerateDownloadLinkWithExpiration` e `GeneratePreviewLinkWithExpiration` para chamadas `PresignGetObject`.
  - Configurar `PresignPutObject` especificando `Bucket`, `Key` e `ContentType` garante que a assinatura SigV4 trave o tipo de conteúdo do arquivo, impedindo que um cliente envie binários disfarçados ou altere o caminho no bucket.
- **Implementation Pattern**:
  ```go
  func (s *s3Storage) GeneratePresignedPutURL(ctx context.Context, key, contentType string, expirationSeconds int) (string, error) {
      presigner := s3.NewPresignClient(s.client)
      params := &s3.PutObjectInput{
          Bucket:      aws.String(s.bucket),
          Key:         aws.String(key),
          ContentType: aws.String(contentType),
      }
      presignedReq, err := presigner.PresignPutObject(ctx, params, func(opts *s3.PresignOptions) {
          opts.Expires = time.Duration(expirationSeconds) * time.Second
      })
      if err != nil {
          return "", fmt.Errorf("erro ao gerar presigned PUT URL: %w", err)
      }
      return presignedReq.URL, nil
  }
  ```
- **Alternatives considered**:
  - *S3 POST com Presigned Policy/Form-Data*: Mais complexo de consumir no frontend (exige múltiplos campos de formulário como policy, signature, credentials). O `PUT` direto é o padrão moderno da AWS para uploads binários diretos de SPAs via `fetch(url, { method: "PUT", body: file })`.
  - *Multi-part Upload S3*: Indicado apenas para arquivos > 100MB. Como o limite do Docffy é de 100MB, um único PUT com URL pré-assinada é ideal, mantendo conformidade com o princípio KISS.

---

## 2. Requisitos de CORS no Amazon S3

- **Decision**: O bucket S3 deve permitir solicitações de origem cruzada (CORS) para o domínio da aplicação frontend (e `localhost` em desenvolvimento).
- **Rationale**:
  - Quando o navegador faz `fetch(presignedUrl, { method: "PUT", body: file, headers: { "Content-Type": ... } })`, ele dispara automaticamente uma requisição preflight `OPTIONS` para o endpoint do S3.
  - Para permitir o upload direto sem erros de CORS no navegador, a regra CORS do bucket S3 deve conter:
    - `AllowedHeaders`: `["*"]` (ou `["Content-Type", "x-amz-*"]`)
    - `AllowedMethods`: `["GET", "PUT", "HEAD"]`
    - `AllowedOrigins`: origens da aplicação (ex.: `http://localhost:5173`, `http://localhost:8080`, `https://*.docffy.com`)
    - `ExposeHeaders`: `["ETag"]`
- **Alternatives considered**:
  - *Proxy reverso*: Anularia o objetivo principal de aliviar o servidor backend.

---

## 3. Isolamento de Namespace por Criador e Prevenção de Injeção de Chave

- **Decision**:
  - A chave S3 (`s3_key`) deve ser construída exclusivamente no backend, com a estrutura:
    `files/{creator_id}/{unique_prefix}-{safe_filename}`.
  - No endpoint de confirmação de upload (`POST /file/complete-upload`), o backend deve validar rigorosamente se a chave fornecida pelo cliente inicia com o prefixo do criador autenticado:
    `strings.HasPrefix(req.Key, fmt.Sprintf("files/%d/", creator.ID))`
- **Rationale**:
  - Previne ataques de hijacking ou apropriação indevida onde um criador A poderia tentar registrar um arquivo de um criador B no seu catálogo.
  - Garante que a exclusão de arquivos ou operações de limpeza não afetem objetos de terceiros.
- **Alternatives considered**:
  - *Confiar no frontend para enviar a chave arbitrária*: Rejeitado por falha de segurança multi-tenant.

---

## 4. Estratégia de Endpoints e Fluxo de Upload

- **Decision**: Adotar o fluxo em duas etapas no módulo `library`:
  1. `POST /file/upload-url` (ou `/api/files/upload-url`):
     - Payload: `{"file_name": "livro.pdf", "file_size": 1048576, "content_type": "application/pdf"}`
     - Resposta: `{"success": true, "upload_url": "https://...", "key": "files/1/...", "method": "PUT", "expires_in": 900}`
  2. Frontend faz `PUT` direto para `upload_url`.
  3. `POST /file/complete-upload` (ou `/api/files/complete-upload`):
     - Payload: `{"key": "files/1/...", "name": "Nome", "description": "...", "original_name": "livro.pdf", "size": 1048576, "content_type": "application/pdf"}`
     - Backend valida chave, insere no banco via repositório e retorna o `File` criado.
- **Rationale**:
  - Separação clara de responsabilidades: autorização/assinatura -> transferência de dados pesados -> persistência no catálogo.
  - Manter o endpoint legado `POST /file/upload` intacto para preservar compatibilidade com testes legados ou chamadas internas até descontinuação formal.

---

## 5. Implementação no Frontend (`FileLibraryPage.tsx`)

- **Decision**: Atualizar `FileLibraryPage.tsx` para substituir o envio de `FormData` para o backend pelo pipeline:
  1. Chamar `apiRequest("/file/upload-url", { method: "POST", body: JSON.stringify({ file_name, file_size, content_type }) })`.
  2. Executar `fetch(upload_url, { method: "PUT", body: selectedFile, headers: { "Content-Type": contentType } })`.
  3. Chamar `apiRequest("/file/complete-upload", { method: "POST", body: JSON.stringify({ ... }) })`.
  4. Exibir feedback ao usuário durante cada etapa.
- **Rationale**:
  - Implementação enxuta, sem dependências adicionais no frontend.
  - Compatível com a infraestrutura existente de `apiRequest` (`frontend/src/lib/api.ts`).
