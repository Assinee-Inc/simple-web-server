# Tasks: Autenticação via Google e Migração Completa para APIs (Sem Senhas)

**Input**: Design documents from `/specs/009-google-auth-api-migration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/api.md, quickstart.md

**Tests**: Testes unitários e de integração estão incluídos e DEVEM ser escritos antes da implementação (TDD estrito conforme Princípio IV da Constituição e `plan.md`).

**Organization**: As tarefas estão agrupadas por User Story para permitir implementação e validação independentes de cada incremento.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode ser executada em paralelo (arquivos diferentes, sem dependência de tarefas incompletas)
- **[Story]**: User story à qual a tarefa pertence (`[US1]`, `[US2]`, `[US3]`)
- Caminhos exatos de arquivos incluídos em todas as descrições

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Configuração de ambiente e utilitários compartilhados de resposta JSON para toda a API

- [ ] T001 Adicionar configurações `GoogleClientID`, `JWTSecret` e `JWTExpirationHours` em `internal/config/config.go` e documentar variáveis em `env.template`
- [ ] T002 [P] Criar testes unitários para helpers de serialização JSON e envelopes de erro padronizados em `internal/shared/web/json_response_test.go`
- [ ] T003 [P] Implementar helpers de resposta JSON (`RespondJSON`, `RespondError`, `RespondValidationError`) em `internal/shared/web/json_response.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Infraestrutura central de identidade sem senha, emissão/validação de JWT stateless e extração de contexto de autenticação que bloqueiam todas as User Stories

**⚠️ CRITICAL**: Nenhuma User Story pode ser iniciada antes da conclusão desta fase

- [ ] T004 Atualizar testes de domínio do modelo `User` (com `GoogleID` e sem `Password`, `PasswordResetToken`, `PasswordResetAt`, `EmailVerificationToken`, `CSRFToken`, `SessionToken`) em `internal/auth/model/user_test.go`
- [ ] T005 Refatorar a entidade `User` removendo campos de senha/recuperação/verificação manual e adicionando `GoogleID` (`uniqueIndex`) em `internal/auth/model/user.go` e atualizar migração em `pkg/database/database.go`
- [ ] T006 [P] Escrever testes unitários para emissão, validação de assinatura HMAC-SHA256 e expiração de JWT Bearer em `internal/auth/service/token_service_test.go`
- [ ] T007 [P] Implementar `TokenService` stateless (geração e parsing de JWT HS256 com claims `sub`, `uid`, `email`, `iat`, `exp`) em `internal/auth/service/token_service.go`
- [ ] T008 [P] Escrever testes unitários para validação de ID Token do Google (`aud`, `iss`, `exp`, `email_verified`) em `internal/auth/service/google_verifier_test.go`
- [ ] T009 [P] Implementar interface `GoogleTokenVerifier` e verificador OIDC Google em `internal/auth/service/google_verifier.go`
- [ ] T010 Escrever testes para `AuthMiddleware` (extração de `Authorization: Bearer <token>`, injeção de `User` no `context.Context` e retorno `401 Unauthorized` em JSON) em `internal/auth/handler/middleware/authorizer_test.go`
- [ ] T011 Atualizar `AuthMiddleware` e helpers de contexto de usuário autenticado para operar via `TokenService` (JWT Bearer) e responder em JSON em `internal/auth/handler/middleware/authorizer.go`

**Checkpoint**: Fundação pronta — modelo `User` sem senha, `TokenService`, `GoogleTokenVerifier` e `AuthMiddleware` operacionais

---

## Phase 3: User Story 1 - Autenticação via Conta Google Sem Senhas (Priority: P1) 🎯 MVP

**Goal**: Autenticar usuários existentes (com vínculo automático por e-mail) e novos usuários exclusivamente via conta Google verificada, emitindo token Bearer JWT stateless e removendo todos os fluxos de senha e confirmação manual de e-mail.

**Independent Test**: Enviar `POST /api/auth/google` com token Google de um e-mail existente e de um novo e-mail, verificar o retorno do `access_token` JWT, consultar `GET /api/auth/me` com `Authorization: Bearer <token>`, executar `POST /api/auth/logout` e confirmar que rotas legadas (`/login`, `/forget-password`, `/reset-password`) retornam `404 Not Found`.

### Tests for User Story 1 ⚠️

> **NOTE: Escreva estes testes PRIMEIRO e garanta que falhem antes da implementação**

- [ ] T012 [P] [US1] Atualizar testes de repositório para busca por `GoogleID` e `Email` sem campos de senha em `internal/auth/repository/user_repository_test.go`
- [ ] T013 [P] [US1] Atualizar testes unitários de `UserService` para cobrir `AuthenticateWithGoogle` (novo usuário, vínculo automático por e-mail existente, rejeição de `email_verified = false`) em `internal/auth/service/user_service_test.go`
- [ ] T014 [P] [US1] Atualizar testes de `AuthHandler` para cobrir `POST /api/auth/google`, `GET /api/auth/me` e `POST /api/auth/logout` em JSON em `internal/auth/handler/auth_handler_test.go`

### Implementation for User Story 1

- [ ] T015 [P] [US1] Definir DTOs de autenticação (`GoogleAuthInput`, `GoogleIdentity`, `AuthResponse`, `UserSummaryDTO`) em `internal/auth/model/dto.go` e `internal/auth/service/dto.go`
- [ ] T016 [US1] Atualizar `UserRepository` adicionando `FindByGoogleID` e removendo métodos legados de token de senha/verificação manual em `internal/auth/repository/user_repository.go` e `internal/mocks/mock_user_repository.go`
- [ ] T017 [US1] Refatorar `UserService` implementando `AuthenticateWithGoogle` (vínculo por e-mail existente ou criação na Etapa 1) e removendo métodos de senha/reset/confirmação em `internal/auth/service/user_service.go`, `internal/auth/service/user_validator.go` e `internal/mocks/mock_user_service.go`
- [ ] T018 [P] [US1] Remover métodos e templates de envio de e-mail de recuperação de senha e confirmação manual de conta em `internal/auth/service/email_service.go` e `internal/auth/service/email_service_test.go`
- [ ] T019 [US1] Refatorar `AuthHandler` para expor `GoogleLogin` (`POST /api/auth/google`), `Me` (`GET /api/auth/me`) e `Logout` (`POST /api/auth/logout`) em JSON, removendo handlers SSR de login/reset/confirmação em `internal/auth/handler/auth_handler.go`

**Checkpoint**: User Story 1 funcional e testável independentemente (MVP de autenticação Google sem senhas pronto)

---

## Phase 4: User Story 2 - Onboarding de Criador em Duas Etapas via API (Priority: P2)

**Goal**: Permitir que o usuário autenticado via Google (Etapa 1) conclua seu cadastro comercial como criador (Etapa 2) via `POST /api/account/onboarding` (validando CPF, maioridade 18+, telefone e aceite de termos) e bloquear o acesso às rotas exclusivas de criador enquanto o perfil `Creator` não estiver concluído.

**Independent Test**: Autenticar um novo usuário via Google (`onboarding_required: true`), verificar que o acesso a rotas de criador retorna `403 Forbidden` (`CREATOR_ONBOARDING_REQUIRED`), enviar os dados fiscais em `POST /api/account/onboarding` para receber `201 Created` e confirmar que o bloqueio de onboarding de criador deixa de ocorrer.

### Tests for User Story 2 ⚠️

- [ ] T020 [P] [US2] Escrever testes unitários para `CreatorOnboardingMiddleware` (bloqueio `403 JSON` com `CREATOR_ONBOARDING_REQUIRED` quando não há `Creator` e passagem com `Creator` no contexto quando existe) em `internal/account/handler/middleware/creator_onboarding_test.go`
- [ ] T021 [P] [US2] Escrever testes unitários para `CreatorService.OnboardCreator` (validação de CPF, maioridade 18+, aceite de termos, conflito de CPF duplicado e atualização de `TermsAcceptedAt`) em `internal/account/service/creator_service_test.go`
- [ ] T022 [P] [US2] Escrever testes de handler para `POST /api/account/onboarding` (`201 Created`, `400 Bad Request`, `409 Conflict`) em `internal/account/handler/creator_handler_test.go`

### Implementation for User Story 2

- [ ] T023 [P] [US2] Definir `InputOnboardingCreator` (sem campos `Password`/`PasswordConfirmation`) e DTOs de resposta de criador em `internal/account/service/dto.go` e `internal/account/model/dto.go`
- [ ] T024 [P] [US2] Implementar `CreatorOnboardingMiddleware` retornando `403 Forbidden` em JSON (`CREATOR_ONBOARDING_REQUIRED`) quando o usuário autenticado não possui registro `Creator` em `internal/account/handler/middleware/creator_onboarding.go`
- [ ] T025 [US2] Atualizar `CreatorService` e `CreatorValidator` para criar o `Creator` na Etapa 2 vinculado ao `UserID` autenticado (validando CPF, idade >= 18 anos, telefone e termos sem exigir senha) em `internal/account/service/creator_service.go`, `internal/account/service/creator_validator.go` e `internal/mocks/mock_creator_service.go`
- [ ] T026 [US2] Refatorar `CreatorHandler` substituindo o registro SSR por `OnboardCreator` (`POST /api/account/onboarding`) em JSON em `internal/account/handler/creator_handler.go`

**Checkpoint**: User Stories 1 e 2 funcionais e testáveis independentemente (fluxo completo de login Google + onboarding de criador em 2 etapas)

---

## Phase 5: User Story 3 - Migração Completa de Rotas SSR do Criador e Fluxos Públicos para APIs (Priority: P3)

**Goal**: Converter 100% dos handlers e middlewares restantes dos módulos `account`, `library`, `sales`, `delivery`, `subscription` e `shared` de renderização SSR (`pkg/template` + flash messages/redirects) para respostas estruturadas em JSON.

**Independent Test**: Executar chamadas HTTP para todos os endpoints de conta/dashboard/configurações/Stripe Connect, biblioteca (ebooks, arquivos, página de vendas), vendas (clientes, vendas, transações, checkout) e entrega (download protegido) e verificar que 100% das respostas retornam JSON estruturado (ou stream PDF/CSV em downloads/exportação com erros em JSON), sem uso de templates HTML.

### Tests for User Story 3 ⚠️

- [ ] T027 [P] [US3] Atualizar testes de `SettingsHandler`, `DashboardHandler` e `StripeConnectHandler` (incluindo `StripeOnboardingMiddleware` em modo JSON `403 STRIPE_ONBOARDING_REQUIRED`) em `internal/account/handler/settings_handler_test.go` e `internal/account/handler/middleware/stripe_onboarding_test.go`
- [ ] T028 [P] [US3] Atualizar testes de `EbookHandler`, `FileHandler` e `SalesPageHandler` para validar contratos JSON e códigos de status HTTP em `internal/library/handler/ebook_handler_test.go`, `internal/library/handler/ebook_file_management_test.go`, `internal/library/handler/file_handler_test.go`, `internal/library/handler/sales_page_handler_test.go` e `internal/library/handler/sales_page_test.go`
- [ ] T029 [P] [US3] Atualizar testes de `ClientHandler`, `PurchaseHandler`, `PurchaseSalesHandler`, `TransactionHandler` e `CheckoutHandler` para validar respostas JSON em `internal/sales/handler/client_handler_test.go`, `internal/sales/handler/purchase_handler_test.go`, `internal/sales/handler/purchase_sales_handler_test.go`, `internal/sales/handler/transaction_handler_test.go` e `internal/sales/handler/checkout_handler_test.go`
- [ ] T030 [P] [US3] Atualizar testes de `DownloadHandler` (metadados JSON sem `file_id`, stream PDF com marca d'água com `file_id` e erros `403`/`404`/`410` em JSON) e middlewares de `subscription` em `internal/delivery/handler/download_handler_test.go` e `internal/subscription/handler/middleware/trial_test.go`

### Implementation for User Story 3

- [ ] T031 [P] [US3] Atualizar `StripeOnboardingMiddleware` e middlewares de assinatura (`SubscriptionMiddleware`, `TrialMiddleware`) para responderem `401`/`403` em JSON em vez de redirects SSR em `internal/account/handler/middleware/stripe_onboarding.go`, `internal/subscription/handler/middleware/subscription.go` e `internal/subscription/handler/middleware/trial.go`
- [ ] T032 [P] [US3] Migrar `DashboardHandler`, `SettingsHandler` e `StripeConnectHandler` para retornarem JSON sem `TemplateRenderer` em `internal/account/handler/dashboard_handler.go`, `internal/account/handler/settings_handler.go` e `internal/account/handler/stripe_connect_handler.go`
- [ ] T033 [P] [US3] Migrar `EbookHandler`, `FileHandler` e `SalesPageHandler` para retornarem JSON (removendo views de formulário HTML e flash messages) em `internal/library/handler/ebook_handler.go`, `internal/library/handler/file_handler.go` e `internal/library/handler/sales_page_handler.go`
- [ ] T034 [P] [US3] Migrar `ClientHandler`, `PurchaseHandler`, `PurchaseSalesHandler`, `TransactionHandler` e `CheckoutHandler` para retornarem JSON (preservando exportação `text/csv` em `/client/export`) em `internal/sales/handler/client_handler.go`, `internal/sales/handler/purchase_handler.go`, `internal/sales/handler/purchase_sales_handler.go`, `internal/sales/handler/transaction_handler.go` e `internal/sales/handler/checkout_handler.go`
- [ ] T035 [P] [US3] Migrar `DownloadHandler` para retornar metadados e erros (`PAYMENT_PENDING`, `DOWNLOAD_EXPIRED`, `DOWNLOAD_LIMIT_EXCEEDED`) em JSON mantendo o stream PDF com marca d'água dinâmica em `internal/delivery/handler/download_handler.go`
- [ ] T036 [P] [US3] Migrar `HomeHandler` e `ErrorHandler` para retornarem status da API e erros HTTP (`404`, `405`, `500`) em JSON em `internal/shared/handler/home.go` e `internal/shared/handler/error.go`
- [ ] T037 [US3] Atualizar o ponto de entrada `cmd/web/main.go` registrando as novas rotas de API (`/api/auth/google`, `/api/auth/me`, `/api/auth/logout`, `/api/account/onboarding`), aplicando `AuthMiddleware` e `CreatorOnboardingMiddleware`, e removendo rotas legadas de senha/SSR e inicialização de `TemplateRenderer` em `cmd/web/main.go`

**Checkpoint**: Todas as User Stories (US1, US2 e US3) estão completas e 100% da aplicação opera via contratos de API estruturados

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Limpeza de código legado SSR/sessão, validação de suíte completa de testes e checagem de segurança

- [ ] T038 [P] Remover middlewares e utilitários SSR obsoletos (`AuthGuard` de redirecionamento HTML e helpers de flash message não utilizados) em `internal/auth/handler/middleware/guard.go` e `internal/shared/web/flash_message.go`
- [ ] T039 Executar a suíte completa de testes com detector de condição de corrida (`go test -v -race ./...`), formatação (`go fmt ./...`), verificação de segurança (`make security-check`) e validar os cenários do guia `specs/009-google-auth-api-migration/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Sem dependências — pode iniciar imediatamente.
- **Foundational (Phase 2)**: Depende de Setup (Phase 1) — **BLOQUEIA** todas as User Stories.
- **User Story 1 (Phase 3 - P1)**: Depende da conclusão da Phase 2 (Foundational).
- **User Story 2 (Phase 4 - P2)**: Depende da Phase 2 (Foundational) e integra-se com o modelo/serviço de usuário da US1.
- **User Story 3 (Phase 5 - P3)**: Depende da Phase 2 (Foundational) e utiliza os middlewares de autenticação (US1) e onboarding de criador (US2) para compor o roteamento completo em `cmd/web/main.go`.
- **Polish (Phase 6)**: Depende da conclusão de todas as User Stories (Phases 3, 4 e 5).

### User Story Dependency Graph

```text
Phase 1: Setup (T001-T003)
   │
   ▼
Phase 2: Foundational (T004-T011)
   ├──► Phase 3: US1 - Autenticação via Google Sem Senhas (T012-T019) 🎯 MVP
   │       │
   │       ▼
   ├──► Phase 4: US2 - Onboarding de Criador em 2 Etapas (T020-T026)
   │       │
   │       ▼
   └──► Phase 5: US3 - Migração Completa SSR → API (T027-T037)
           │
           ▼
        Phase 6: Polish & Validação Final (T038-T039)
```

### Parallel Opportunities

- **Phase 1**: `T002` e `T003` podem rodar em paralelo com `T001`.
- **Phase 2**: `T006`/`T007` (`TokenService`) e `T008`/`T009` (`GoogleTokenVerifier`) podem ser desenvolvidos em paralelo a `T004`/`T005` (`User` model).
- **Phase 3 (US1)**: Os testes `T012`, `T013` e `T014` podem ser escritos em paralelo; `T015` (DTOs) e `T018` (limpeza de `EmailService`) podem rodar em paralelo com `T016`.
- **Phase 4 (US2)**: Os testes `T020`, `T021` e `T022` podem rodar em paralelo; `T023` (DTOs) e `T024` (`CreatorOnboardingMiddleware`) podem rodar em paralelo.
- **Phase 5 (US3)**: Os testes `T027`, `T028`, `T029` e `T030` atuam em pacotes de domínio distintos (`account`, `library`, `sales`, `delivery`) e podem rodar totalmente em paralelo; da mesma forma, as migrações de handlers `T031`, `T032`, `T033`, `T034`, `T035` e `T036` são independentes entre si antes da montagem final em `T037` (`cmd/web/main.go`).

---

## Parallel Example: User Story 1, 2 & 3

```bash
# User Story 1 - Executar criação de testes em paralelo:
Task: "T012 [P] [US1] Atualizar testes de repositório em internal/auth/repository/user_repository_test.go"
Task: "T013 [P] [US1] Atualizar testes unitários de UserService em internal/auth/service/user_service_test.go"
Task: "T014 [P] [US1] Atualizar testes de AuthHandler em internal/auth/handler/auth_handler_test.go"

# User Story 2 - Executar middleware e DTOs de onboarding em paralelo:
Task: "T023 [P] [US2] Definir InputOnboardingCreator em internal/account/service/dto.go"
Task: "T024 [P] [US2] Implementar CreatorOnboardingMiddleware em internal/account/handler/middleware/creator_onboarding.go"

# User Story 3 - Migrar handlers de cada módulo de domínio em paralelo:
Task: "T032 [P] [US3] Migrar DashboardHandler, SettingsHandler e StripeConnectHandler em internal/account/handler/"
Task: "T033 [P] [US3] Migrar EbookHandler, FileHandler e SalesPageHandler em internal/library/handler/"
Task: "T034 [P] [US3] Migrar ClientHandler, PurchaseHandler, PurchaseSalesHandler, TransactionHandler e CheckoutHandler em internal/sales/handler/"
Task: "T035 [P] [US3] Migrar DownloadHandler em internal/delivery/handler/download_handler.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Concluir **Phase 1: Setup** (`T001`–`T003`)
2. Concluir **Phase 2: Foundational** (`T004`–`T011`)
3. Concluir **Phase 3: User Story 1** (`T012`–`T019`)
4. **PARAR e VALIDAR**: Testar independentemente a autenticação via Google (`POST /api/auth/google`), o vínculo automático de usuários existentes por e-mail sem senha e a consulta de sessão (`GET /api/auth/me`).

### Incremental Delivery

1. **Setup + Foundational** → Base JWT stateless, verificador Google OIDC e modelo `User` sem senha prontos.
2. **Add User Story 1 (P1)** → Login Google sem senhas e vínculo automático funcionando (MVP!).
3. **Add User Story 2 (P2)** → Onboarding de criador em 2 etapas (`POST /api/account/onboarding`) e proteção via `CreatorOnboardingMiddleware`.
4. **Add User Story 3 (P3)** → Migração 100% dos módulos (`account`, `library`, `sales`, `delivery`, `subscription`, `shared`) de SSR para JSON.
5. **Polish** → Limpeza final de artefatos SSR e validação ponta a ponta.
