# Modularização da Aplicação

## Visão Geral

Esta aplicação está sendo migrada de uma arquitetura em **camadas planas** para uma arquitetura **orientada a módulos de negócio**. A migração é realizada de forma gradual — um módulo por vez — garantindo que os testes passem a cada etapa.

### Estrutura anterior (camadas planas)
```
internal/
├── models/       → todos os modelos de domínio juntos
├── service/      → todos os serviços juntos
├── repository/   → todos os repositórios juntos
└── handler/      → todos os handlers juntos
```

### Estrutura-alvo (módulos de negócio)
```
internal/
├── auth/         → User, Login, sessão, autenticação
├── account/      → Creator (perfil do criador de conteúdo)
├── library/      → Ebook, File (conteúdo digital)
├── sales/        → Purchase, Transaction, Client
├── delivery/     → DownloadLog (entrega do conteúdo)
└── subscription/ → Subscription (assinatura da plataforma)
```

Cada módulo contém suas próprias camadas:
```
internal/<módulo>/
├── model/        → structs de domínio + DTOs de input
├── repository/   → interfaces + implementações GORM
├── service/      → lógica de negócio + validators
└── handler/      → HTTP handlers e middleware
```

---

## Regras de Comunicação entre Módulos

### 1. Comunicação via IDs primitivos (não via objetos)

Módulos **não importam tipos de domínio de outros módulos**. A comunicação entre módulos é feita exclusivamente por IDs primitivos (`uint`):

```go
// CORRETO — retorna uint, não *auth.User
func (s *UserServiceImpl) CreateUser(input InputCreateUser) (uint, error)

// CORRETO — recebe subscriptionID uint, não *Subscription
func (s *subscriptionServiceImpl) ActivateSubscription(subscriptionID uint, ...) error

// ERRADO — cria acoplamento entre módulos
func (s *CreatorServiceImpl) CreateCreator(input Input) (*auth.User, error)
```

As FKs nas structs GORM são campos `uint` (`UserID`, `CreatorID`), sem campos de associação (`User`, `Creator`) que exigiriam imports cruzados.

### 2. DTOs de input ficam no pacote `model/` (pacote folha)

DTOs usados como parâmetros de entrada (`Input*`) são definidos em `<módulo>/model/`, **não** em `<módulo>/service/`. Isso evita ciclos de importação em testes:

```
// Ciclo que ocorre se o DTO ficar em service/:
mocks → account/service → subscription/service → mocks  ❌

// Com o DTO em model/ (pacote folha sem dependências de serviço):
mocks → account/model  ✅
```

O arquivo `<módulo>/service/dto.go` pode existir como alias de retrocompatibilidade:
```go
// internal/account/service/dto.go
type InputCreateCreator = accountmodel.InputCreateCreator
```

### 3. Imports permitidos entre módulos

Um módulo **só pode importar módulos abaixo dele** na hierarquia de dependências. Nunca no sentido inverso.

Qualquer import cruzado de tipos de domínio deve ser substituído por uma referência via ID (`uint`).

---

## Módulos e Responsabilidades

### `auth` — Autenticação
**Entidades:** `User`, `Login`

Responsável por registro, login, logout, recuperação de senha e gerenciamento de sessão.

**Conteúdo:**
- `model/user.go` — struct User
- `model/login.go` — struct Login
- `model/dto.go` — `InputCreateUser`, `InputLogin`
- `repository/user_repository.go` — interface + implementação GORM
- `service/user_service.go` — lógica de criação e autenticação de usuários
- `service/user_validator.go` — validação de inputs de usuário
- `service/session.go` — gerenciamento de sessão e CSRF
- `handler/auth_handler.go` — rotas de login, registro, reset de senha
- `handler/middleware/authorizer.go` — middleware de autenticação
- `handler/middleware/guard.go` — middleware de redirecionamento para usuários autenticados

---

### `subscription` — Assinatura
**Entidades:** `Subscription`

Gerencia assinaturas da plataforma, período de trial e integração com Stripe para cobranças recorrentes.

**Conteúdo:**
- `model/subscription.go`
- `repository/subscription_repository.go` + `subscription_gorm.go`
- `service/subscription_service.go`
- `service/stripe_service.go` — cliente Stripe de baixo nível
- `service/stripe_payment_gateway.go` — implementação do gateway de pagamento para assinaturas
- `handler/middleware/subscription.go` — verifica status da assinatura
- `handler/middleware/trial.go` — verifica período de trial

---

### `account` — Conta do Criador
**Entidades:** `Creator`

Perfil do criador de conteúdo, onboarding no Stripe Connect, dashboard e configurações.

**Conteúdo:**
- `model/creator.go`
- `model/dto.go` — `InputCreateCreator`
- `repository/creator_repository.go` + `creator_gorm.go`
- `repository/dashboard_repository.go`
- `service/creator_service.go`
- `service/creator_validator.go`
- `service/stripe_connect_service.go`
- `handler/creator_handler.go`
- `handler/stripe_connect_handler.go`
- `handler/settings_handler.go`
- `handler/dashboard_handler.go`
- `handler/middleware/stripe_onboarding.go`

---

### `library` — Biblioteca de Conteúdo
**Entidades:** `Ebook`, `File`

Gerenciamento de ebooks e arquivos digitais, upload para S3 e DRM.

**Conteúdo:**
- `model/ebook.go`
- `model/file.go`
- `repository/ebook_repository.go`
- `repository/file_repository.go`
- `service/ebook_service.go`
- `service/file_service.go`
- `handler/ebook_handler.go`
- `handler/file_handler.go`
- `handler/sales_page_handler.go`

---

### `sales` — Vendas
**Entidades:** `Purchase`, `Transaction`, `Client`, `ClientCreator`

Processo de compra, checkout, pagamentos via Stripe, gestão de clientes e relatórios de vendas.

**Conteúdo:**
- `model/client.go`
- `model/client_creator.go` — tabela de junção Client ↔ Creator
- `model/purchase.go`
- `model/transaction.go`
- `repository/client_repository.go` + `client_gorm.go`
- `repository/purchase_repository.go`
- `repository/transaction_repository.go`
- `service/client_service.go` + `client_validator.go`
- `service/purchase_service.go`
- `service/transaction_service.go`
- `service/payment_gateway.go` — interface do gateway de pagamento
- `service/resend_download_link_service.go`
- `handler/client_handler.go`
- `handler/purchase_handler.go`
- `handler/purchase_sales_handler.go`
- `handler/transaction_handler.go`
- `handler/checkout_handler.go`
- `handler/stripe_handler.go` — webhooks Stripe

---

### `delivery` — Entrega
**Entidades:** `DownloadLog`

Controle de download do conteúdo adquirido: geração de links com hash, validação de acesso e registro de logs de download. Não é responsável por transformações no arquivo entregue.

**Conteúdo:**
- `model/download.go`
- `repository/download_repository.go` — persistência de DownloadLog
- `service/download_service.go` — geração e validação de links de download
- `handler/download_handler.go` — rota pública de download

---

## Hierarquia de Dependências

```
auth (base — sem dependências internas)
  ↑
  ├── account (Creator.UserID uint → referencia auth.User por ID)
  │     ↑
  │     └── library (Ebook.CreatorID uint → referencia account.Creator por ID)
  │               ↑
  │               └── sales (Purchase.EbookID uint → referencia library.Ebook por ID
  │                          Transaction.CreatorID uint → referencia account.Creator por ID
  │                          ClientCreator.ClientID + CreatorID uint → junção por IDs)
  │                         ↑
  │                         └── delivery (DownloadLog.PurchaseID uint → referencia sales.Purchase por ID)
  │
  └── subscription (Subscription.UserID uint → referencia auth.User por ID)
```

**Regra fundamental:** um módulo só pode importar módulos **abaixo** dele na hierarquia. Referências a entidades de outros módulos são feitas exclusivamente por `uint` (FK), nunca importando o tipo do módulo referenciado.

---

## Infraestrutura Compartilhada (não muda)

Os pacotes abaixo não fazem parte de nenhum módulo de negócio e permanecem em `pkg/`:

| Pacote | Responsabilidade |
|---|---|
| `pkg/database/` | Conexão e migração do banco de dados |
| `pkg/mail/` | Envio de emails (SMTP) |
| `pkg/storage/` | Upload de arquivos para AWS S3 |
| `pkg/template/` | Renderização de templates HTML |
| `pkg/utils/` | Utilitários genéricos (formatação, validação) |
| `pkg/cookie/` | Gerenciamento de cookies |
| `pkg/gov/` | Integração com a Receita Federal |
| `pkg/drm/` | Aplicação de watermark em PDFs e integração com SheetDB para registro de marcas d'água |
| `internal/config/` | Configuração global da aplicação |

> **Atenção:** `pkg/database` importa os pacotes `model/` de cada módulo para o `AutoMigrate`. Por isso, os pacotes `model/` **não podem importar** nada de `pkg/database` (nem indiretamente via service). Qualquer import de `pkg/database` deve ficar em `repository/` ou em `cmd/web/main.go`.

---

## Tipos Compartilhados

Tipos usados por múltiplos módulos que não pertencem a um módulo específico ficam em `internal/shared/`:

| Arquivo | Conteúdo |
|---|---|
| `internal/shared/types.go` | DTOs e tipos de formulário genéricos |
| `internal/shared/filter.go` | Paginação e filtros de listagem |
| `internal/shared/contact.go` | Struct Contact |
| `internal/shared/dto/email_dto.go` | DTOs de email |

---

## Plano de Migração

### Ordem das fases (do mais isolado ao mais dependente)

| Fase | Módulo | Status |
|---|---|---|
| 1 | `auth` | ✅ Concluído |
| 2 | `subscription` | ✅ Concluído |
| 3 | `account` | ✅ Concluído |
| 4 | `library` | ✅ Concluído |
| 5 | `sales` | ✅ Concluído |
| 6 | `delivery` | ⏳ Pendente |

### Passos para cada fase

Para cada módulo, executar nesta ordem:

1. Criar estrutura de diretórios do módulo
2. Mover arquivos de model + atualizar `package` declarations
3. Definir DTOs de input em `model/dto.go` (pacote folha, sem imports de service)
4. Mover repositórios + atualizar imports
5. Mover services + atualizar imports; criar alias em `service/dto.go` se necessário
6. Mover handlers + atualizar imports
7. Atualizar `cmd/web/main.go` com os novos import paths
8. Criar shims de retrocompatibilidade em `internal/service/`, `internal/repository/` e `internal/handler/` para não quebrar código ainda não migrado
9. Atualizar mocks em `internal/mocks/` para importar de `model/` (não de `service/`)
10. Rodar `go build ./...` — deve compilar sem erros
11. Rodar `go test ./...` — todos os testes devem passar
12. Commit da fase

### Comandos úteis

```bash
# Encontrar arquivos que importam um pacote
grep -r "simple-web-server/internal/models" --include="*.go" -l

# Verificar compilação
go build ./...

# Rodar testes
go test ./...

# Verificar sem warnings
go vet ./...

# Verificar dependências de um pacote (detecta ciclos potenciais)
go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/...
```

---

## Checklist de conclusão (todas as fases)

- [ ] `go build ./...` sem erros
- [ ] `go test ./...` todos passando
- [ ] `go vet ./...` sem warnings
- [ ] `internal/models/` removido (ou apenas com shims temporários)
- [ ] `internal/service/` removido
- [ ] `internal/repository/` removido (exceto `memory.go` se usado em testes)
- [ ] `internal/handler/` contém apenas arquivos que não pertencem a nenhum módulo
- [ ] `cmd/web/main.go` importa apenas os novos módulos
