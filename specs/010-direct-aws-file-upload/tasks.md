# Tasks: Upload Direto para AWS S3 com URLs Pré-assinadas

**Input**: Design documents from `/specs/010-direct-aws-file-upload/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/upload-api.md, quickstart.md

**Tests**: Testes unitários e de integração estão incluídos e DEVEM ser escritos antes da implementação (TDD estrito conforme Princípio IV da Constituição, `plan.md` e regras de subagente `golang-tdd-dev`).

**Organization**: As tarefas estão agrupadas por User Story para permitir implementação e validação independentes de cada incremento.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode ser executada em paralelo (arquivos diferentes, sem dependência de tarefas incompletas)
- **[Story]**: User story à qual a tarefa pertence (`[US1]`, `[US2]`, `[US3]`, `[US4]`)
- Caminhos exatos de arquivos incluídos em todas as descrições

---

## Phase 1: Setup & Storage Infrastructure

**Purpose**: Extensão da camada de armazenamento AWS S3 (`pkg/storage`) para suportar a geração de URLs pré-assinadas com método HTTP PUT

- [x] T001 [P] Adicionar método `GeneratePresignedPutURL(ctx context.Context, key, contentType string, expirationSeconds int) (string, error)` à interface `S3Storage` em `pkg/storage/s3.go`
- [x] T002 Implementar `GeneratePresignedPutURL` em `s3Storage` (`pkg/storage/s3.go`) utilizando `s3.NewPresignClient(s.client).PresignPutObject` com expiração configurável
- [x] T003 [P] Atualizar mocks de armazenamento S3 (`internal/mocks/mock_s3_storage.go` e mocks de testes de library) adicionando `GeneratePresignedPutURL`

---

## Phase 2: Foundational (DTOs & Validation)

**Purpose**: Estruturas de dados e regras de validação para a solicitação de URL pré-assinada e confirmação de upload

- [x] T004 [P] Criar DTOs de requisição e resposta de upload pré-assinado (`PresignedUploadRequest`, `PresignedUploadResponse`, `CompleteUploadRequest`, `CompleteUploadResponse`) em `internal/library/service/dto.go`
- [x] T005 [P] Escrever testes unitários para validação de metadados de arquivo (extensão PDF, limite de tamanho <= 100MB e sanitização de nome) em `internal/library/service/file_validator_test.go`
- [x] T006 Implementar validações de tamanho máximo (100MB) e tipos de arquivo permitidos em `internal/library/service/file_service.go`

**Checkpoint**: Camada de armazenamento e DTOs prontos para implementação das User Stories.

---

## Phase 3: User Story 1 - Obtenção de URL Pré-assinada para Upload (Priority: P1) 🎯 MVP 1

**Goal**: Permitir que criadores autenticados solicitem uma URL pré-assinada (`POST /file/upload-url`) contendo assinatura SigV4 do AWS S3 (`PUT`) e chave restrita ao seu namespace (`files/{creator_id}/...`).

**Independent Test**: Autenticar como criador, enviar `POST /file/upload-url` com `file_name: "apostila.pdf"`, `file_size: 2048576` e `content_type: "application/pdf"`. O backend deve responder `200 OK` com `upload_url`, `method: "PUT"`, `key` no formato `files/{creator_id}/...` e tempo de expiração.

### Tests for User Story 1 ⚠️
> **NOTE: Escreva estes testes PRIMEIRO e garanta que falhem antes da implementação (delegar a `golang-tdd-dev`)**

- [x] T007 [P] [US1] Escrever testes unitários para `FileService.GenerateUploadURL` cobrindo validação de tamanho, tipo de arquivo, isolamento de chave por criador (`files/{creatorID}/...`) e retorno da URL pré-assinada em `internal/library/service/file_service_test.go`
- [x] T008 [P] [US1] Escrever testes unitários de handler para `POST /file/upload-url` cobrindo autorização, respostas `200 OK`, `400 Bad Request` e `401/403` em `internal/library/handler/file_handler_test.go`

### Implementation for User Story 1
- [x] T009 [US1] Adicionar `GenerateUploadURL(ctx context.Context, req PresignedUploadRequest, creatorID uint) (*PresignedUploadResponse, error)` à interface `FileService` e implementar em `internal/library/service/file_service.go`
- [x] T010 [US1] Atualizar `internal/mocks/mock_file_service.go` com o novo método `GenerateUploadURL`
- [x] T011 [US1] Implementar o endpoint `FilePresignUpload` (`POST /file/upload-url`) em `internal/library/handler/file_handler.go`
- [x] T012 [US1] Registrar a rota `r.Post("/file/upload-url", fileHandler.FilePresignUpload)` dentro do grupo autenticado de criadores em `cmd/web/main.go`

**Checkpoint**: User Story 1 concluída e testável independentemente via cURL/Postman.

---

## Phase 4: User Story 2 & 3 - Confirmação e Registro de Metadados do Arquivo (Priority: P1) 🎯 MVP 2

**Goal**: Validar a chave S3 enviada pelo cliente (`files/{creator_id}/...`), impedir qualquer manipulação multi-tenant e persistir o registro do arquivo no banco de dados após o sucesso do upload direto no S3.

**Independent Test**: Enviar `POST /file/complete-upload` com a `key` emitida na etapa anterior e metadados (`name`, `description`, `size`, `original_name`). Validar que o retorno é `201 Created` e que o arquivo aparece na listagem `GET /file`. Tentar confirmar com a `key` de outro criador e verificar retorno imediato `403 Forbidden`.

### Tests for User Stories 2 & 3 ⚠️
> **NOTE: Escreva estes testes PRIMEIRO e garanta que falhem antes da implementação (delegar a `golang-tdd-dev`)**

- [x] T013 [P] [US3] Escrever testes unitários para `FileService.CompleteUpload` cobrindo validação de prefixo do criador (`files/{creatorID}/...`), rejeição de chaves inválidas (`403`), criação da entidade `File` e persistência via `FileRepository` em `internal/library/service/file_service_test.go`
- [x] T014 [P] [US3] Escrever testes unitários de handler para `POST /file/complete-upload` cobrindo sucesso `201 Created`, erros de validação `400 Bad Request` e recusa por namespace `403 Forbidden` em `internal/library/handler/file_handler_test.go`

### Implementation for User Stories 2 & 3
- [x] T015 [US3] Adicionar `CompleteUpload(ctx context.Context, req CompleteUploadRequest, creatorID uint) (*librarymodel.File, error)` à interface `FileService` e implementar em `internal/library/service/file_service.go`
- [x] T016 [US3] Atualizar `internal/mocks/mock_file_service.go` com o método `CompleteUpload`
- [x] T017 [US3] Implementar o endpoint `FileCompleteUpload` (`POST /file/complete-upload`) em `internal/library/handler/file_handler.go`
- [x] T018 [US3] Registrar a rota `r.Post("/file/complete-upload", fileHandler.FileCompleteUpload)` no roteador em `cmd/web/main.go`

**Checkpoint**: Backend completo e validado com 100% de cobertura TDD nas rotas `/file/upload-url` e `/file/complete-upload`.

---

## Phase 5: User Story 4 - Atualização da Interface do Usuário (Frontend SPA) (Priority: P2)

**Goal**: Atualizar o formulário de upload de arquivos no React SPA (`FileLibraryPage.tsx`) para orquestrar a solicitação da URL pré-assinada, upload binário direto para a AWS S3 via `fetch(PUT)` e confirmação com o backend, exibindo estados visuais em tempo real.

**Independent Test**: Acessar a página `/files` na interface SPA, selecionar um PDF, clicar em enviar e verificar na aba Network do navegador que a requisição binária é enviada para o S3 e nenhuma carga pesada passa pelo servidor Go.

- [x] T019 [US4] Criar função helper de upload direto S3 em `frontend/src/lib/s3-upload.ts` (executa `fetch(uploadUrl, { method: "PUT", body: file, headers: { "Content-Type": file.type } })`)
- [x] T020 [US4] Atualizar `handleUpload` em `frontend/src/pages/files/FileLibraryPage.tsx` para executar o fluxo em 3 passos:
  1. `apiRequest("/file/upload-url", { method: "POST", body: JSON.stringify(...) })`
  2. Upload direto do binário para o S3 com a URL pré-assinada
  3. `apiRequest("/file/complete-upload", { method: "POST", body: JSON.stringify(...) })`
- [x] T021 [US4] Adicionar estados visuais no modal de upload ("Preparando...", "Enviando para a nuvem...", "Finalizando...") e tratamento de erros amigável em `frontend/src/pages/files/FileLibraryPage.tsx`

---

## Phase 6: Validação End-to-End & Documentação

**Purpose**: Testes integrados ponta a ponta e garantia de qualidade

- [x] T022 Executar bateria completa de testes automatizados do backend (`go test -v -race ./...`) e testes de frontend
- [x] T023 Validar fluxo manual completo de upload direto conforme roteiro de testes em `specs/010-direct-aws-file-upload/quickstart.md`
- [x] T024 Atualizar a coleção Postman (`specs/009-google-auth-api-migration/contracts/docffy_api.postman_collection.json` e `specs/010-direct-aws-file-upload/contracts/docffy_api.postman_collection.json`) com os novos endpoints `/file/upload-url` e `/file/complete-upload`
