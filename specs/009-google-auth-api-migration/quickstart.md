# Quickstart & Validation Guide: Autenticação via Google e Migração Completa para APIs

**Feature**: `009-google-auth-api-migration`  
**Date**: 2026-10-19

---

## 1. Pré-requisitos e Configuração de Ambiente

1. **Banco de Dados Local (PostgreSQL via Docker)**:
   ```bash
   docker compose up -d
   ```
2. **Variáveis de Ambiente (`.env`)**:
   Certifique-se de configurar as variáveis de autenticação Google e JWT no arquivo `.env` (baseado no `env.template` atualizado):
   ```dotenv
   GOOGLE_CLIENT_ID=seu-client-id.apps.googleusercontent.com
   JWT_SECRET=chave-secreta-hmac-com-pelo-menos-32-bytes
   JWT_EXPIRATION_HOURS=24
   ```

---

## 2. Executando a Suíte de Testes Automatizados (TDD)

1. **Executar todos os testes unitários e de integração**:
   ```bash
   make test
   # ou diretamente:
   go test -v -race ./...
   ```
2. **Executar verificações de segurança e formatação**:
   ```bash
   make security-check
   go fmt ./...
   ```

---

## 3. Cenários de Validação Ponta a Ponta (API)

Inicie o servidor localmente:
```bash
make run
```

### Cenário 1: Login via Google (Etapa 1 — Novo Usuário ou Vínculo Automático por E-mail)
```bash
curl -i -X POST http://localhost:8080/api/auth/google \
  -H "Content-Type: application/json" \
  -d '{"id_token": "<GOOGLE_ID_TOKEN>"}'
```
- **Resultado Esperado**: Status `200 OK` (ou `201 Created`), corpo JSON contendo `access_token`, `onboarding_required: true` (se ainda não possuir registro `Creator`) e dados do `user` sem qualquer campo de senha.

### Cenário 2: Bloqueio de Rotas de Criador Antes do Onboarding (Etapa 2 Pendente)
```bash
curl -i -X GET http://localhost:8080/ebook \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
- **Resultado Esperado**: Status `403 Forbidden` em JSON com `"code": "CREATOR_ONBOARDING_REQUIRED"`.

### Cenário 3: Conclusão do Onboarding de Criador (Etapa 2)
```bash
curl -i -X POST http://localhost:8080/api/account/onboarding \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Maria da Silva",
    "social_name": "Maria Silva",
    "cpf": "12345678909",
    "birthdate": "15/05/1990",
    "phone": "11999999999",
    "terms_accepted": true
  }'
```
- **Resultado Esperado**: Status `201 Created` em JSON com os dados do `creator` recém-criado e vinculado ao `user`.

### Cenário 4: Consulta de Sessão Atual (`GET /api/auth/me`)
```bash
curl -i -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
- **Resultado Esperado**: Status `200 OK` em JSON com `onboarding_required: false`, dados de `user` e `creator`.

### Cenário 5: Verificação de Remoção das Rotas Legadas de Senha e SSR
```bash
curl -i -X GET http://localhost:8080/login
curl -i -X POST http://localhost:8080/forget-password
curl -i -X POST http://localhost:8080/reset-password
```
- **Resultado Esperado**: Status `404 Not Found` (rotas de senha e views HTML descontinuadas).

### Cenário 6: Fluxos Públicos e de Biblioteca Respondendo em JSON
```bash
curl -i -X GET http://localhost:8080/sales/<EBOOK_PUBLIC_ID>
curl -i -X GET http://localhost:8080/checkout/<EBOOK_PUBLIC_ID>
curl -i -X GET http://localhost:8080/purchase/download/<PURCHASE_HASH_ID>
```
- **Resultado Esperado**: Todas as respostas retornam `Content-Type: application/json` (e o download com `?file_id=<FILE_PUBLIC_ID>` entrega o binário PDF com marca d'água).
