# API Contracts: Docffy REST API (Google Auth & Full SSR-to-API Migration)

**Feature**: `009-google-auth-api-migration`  
**Date**: 2026-10-19  
**Format**: JSON (`Content-Type: application/json; charset=utf-8`), exceto endpoints de upload (`multipart/form-data`) e download binário/CSV (`application/pdf` / `text/csv`).

---

## 1. Convenções Gerais de Resposta e Autenticação

### Cabeçalho de Autenticação (Rotas Protegidas)
```http
Authorization: Bearer <access_token>
```

### Envelope Padrão de Erro (JSON)
```json
{
  "success": false,
  "code": "ERROR_CODE_OPTIONAL",
  "error": "Mensagem descritiva do erro",
  "details": {
    "campo": "Detalhe de validação quando aplicável"
  }
}
```

### Códigos HTTP Semânticos
- `200 OK` / `201 Created`: Operação realizada com sucesso.
- `400 Bad Request` / `422 Unprocessable Entity`: Falha de validação de dados de entrada.
- `401 Unauthorized`: Token ausente, inválido, expirado ou credencial Google inválida.
- `403 Forbidden`: Usuário autenticado sem permissão para o recurso, com onboarding de criador pendente (`CREATOR_ONBOARDING_REQUIRED`) ou com onboarding Stripe Connect pendente (`STRIPE_ONBOARDING_REQUIRED`).
- `404 Not Found`: Recurso não encontrado.
- `409 Conflict`: Conflito de unicidade (ex.: CPF já cadastrado ou compra duplicada).
- `500 Internal Server Error`: Erro inesperado no servidor.

---

## 2. Módulo `auth` (Autenticação via Google Sem Senha)

### 2.1. `POST /api/auth/google` (Público)
Autentica um usuário existente (vinculando automaticamente pelo e-mail se ainda não vinculado ao `google_id`) ou cria um novo `User` (Etapa 1 do onboarding).

- **Request Body**:
  ```json
  {
    "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6..."
  }
  ```
- **Response `200 OK` (Usuário existente) / `201 Created` (Novo usuário)**:
  ```json
  {
    "success": true,
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "onboarding_required": true,
    "user": {
      "public_id": "usr_abc123",
      "username": "Maria Silva",
      "email": "maria@example.com",
      "google_id": "109876543210",
      "email_verified": true,
      "terms_accepted": false
    }
  }
  ```
- **Response `401 Unauthorized`** (Token Google inválido ou `email_verified == false`):
  ```json
  {
    "success": false,
    "error": "Credencial Google inválida ou e-mail não verificado"
  }
  ```

### 2.2. `GET /api/auth/me` (Autenticado: `AuthMiddleware`)
Retorna os dados do usuário autenticado, o estado do perfil de criador (`Creator`) e o status da assinatura.

- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "onboarding_required": false,
    "user": {
      "public_id": "usr_abc123",
      "username": "Maria Silva",
      "email": "maria@example.com",
      "email_verified": true,
      "terms_accepted": true,
      "subscription_status": "active",
      "days_left": 25
    },
    "creator": {
      "public_id": "crt_xyz789",
      "name": "Maria da Silva",
      "social_name": "Maria Silva",
      "email": "maria@example.com",
      "phone": "11999999999",
      "onboarding_completed": true,
      "charges_enabled": true,
      "payouts_enabled": true,
      "needs_stripe_onboarding": false
    }
  }
  ```

### 2.3. `POST /api/auth/logout` (Autenticado: `AuthMiddleware`)
Encerra a sessão atual do usuário (limpando eventuais cookies residuais e confirmando encerramento para o cliente descartar o token).
- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Sessão encerrada com sucesso"
  }
  ```

> **Rotas Legadas Removidas (`404 Not Found`)**:
> `GET /login`, `POST /login`, `GET /register`, `POST /register`, `GET /forget-password`, `POST /forget-password`, `GET /reset-password`, `POST /reset-password`, `GET /password-reset-success`, `GET /check-email`, `GET /account-confirmation`, `GET /email-not-verified`, `POST /resend-confirmation`.

---

## 3. Módulo `account` (Onboarding de Criador, Dashboard, Configurações e Stripe Connect)

### 3.1. `POST /api/account/onboarding` (Autenticado: `AuthMiddleware`)
**Etapa 2 do Onboarding**: Cria o perfil `Creator` para o usuário autenticado via Google, valida CPF/maioridade/termos e inicia o provisionamento da conta Stripe Connect.

- **Request Body**:
  ```json
  {
    "name": "Maria da Silva",
    "social_name": "Maria Silva",
    "cpf": "12345678909",
    "birthdate": "15/05/1990",
    "phone": "11999999999",
    "terms_accepted": true
  }
  ```
- **Response `201 Created`**:
  ```json
  {
    "success": true,
    "creator": {
      "public_id": "crt_xyz789",
      "name": "Maria da Silva",
      "social_name": "Maria Silva",
      "email": "maria@example.com",
      "cpf": "12345678909",
      "phone": "11999999999",
      "stripe_connect_account_id": "acct_123",
      "onboarding_completed": false
    }
  }
  ```
- **Response `400 Bad Request` / `409 Conflict`**: CPF inválido/duplicado, menor de 18 anos ou termos não aceitos.

### 3.2. `GET /dashboard` (ou `GET /api/dashboard`) (Autenticado + Criador)
Retorna as métricas consolidadas do dashboard do criador em JSON.
- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "total_ebooks": 5,
    "total_sent_ebooks": 42,
    "total_clients": 38,
    "ebook_stats": {},
    "daily_purchases": [],
    "daily_downloads": [],
    "top_ebooks": [],
    "top_clients": [],
    "top_downloaded_ebooks": []
  }
  ```

### 3.3. Configurações do Criador (`/settings`)
- **`GET /settings`**: Retorna os dados do criador e configurações atuais (`social_name`, `facebook_pixel_id`) em JSON (`200 OK`).
- **`POST /settings/social-name`**: Atualiza o nome social via JSON ou form (`{"social_name": "..."}`) e retorna `200 OK` com `{"success": true, "social_name": "..."}`.
- **`POST /settings/facebook-pixel`**: Valida (10–20 dígitos ou vazio) e atualiza o Meta Pixel ID (`{"facebook_pixel_id": "..."}`), retornando `200 OK` ou `400 Bad Request` em JSON.

### 3.4. Onboarding Stripe Connect (`/stripe-connect/*`)
- **`GET /stripe-connect/status`** (e `GET /stripe-connect/welcome`): Retorna o status detalhado da conta Stripe Connect em JSON (`has_stripe_account`, `onboarding_completed`, `payouts_enabled`, `charges_enabled`, `needs_onboarding`, `remediation_link`).
- **`GET /stripe-connect/onboard`** (ou `POST /stripe-connect/onboard`): Cria (se necessário) a conta Stripe Connect e gera o link de onboarding, retornando `200 OK` em JSON com `{"success": true, "onboarding_url": "https://connect.stripe.com/..."}`.
- **`GET /stripe-connect/complete`**: Sincroniza os detalhes da conta Stripe após retorno do onboarding e retorna `200 OK` em JSON com `{"success": true, "onboarding_status": "completed|pending", "creator": ...}`.

---

## 4. Módulo `library` (Ebooks, Arquivos e Página de Vendas)

### 4.1. Ebooks (Autenticado + Criador)
- **`GET /ebook`**: Lista paginada de ebooks do criador (`?page=1&page_size=10&title=...`) em JSON (`{"success": true, "ebooks": [...], "pagination": {...}}`).
- **`POST /ebook/create`**: Cria um novo ebook (suporta `multipart/form-data` para upload de imagem de capa/arquivos ou JSON com IDs de arquivos existentes), retornando `201 Created` com os dados do ebook criado ou `400 Bad Request` com mapa `errors` em JSON.
- **`GET /ebook/view/{id}`**: Retorna os detalhes do ebook e a lista paginada de clientes compradores em JSON (`200 OK`).
- **`POST /ebook/update/{id}`** (e `PUT /ebook/{id}`): Atualiza dados do ebook e anexos, retornando `200 OK` em JSON.
- **`POST /ebook/delete/{id}`** (e `DELETE /ebook/{id}`): Remove o ebook do criador, retornando `200 OK` em JSON (`{"success": true, "message": "E-book removido com sucesso"}`).
- **`POST /ebook/{id}/remove-file/{fileId}`**: Desvincula um arquivo de um ebook, retornando `200 OK` em JSON.
- **`GET /ebook/{id}/image`**: Retorna `200 OK` em JSON com a URL pré-assinada temporária do S3 (`{"success": true, "image_url": "https://..."}`).
- **`GET /ebook/preview/{id}`**: Retorna os dados da página de vendas em modo preview (`is_preview: true`, `original_price`, `savings`, `ebook`, `creator`) em JSON (`200 OK`).

### 4.2. Arquivos (`/file`) (Autenticado + Criador)
- **`GET /file`**: Lista paginada de arquivos do criador (`?page=1&page_size=20&search=...&type=...`) em JSON (`200 OK`).
- **`POST /file/upload`**: Recebe upload `multipart/form-data` (`file`, `name`, `description`), armazena no S3 e retorna `201 Created` em JSON com os metadados do arquivo criado.
- **`POST /file/{id}/update`**: Atualiza `name` e `description` do arquivo, retornando `200 OK` em JSON.
- **`POST /file/{id}/delete`**: Remove o arquivo do S3 e do banco, retornando `200 OK` em JSON.

### 4.3. Página de Vendas Pública (`/sales/{id}`)
- **`GET /sales/{id}`** (Público): Incrementa a contagem de visualização (controlada por cookie de 24h) e retorna os dados públicos da oferta do ebook e do autor em JSON (`200 OK`), ou `404 Not Found` em JSON se inexistente/inativo.

---

## 5. Módulo `sales` (Clientes, Vendas, Transações e Checkout)

### 5.1. Clientes (`/client`) (Autenticado + Criador)
- **`GET /client`**: Lista paginada de clientes do criador (`?page=1&page_size=10&term=...`) em JSON (`200 OK`).
- **`POST /client/{id}/update`**: Atualiza `email` e `phone` do cliente, retornando `200 OK` em JSON.
- **`GET /client/export`**: Mantém a exportação `text/csv` (`Nome,Email,Telefone`), retornando erro estruturado em JSON (`401`/`500`) caso ocorra falha antes da escrita do CSV.

### 5.2. Vendas e Transações (Autenticado + Criador)
- **`GET /purchase/sales`**: Lista paginada de vendas do criador com filtros (`ebook_id`, `client_name`, `client_email`, `page`) e mapa de transações associadas em JSON (`200 OK`).
- **`POST /purchase/sales/block-download`**: Bloqueia o download de uma compra (`purchase_id`), retornando `200 OK` em JSON (`{"success": true, "download_blocked": true}`).
- **`POST /purchase/sales/unblock-download`**: Desbloqueia o download de uma compra (`purchase_id`), retornando `200 OK` em JSON (`{"success": true, "download_blocked": false}`).
- **`POST /purchase/sales/resend-link`**: Reenvia o link de download por e-mail ao comprador, retornando `200 OK` em JSON.
- **`GET /transactions/detail`**: Retorna os detalhes completos da transação (`?id=tx_...`) em JSON (`200 OK`).
- **`POST /transactions/resend-download-link`**: Reenvia o link de download a partir de uma transação completada, retornando `200 OK` em JSON.

### 5.3. Checkout Público (`/checkout/{id}`, `/purchase/success`, `/api/*`)
- **`GET /checkout/{id}`**: Retorna os dados necessários para o checkout público do ebook (`ebook`, `creator`) em JSON (`200 OK`).
- **`POST /api/validate-customer`**: Valida CPF, maioridade, duplicidade de compra e Receita Federal em JSON (já existente, mantido).
- **`POST /api/create-ebook-checkout`**: Cria o cliente/compra pendente e retorna a URL da sessão de checkout Stripe em JSON (já existente, mantido).
- **`GET /purchase/success`**: Valida a `session_id` do Stripe e retorna os dados de confirmação da compra (`ebook`, `purchase`, `customer_email`, `creator_email`) em JSON (`200 OK`).

---

## 6. Módulo `delivery` (Entrega e Download Protegido)

### 6.1. `GET /purchase/download/{hash_id}` (Público via Hash de Compra)
- **Sem `?file_id`**: Retorna em JSON (`200 OK`) os metadados da compra (`ebook_title`, `expires_at`, `downloads_used`, `download_limit`) e a lista de arquivos disponíveis para download (`files`).
  - Caso o pagamento esteja pendente, o link tenha expirado ou o limite de downloads tenha sido atingido, retorna `403 Forbidden` / `410 Gone` em JSON com código semântico (`PAYMENT_PENDING`, `DOWNLOAD_EXPIRED`, `DOWNLOAD_LIMIT_EXCEEDED`).
- **Com `?file_id={file_public_id}`**: Gera o PDF com marca d'água dinâmica contendo nome/CPF do comprador e transmite o binário (`application/octet-stream` / `application/pdf`), ou retorna erro em JSON (`403`/`404`/`410`/`500`) caso a validação falhe.
