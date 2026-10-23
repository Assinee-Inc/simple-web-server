# Implementation Plan: Upload Direto para AWS S3 com URLs Pré-assinadas

**Branch**: `direct_aws_file_upload` | **Date**: 2026-10-23 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/010-direct-aws-file-upload/spec.md`

## Summary

Desacoplar o upload de arquivos binários do servidor backend Go, implementando a geração de URLs pré-assinadas de upload (`PUT`) no AWS S3 com assinatura SigV4 e isolamento multi-tenant por criador (`files/{creator_id}/{unique_prefix}-{safe_filename}`). O frontend web (React SPA) passa a solicitar a URL pré-assinada (`POST /file/upload-url`), realizar o upload do binário diretamente para a AWS S3 via HTTP `PUT` (aliviando 100% da largura de banda, memória e CPU do servidor de aplicação) e, em seguida, confirmar os metadados no backend (`POST /file/complete-upload`), validando a chave e registrando o arquivo no banco de dados.

## Technical Context

**Language/Version**: Go 1.23.4 (Backend) / TypeScript 5.x + React 18 + Vite (Frontend)  
**Primary Dependencies**: Chi Router (`github.com/go-chi/chi/v5`), AWS SDK v2 S3 (`github.com/aws/aws-sdk-go-v2/service/s3`), GORM (`gorm.io/gorm`), Tailwind CSS + Radix UI / Shadcn  
**Storage**: AWS S3 (armazenamento de arquivos e pré-assinatura de URLs) + PostgreSQL (registro de metadados da tabela `files`)  
**Testing**: `go test -v -race ./...` (TDD estrito no backend via `golang-tdd-dev`), Jest / Cypress (testes de frontend e e2e)  
**Target Platform**: Linux Container / Heroku / AWS S3  
**Project Type**: Web Application (Go REST API Backend + React SPA Frontend)  
**Performance Goals**: Emissão de URL pré-assinada em < 150ms; confirmação de metadados em < 150ms; 0 bytes de dados binários pesados processados no servidor backend  
**Constraints**: Limite máximo de arquivo de 100MB; isolamento obrigatório de namespace por criador (`files/{creator_id}/...`); conformidade com CORS no bucket S3; preservação da arquitetura modular (`handler` → `service` → `repository`)  
**Scale/Scope**: Módulo `internal/library` (handler, service e testes), camada de infraestrutura `pkg/storage` (S3 presign PUT), rotas em `cmd/web/main.go` e página frontend `frontend/src/pages/files/FileLibraryPage.tsx`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Content Protection (`PASS`)**: O isolamento de arquivos por criador (`files/{creator_id}/...`) é mantido estritamente. O download de arquivos para compradores através do módulo `delivery` continua aplicando marca d'água dinâmica e controle de acessos normalmente.
- **II. Secure Payments (`PASS`)**: Não afeta pagamentos nem webhooks de cobrança.
- **III. Data Validation (`PASS`)**: Validação de tipo de arquivo (extensões e content-type permitidos), tamanho máximo (<= 100MB) e verificação de namespace da chave S3 no endpoint de confirmação são aplicadas no backend.
- **IV. Testability (`PASS`)**: Toda a lógica Go (serviço de arquivos, handlers, mocks de S3) será implementada seguindo TDD rigoroso (`*_test.go` antes da implementação) pelo subagente `golang-tdd-dev`.
- **V. Modularity (`PASS`)**: A divisão em camadas (`handler` → `service` → `repository`) é mantida sem vazamento de regras de persistência para o controller.

## Project Structure

### Documentation (this feature)

```text
specs/010-direct-aws-file-upload/
├── plan.md              # Este arquivo (Plano de Implementação)
├── spec.md              # Especificação de requisitos e cenários de aceitação
├── research.md          # Decisões arquiteturais, CORS e AWS SDK Presign (Fase 0)
├── data-model.md        # DTOs, ciclo de vida e máquina de estados (Fase 1)
├── quickstart.md        # Guia de execução e testes cURL e frontend (Fase 1)
├── contracts/           # Contratos REST JSON de upload e confirmação (Fase 1)
│   └── upload-api.md
└── tasks.md             # Tarefas ordenadas (/speckit-tasks - Fase 2)
```

### Source Code (repository root)

```text
cmd/
└── web/
    └── main.go                                          # Registro das rotas /file/upload-url e /file/complete-upload

pkg/
└── storage/
    └── s3.go                                            # Adição de GeneratePresignedPutURL à interface S3Storage

internal/
├── library/
│   ├── service/
│   │   ├── dto.go                                       # DTOs PresignedUploadRequest, PresignedUploadResponse, CompleteUploadRequest
│   │   ├── file_service.go                              # Métodos GenerateUploadURL e CompleteUpload com validações
│   │   └── file_service_test.go                         # Testes unitários TDD de serviço de arquivo
│   └── handler/
│       ├── file_handler.go                              # Handlers FilePresignUpload e FileCompleteUpload
│       └── file_handler_test.go                         # Testes unitários TDD de handlers HTTP

frontend/
└── src/
    └── pages/
        └── files/
            └── FileLibraryPage.tsx                      # Atualização do formulário para fluxo em 2 etapas com upload direto
```

**Structure Decision**: A implementação estende os pacotes existentes em `pkg/storage` e `internal/library`, mantendo a arquitetura limpa sem criar módulos desnecessários. Toda a alteração em arquivos `.go` será delegada obrigatoriamente para o subagente especializado `golang-tdd-dev`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*Nenhuma violação identificada. O design utiliza os padrões nativos do AWS SDK Go v2 e endpoints REST padrão.*
