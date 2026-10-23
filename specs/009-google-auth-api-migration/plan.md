# Implementation Plan: Autenticação via Google e Migração Completa para APIs (Sem Senhas)

**Branch**: `009-google-auth-api-migration` | **Date**: 2026-10-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/009-google-auth-api-migration/spec.md`

## Summary

Substituir integralmente a autenticação baseada em e-mail/senha do Docffy por autenticação via conta Google verificada (OIDC ID Token) com sessão stateless baseada em token Bearer (JWT), eliminando o armazenamento de senhas, tokens de recuperação de senha e fluxos manuais de verificação de e-mail. Implementar o vínculo automático por e-mail para contas já existentes e o onboarding de novos criadores em 2 etapas (Etapa 1: login Google cria/autentica `User`; Etapa 2: `POST /api/account/onboarding` valida CPF, maioridade 18+, telefone e termos para criar o `Creator`). Migrar 100% dos handlers HTTP de todos os módulos (`auth`, `account`, `library`, `sales`, `delivery`, `subscription` e `shared`) de renderização SSR (`pkg/template` + redirects/flash messages) para respostas estruturadas de API em JSON.

## Technical Context

**Language/Version**: Go 1.23.4  
**Primary Dependencies**: Chi Router (`github.com/go-chi/chi/v5`), GORM (`gorm.io/gorm`), AWS SDK v2 S3, Stripe Go SDK v76, UniPDF/pdfcpu, Go Playground Validator v10, Standard Library `crypto/hmac` + `crypto/sha256` + `encoding/json`  
**Storage**: PostgreSQL (via Docker localmente e em produção) / SQLite (em memória para testes isolados) + AWS S3 (arquivos e capas)  
**Testing**: `go test -v -race ./...`, `github.com/stretchr/testify` (TDD estrito com Table-Driven Tests)  
**Target Platform**: Linux Server / Docker Container / Heroku  
**Project Type**: Web Service / REST API Backend (Go)  
**Performance Goals**: Validação de token JWT em < 1ms; autenticação Google + emissão de JWT em < 500ms; respostas de endpoints JSON em < 200ms p95  
**Constraints**: Zero armazenamento de senhas/hashes no banco; preservação estrita da arquitetura modular (`handler` → `service` → `repository`) e comunicação entre módulos apenas por IDs primitivos (`uint`)  
**Scale/Scope**: 7 módulos internos (`auth`, `account`, `library`, `sales`, `delivery`, `subscription`, `shared`) e ponto de entrada `cmd/web/main.go`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Content Protection (`PASS`)**: O fluxo de entrega digital (`internal/delivery`) mantém a aplicação obrigatória de marca d'água dinâmica em PDF com identificação do comprador, limite de downloads e expiração temporal antes de servir o binário em `GET /purchase/download/{hash_id}?file_id=...`, convertendo apenas as respostas de listagem/erro de HTML para JSON.
- **II. Secure Payments (`PASS`)**: Os fluxos de pagamento e webhooks Stripe (`internal/sales` e `internal/subscription`) permanecem com verificação de assinatura, idempotência e ciclo de vida de transações inalterados.
- **III. Data Validation (`PASS`)**: As validações server-side de CPF (checksum + Receita Federal em produção), maioridade (18+ anos), formato de e-mail, Meta Pixel ID e sanitização de inputs são integralmente preservadas no `POST /api/account/onboarding` e nos demais endpoints JSON.
- **IV. Testability (`PASS`)**: Todos os serviços (`GoogleTokenVerifier`, `TokenService`, `UserService`, `CreatorService`) e handlers HTTP serão implementados/atualizados seguindo TDD estrito (`*_test.go` antes da implementação) com cobertura >= 80% na lógica de negócio.
- **V. Modularity (`PASS`)**: A divisão em módulos de domínio (`auth`, `account`, `library`, `sales`, `delivery`, `subscription`) e o isolamento via IDs primitivos (`uint`) são mantidos integralmente.
- **Nota de Governança / Stack**: A transição de SSR para APIs descontinua o uso de templates HTML renderizados pelo backend (`pkg/template` / `web/views`) em favor de contratos REST JSON consumidos por clientes externos.

## Project Structure

### Documentation (this feature)

```text
specs/009-google-auth-api-migration/
├── plan.md              # Este arquivo (Plano de Implementação)
├── research.md          # Decisões arquiteturais e técnicas (Fase 0)
├── data-model.md        # Modelos de dados, DTOs e máquina de estados (Fase 1)
├── quickstart.md        # Guia de execução e validação da API (Fase 1)
├── contracts/           # Contratos REST JSON de todos os módulos (Fase 1)
│   └── api.md
└── tasks.md             # Tarefas ordenadas por dependência (/speckit-tasks - Fase 2)
```

### Source Code (repository root)

```text
cmd/
└── web/
    └── main.go                                      # Registro de rotas da API, middlewares JWT/Onboarding e injeção de dependências

internal/
├── auth/
│   ├── model/
│   │   └── user.go                                  # User com GoogleID e sem Password / tokens de reset / confirmação manual
│   ├── repository/
│   │   ├── user_repository.go                       # Busca por GoogleID, Email e ID
│   │   └── user_repository_test.go
│   ├── service/
│   │   ├── dto.go                                   # GoogleAuthInput, GoogleIdentity, AuthResponse
│   │   ├── google_verifier.go                       # Interface GoogleTokenVerifier + implementação OIDC Google
│   │   ├── token_service.go                         # Emissão e validação de JWT Bearer stateless
│   │   ├── user_service.go                          # AuthenticateWithGoogle (vínculo por e-mail + criação Etapa 1)
│   │   └── user_service_test.go
│   └── handler/
│       ├── auth_handler.go                          # GoogleLogin, Me, Logout (JSON)
│       ├── auth_handler_test.go
│       └── middleware/
│           ├── authorizer.go                        # AuthMiddleware via Bearer JWT (retorna 401 JSON)
│           └── authorizer_test.go
├── account/
│   ├── model/
│   │   └── creator.go
│   ├── repository/
│   │   ├── creator_gorm.go
│   │   └── dashboard_repository.go
│   ├── service/
│   │   ├── creator_service.go                       # CompleteOnboarding / CreateCreatorForUser (sem senha)
│   │   ├── dto.go                                   # InputOnboardingCreator
│   │   └── stripe_connect_service.go
│   └── handler/
│       ├── creator_handler.go                       # POST /api/account/onboarding (Etapa 2 em JSON)
│       ├── dashboard_handler.go                     # GET /dashboard em JSON
│       ├── settings_handler.go                      # GET/POST /settings/* em JSON
│       ├── stripe_connect_handler.go                # GET /stripe-connect/* em JSON
│       └── middleware/
│           ├── creator_onboarding.go                # Bloqueia rotas de criador se Creator não existir (403 JSON)
│           └── stripe_onboarding.go                 # Retorna 403 JSON quando Stripe Connect pendente
├── library/
│   └── handler/
│       ├── ebook_handler.go                         # CRUD e listagem de ebooks em JSON
│       ├── file_handler.go                          # Upload, listagem, atualização e exclusão de arquivos em JSON
│       └── sales_page_handler.go                    # Dados públicos e preview da página de vendas em JSON
├── sales/
│   └── handler/
│       ├── checkout_handler.go                      # GET /checkout/{id} e GET /purchase/success em JSON
│       ├── client_handler.go                        # Listagem e atualização de clientes em JSON + export CSV
│       ├── purchase_handler.go                      # Criação de purchase em JSON
│       ├── purchase_sales_handler.go                # Listagem de vendas, bloqueio/desbloqueio e reenvio em JSON
│       └── transaction_handler.go                   # Listagem, detalhes e reenvio de transação em JSON
├── delivery/
│   └── handler/
│       └── download_handler.go                      # Metadados/erros de download em JSON + stream PDF com marca d'água
└── shared/
    ├── handler/
    │   ├── home.go                                  # Status/health da API em JSON
    │   └── error.go                                 # Respostas padronizadas de erro em JSON
    └── web/
        └── json_response.go                         # Helpers para serialização de respostas e erros JSON
```

**Structure Decision**: Mantida a estrutura monolítica modular Go do repositório (`cmd/web` + `internal/<domínio>/{model,repository,service,handler}`), substituindo a camada de apresentação SSR nos pacotes `handler/` e `middleware/` por respostas JSON padronizadas e autenticação stateless via JWT Bearer.

## Complexity Tracking

> Nenhuma violação aos princípios da Constituição.
