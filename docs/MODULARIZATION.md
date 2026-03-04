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
├── model/        → structs de domínio
├── repository/   → interfaces + implementações GORM
├── service/      → lógica de negócio + validators
└── handler/      → HTTP handlers e middleware
```

---

## Módulos e Responsabilidades

### `auth` — Autenticação
**Entidades:** `User`, `Login`

Responsável por registro, login, logout, recuperação de senha e gerenciamento de sessão.

**Conteúdo:**
- `model/user.go` — struct User
- `model/login.go` — DTO Login
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

Gerenciamento de ebooks e arquivos digitais, upload para S3, watermark e DRM.

**Conteúdo:**
- `model/ebook.go`
- `model/file.go`
- `repository/ebook_repository.go`
- `repository/file_repository.go`
- `service/ebook_service.go`
- `service/file_service.go`
- `service/watermark_service.go`
- `service/drm_service.go`
- `handler/ebook_handler.go`
- `handler/file_handler.go`
- `handler/watermark_handler.go`
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

Controle de download do conteúdo adquirido, links de download com hash, logs de acesso.

**Conteúdo:**
- `model/download.go`
- `handler/download_handler.go` — rota pública de download

---

## Hierarquia de Dependências

```
auth (base — sem dependências internas)
  ↑
  ├── account (Creator.UserID → auth.User)
  │     ↑
  │     └── library (Ebook.CreatorID → account.Creator)
  │               ↑
  │               └── sales (Purchase.EbookID → library.Ebook
  │                          Transaction.CreatorID → account.Creator
  │                          Client ←→ Creator via ClientCreator)
  │                         ↑
  │                         └── delivery (DownloadLog.PurchaseID → sales.Purchase)
  │
  └── subscription (Subscription.UserID → auth.User)
```

**Regra fundamental:** um módulo só pode importar módulos **abaixo** dele na hierarquia. Importações circulares não são permitidas.

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
| `internal/config/` | Configuração global da aplicação |

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
| 2 | `subscription` | ⏳ Pendente |
| 3 | `account` | ⏳ Pendente |
| 4 | `library` | ⏳ Pendente |
| 5 | `sales` | ⏳ Pendente |
| 6 | `delivery` | ⏳ Pendente |

### Passos para cada fase

Para cada módulo, executar nesta ordem:

1. Criar estrutura de diretórios do módulo
2. Mover arquivos de model + atualizar `package` declarations
3. Mover repositórios + atualizar imports
4. Mover services + atualizar imports
5. Mover handlers + atualizar imports
6. Atualizar `cmd/web/main.go` com os novos import paths
7. Rodar `go build ./...` — deve compilar sem erros
8. Rodar `go test ./...` — todos os testes devem passar
9. Atualizar mocks em `internal/mocks/` se necessário
10. Commit da fase

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
