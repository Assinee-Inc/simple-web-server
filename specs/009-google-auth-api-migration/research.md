# Phase 0 Research: Autenticação via Google e Migração Completa para APIs (Sem Senhas)

**Feature**: `009-google-auth-api-migration`  
**Date**: 2026-10-19

---

## 1. Validação de Identidade Google no Backend Go

- **Decision**: Suportar validação de **Google ID Token (OIDC)** enviado pelo cliente (`POST /api/auth/google` com `{"id_token": "..."}`) e/ou troca de Authorization Code OAuth2 (`{"code": "..."}`), validando a assinatura criptográfica JWKS do Google, `aud` (`GOOGLE_CLIENT_ID`), `iss` (`accounts.google.com` ou `https://accounts.google.com`), `exp` e `email_verified == true` por meio de uma interface `GoogleTokenVerifier` injetável no `auth/service`.
- **Rationale**:
  - Permite que clientes SPA/Mobile utilizem o Google Identity Services (GIS) para obter o `id_token` diretamente e enviá-lo ao backend em uma única chamada REST (`POST /api/auth/google`), além de ser trivialmente testável via mock da interface `GoogleTokenVerifier` em testes unitários e de integração (sem depender de chamadas de rede externas nos testes).
  - Garante que apenas contas com `email_verified: true` sejam aceitas (`FR-001`).
- **Alternatives considered**:
  - *Fluxo exclusivamente por redirecionamento OAuth2 server-side*: Acopla o backend a redirects de navegador e dificulta clientes SPA/mobile puros e testes automatizados de API.
  - *Confiar no payload decodificado sem verificar assinatura/audience*: Rejeitado por violar segurança básica (qualquer cliente poderia forjar tokens).

---

## 2. Mecanismo de Sessão Stateless para a API (JWT Bearer Token)

- **Decision**: Utilizar **JWT (JSON Web Token) assinado com HMAC-SHA256 (`HS256`)** via segredo de ambiente (`JWT_SECRET`, com fallback para `SESSION_AUTH_KEY` se não definido em dev), enviado no cabeçalho `Authorization: Bearer <token>`.
- **Rationale**:
  - Pode ser implementado diretamente com a biblioteca padrão do Go (`crypto/hmac`, `crypto/sha256`, `encoding/base64`, `encoding/json`) ou `golang-jwt/jwt/v5`, mantendo o princípio KISS e evitando dependências pesadas.
  - Elimina a necessidade de cookies de sessão `gorilla/sessions`, tokens CSRF baseados em formulário HTML e armazenamento de `FlashMessages` no servidor.
  - As claims do token incluem `sub` (`User.PublicID`), `uid` (`User.ID`), `email` (`User.Email`), `iat` e `exp` (padrão configurável, ex.: 24h ou 7 dias).
- **Alternatives considered**:
  - *Sessões baseadas em Cookie HTTP-Only (`gorilla/sessions`)*: Rejeitado pela definição de escopo `Q3: A` (Autenticação via Token Bearer/JWT stateless).
  - *Tokens opacos persistidos em banco a cada requisição*: Adiciona escrita/leitura extra em todas as rotas sem ganho necessário para o escopo atual.

---

## 3. Migração do Modelo `User` e Vínculo Automático de Contas Existentes

- **Decision**:
  - Adicionar o campo `GoogleID string` (`gorm:"type:varchar(128);uniqueIndex"`) na struct `auth/model.User`.
  - Remover os campos `Password`, `PasswordResetToken`, `PasswordResetAt`, `EmailVerificationToken`, `CSRFToken` e `SessionToken` da struct `auth/model.User` e criar migração/rotina de limpeza para remover/ignorar colunas legadas no PostgreSQL.
  - No fluxo de autenticação Google (`AuthenticateWithGoogle`):
    1. Buscar primeiro por `GoogleID == payload.Sub`.
    2. Caso não encontrado por `GoogleID`, buscar por `Email == payload.Email`.
    3. Se encontrado por `Email` (usuário legado ou já existente sem `GoogleID`), atualizar `user.GoogleID = payload.Sub` e preencher `user.EmailVerifiedAt` (se `nil`), preservando seu `ID` primário (`uint`) e mantendo intacto o vínculo com `Creator`, `Ebooks`, `Files`, `Subscriptions` e `Purchases`.
    4. Se não encontrado nem por `GoogleID` nem por `Email`, criar um novo `User` com `Username = payload.Name`, `Email = payload.Email`, `GoogleID = payload.Sub` e `EmailVerifiedAt = &now`.
- **Rationale**:
  - Atende integralmente `FR-001`, `FR-002`, `FR-004` e `SC-002`, garantindo zero perda de histórico para criadores já existentes e zero armazenamento de senhas.
- **Alternatives considered**:
  - *Exigir recadastro de contas existentes*: Rejeitado em `Q3: A` pois quebraria o vínculo com `Creator` e o histórico de vendas.

---

## 4. Onboarding de Criador em 2 Etapas e Middlewares de Autorização

- **Decision**:
  - Separar claramente os middlewares de proteção da API:
    1. **`AuthMiddleware` (JWT Bearer)**: Valida o header `Authorization: Bearer <token>`, carrega o `User` no contexto da requisição e retorna `401 Unauthorized` em JSON (`{"success": false, "error": "Não autorizado"}`) se ausente/inválido.
    2. **`CreatorOnboardingMiddleware`**: Aplicado às rotas de operação de negócio do criador (`dashboard`, `ebook`, `file`, `client`, `purchase/sales`, `transactions`, `settings`, `stripe-connect`). Verifica se existe um registro `Creator` para `user.ID`. Caso ainda não exista (Etapa 2 pendente), retorna `403 Forbidden` em JSON com código semântico (`{"success": false, "code": "CREATOR_ONBOARDING_REQUIRED", "error": "Conclua o cadastro de criador para acessar este recurso"}`).
    3. **`StripeOnboardingMiddleware` (API Mode)**: Nas rotas que exigem conta Stripe Connect ativa, em vez de redirecionar via `303 See Other` para `/stripe-connect/welcome` ou `/stripe-connect/status`, retorna `403 Forbidden` em JSON (`{"success": false, "code": "STRIPE_ONBOARDING_REQUIRED", "error": "Complete o onboarding da conta de recebimento", "onboarding_status": ...}`).
  - Disponibilizar o endpoint `POST /api/account/onboarding` (protegido apenas pelo `AuthMiddleware`), que recebe `name`, `social_name`, `cpf`, `birthdate`, `phone` e `terms_accepted`, valida maioridade (18+), CPF (checksum e Receita Federal quando em produção) e termos, cria o `Creator` vinculado ao `user.ID`, atualiza `user.TermsAcceptedAt` e provisiona a conta Stripe Connect.
- **Rationale**:
  - Implementa exatamente a jornada `Q2: A` (Login Google cria `User` na Etapa 1; endpoint de onboarding cria `Creator` na Etapa 2) e substitui os redirecionamentos SSR (`303 See Other`) por respostas JSON determinísticas para o cliente da API.
- **Alternatives considered**:
  - *Bloquear o login se não enviar CPF junto com o token Google*: Rejeitado em `Q2: A`.

---

## 5. Padronização dos Contratos de Resposta da API (100% SSR → API)

- **Decision**:
  - Eliminar a dependência de `pkg/template.TemplateRenderer` e `gorilla/sessions` (flash messages, form state em sessão e CSRF de formulário HTML) de todos os handlers HTTP (`auth`, `account`, `library`, `sales`, `delivery`, `subscription`, `shared`).
  - Adotar um helper leve de resposta JSON (`pkg/httputil` ou `internal/shared/web`) que define `Content-Type: application/json; charset=utf-8`, escreve o status HTTP apropriado (`200`, `201`, `400`, `401`, `403`, `404`, `409`, `422`, `500`) e serializa envelopes JSON consistentes.
  - Remover todas as rotas puramente de visualização de formulário HTML (`GET /login`, `GET /register`, `GET /forget-password`, `GET /reset-password`, `GET /password-reset-success`, `GET /check-email`, `GET /email-not-verified`, `GET /account-confirmation`, `GET /ebook/create`, `GET /ebook/edit/{id}`, `GET /file/upload`).
  - Converter as rotas de leitura e mutação de todos os módulos para retornarem JSON (mantendo stream binário/CSV apenas onde o próprio recurso é um arquivo: download de PDF com marca d'água em `GET /purchase/download/{hash_id}?file_id=...` e exportação CSV em `GET /client/export`, com erros retornados em JSON).
- **Rationale**:
  - Cumpre `Q1: C`, `FR-007`, `FR-008`, `FR-009` e `SC-004` mantendo a arquitetura modular existente (`handler` → `service` → `repository`) intacta.
- **Alternatives considered**:
  - *Manter handlers duplicados (SSR + API)*: Rejeitado por aumentar a complexidade de manutenção e contrariar a diretriz explícita de deixar de ser SSR.
