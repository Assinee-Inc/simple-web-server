# SimpleWebServer

## 🐳 Desenvolvimento Local com Docker

### Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- `make`

### Configuração

1. **Clonar o repositório:**
   ```bash
   git clone https://github.com/anglesson/simple-web-server.git
   cd simple-web-server
   ```

2. **Configurar variáveis de ambiente:**
   ```bash
   make setup-env
   ```
   Edite o `.env` preenchendo os valores necessários, especialmente `SESSION_AUTH_KEY` e `SESSION_ENC_KEY`.

3. **Iniciar o ambiente:**
   ```bash
   make up
   ```
   Este comando builda as imagens Docker e inicia os containers `app` e `postgres` em background.

### Comandos

- **`make up`** - Inicia todos os containers Docker
- **`make down`** - Para todos os containers Docker
- **`make logs`** - Visualiza os logs dos serviços em tempo real
- **`make dev`** - Inicia o ambiente e o servidor com hot-reload

### Acesso

A aplicação estará disponível em `http://localhost:8080`

---

## 🚀 Configuração Rápida

### 1. Configurar Variáveis de Ambiente

```bash
# Criar arquivo .env com as configurações padrão
make setup-env

# Ou manualmente
cp env.template .env
```

### 2. Editar Configurações

Edite o arquivo `.env` com suas configurações:

```bash
# Configurações obrigatórias para produção
MAIL_USERNAME=seu_email@gmail.com
MAIL_PASSWORD=sua_senha_de_app
STRIPE_SECRET_KEY=sk_test_...
S3_ACCESS_KEY=sua_access_key
S3_SECRET_KEY=sua_secret_key
HUB_DEVSENVOLVEDOR_TOKEN=seu_token
SESSION_AUTH_KEY=... # Chave de 64 bytes (128 caracteres hex)
SESSION_ENC_KEY=...  # Chave de 32 bytes (64 caracteres hex)
```

### 3. Executar Aplicação

```bash
# Instalar dependências
go mod download

# Executar em desenvolvimento
make run

# Ou executar diretamente
go run cmd/web/main.go
```

## 📋 Configuração Completa

### Variáveis de Ambiente

| Variável | Descrição | Padrão | Obrigatória |
|----------|-----------|--------|-------------|
| `APPLICATION_MODE` | Modo da aplicação | `development` | Não |
| `APPLICATION_NAME` | Nome da aplicação | `Docffy` | Não |
| `APP_KEY` | Chave da aplicação | `Docffy` | Sim |
| `HOST` | Host da aplicação | `http://localhost` | Não |
| `PORT` | Porta da aplicação | `8080` | Não |
| `DATABASE_URL` | URL do banco de dados | `./mydb.db` | Não |
| `MAIL_HOST` | Servidor SMTP | `sandbox.smtp.mailtrap.io` | Não |
| `MAIL_PORT` | Porta SMTP | `2525` | Não |
| `MAIL_USERNAME` | Usuário SMTP | - | Sim (prod) |
| `MAIL_PASSWORD` | Senha SMTP | - | Sim (prod) |
| `MAIL_FROM_ADDRESS` | Email remetente | - | Sim (prod) |
| `S3_ACCESS_KEY` | AWS Access Key | - | Não |
| `S3_SECRET_KEY` | AWS Secret Key | - | Não |
| `S3_REGION` | Região AWS | `sa-east-1` | Não |
| `S3_BUCKET_NAME` | Nome do bucket S3 | - | Não |
| `STRIPE_SECRET_KEY` | Chave secreta Stripe | - | Sim (prod) |
| `STRIPE_PRICE_ID` | ID do preço Stripe | - | Não |
| `STRIPE_WEBHOOK_SECRET` | Segredo do webhook | - | Não |
| `HUB_DEVSENVOLVEDOR_TOKEN` | Token Receita Federal | - | Não |

### Configurações por Ambiente

#### Desenvolvimento
```bash
APPLICATION_MODE=development
DATABASE_URL=./mydb.db
MAIL_HOST=sandbox.smtp.mailtrap.io
```

#### Produção
```bash
APPLICATION_MODE=production
DATABASE_URL=postgres://user:pass@localhost/dbname
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=seu_email@gmail.com
MAIL_PASSWORD=sua_senha_de_app
STRIPE_SECRET_KEY=sk_live_...
S3_ACCESS_KEY=sua_access_key
S3_SECRET_KEY=sua_secret_key
```

## 🔒 Segurança

### Arquivo .env
- ✅ **NUNCA** commite o arquivo `.env` no repositório
- ✅ O arquivo `.env` está no `.gitignore` por segurança
- ✅ Use o arquivo `env.template` como base
- ✅ Configure credenciais reais apenas em produção

### Verificações de Segurança
```bash
# Verificar configurações de segurança
make security-check

# Verificar headers de segurança
make security-headers-test

# Verificar rate limiting
make rate-limit-test
```

## 📚 Documentação

- [Regras de Segurança](docs/SECURITY_RULES.md)
- [Guia DashUI](docs/DASHUI_GUIDE.md)
- [Sobre o Projeto](docs/ABOUT_PROJETCT.md)

## 🛠️ Comandos Úteis

```bash
# Configurar ambiente
make setup-env

# Executar aplicação
make run

# Executar testes
make test

# Verificar segurança
make security-check

# Build para produção
make build

# Executar com Docker
make docker-build
make docker-run
```

## 🚨 Importante

1. **Configure sempre as credenciais reais em produção**
2. **Nunca use credenciais de exemplo em produção**
3. **Mantenha o arquivo .env seguro e nunca o commite**
4. **Execute as verificações de segurança regularmente**

---

Para mais informações, consulte a [documentação completa](docs/).
