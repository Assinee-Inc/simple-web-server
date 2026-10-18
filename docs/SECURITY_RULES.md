# Regras de Segurança - Docffy

## 🚨 **REGRAS OBRIGATÓRIAS - SEMPRE RESPEITAR**

### 1. **Configuração de Cookies**
**REGRAS:**
- ✅ `Secure: true` em produção, `false` apenas em desenvolvimento
- ✅ `HttpOnly: true` para todos os cookies sensíveis (sessão, CSRF)
- ✅ `SameSite: http.SameSiteStrictMode` para cookies de autenticação
- ❌ NUNCA usar `Secure: false` em produção
- ❌ NUNCA usar `HttpOnly: false` para tokens de autenticação

**Exemplo Correto:**
```go
http.SetCookie(w, &http.Cookie{
    Name:     "session_token",
    Value:    token,
    HttpOnly: true,
    Secure:   config.AppConfig.IsProduction(),
    SameSite: http.SameSiteStrictMode,
})
```

### 2. **Credenciais e Configurações**
**REGRAS:**
- ✅ Usar APENAS variáveis de ambiente para credenciais
- ✅ NUNCA hardcodar senhas, chaves API ou tokens no código
- ✅ Usar valores vazios como fallback para credenciais obrigatórias
- ❌ NUNCA commitar arquivos `.env` no repositório
- ❌ NUNCA usar credenciais de exemplo no código

**Exemplo Correto:**
```go
AppConfig.StripeSecretKey = GetEnv("STRIPE_SECRET_KEY", "")
AppConfig.MailPassword = GetEnv("MAIL_PASSWORD", "")
```

### 3. **Logs e Informações Sensíveis**
**REGRAS:**
- ✅ Logs devem ser genéricos para informações sensíveis
- ✅ NUNCA logar tokens, senhas, emails completos ou dados pessoais
- ✅ Usar placeholders para informações sensíveis em logs
- ❌ NUNCA logar: tokens CSRF, session tokens, senhas, CPFs completos

**Exemplo Correto:**
```go
log.Printf("CSRF token mismatch for user: %s", user.Email)
log.Printf("User not found for session token")
```

**Exemplo Incorreto:**
```go
log.Printf("CSRF token: %s", csrfToken)
log.Printf("Session token: %s", sessionToken)
```

### 4. **Validação de Arquivos**
**REGRAS:**
- ✅ Sempre validar extensão E MIME type
- ✅ Verificar tamanho máximo do arquivo
- ✅ Usar `http.DetectContentType()` para validação real
- ✅ Manter lista de MIME types permitidos atualizada
- ❌ NUNCA confiar apenas na extensão do arquivo
- ❌ NUNCA permitir execução de arquivos

**Exemplo Correto:**
```go
// Validar extensão
ext := strings.ToLower(filepath.Ext(filename))
// Validar MIME type
mimeType := http.DetectContentType(buffer)
// Verificar lista permitida
if !allowedMimeTypes[mimeType] {
    return fmt.Errorf("tipo MIME não permitido: %s", mimeType)
}
```

### 5. **Headers de Segurança**
**REGRAS:**
- ✅ Sempre aplicar headers de segurança em todas as rotas
- ✅ Usar Content Security Policy (CSP)
- ✅ Configurar HSTS apenas em produção
- ✅ Remover headers que expõem informações do servidor
- ❌ NUNCA remover headers de segurança
- ❌ NUNCA usar CSP muito permissivo

**Headers Obrigatórios:**
```go
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("X-XSS-Protection", "1; mode=block")
w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
```

### 6. **Rate Limiting**
**REGRAS:**
- ✅ Sempre aplicar rate limiting em endpoints sensíveis
- ✅ Rate limiting diferente para diferentes tipos de endpoint
- ✅ Limpeza automática de dados de rate limiting
- ✅ Logs de tentativas de rate limiting excedido
- ❌ NUNCA remover rate limiting de endpoints de autenticação
- ❌ NUNCA usar rate limiting muito permissivo

**Limites Recomendados:**
- Autenticação: 5 requests/minuto
- Upload: 10 requests/minuto
- API: 100 requests/minuto

### 7. **Autenticação e Autorização**
**REGRAS:**
- ✅ Sempre validar tokens CSRF em operações de escrita
- ✅ Verificar permissões antes de acessar recursos
- ✅ Usar bcrypt para hash de senhas
- ✅ Tokens de sessão únicos e seguros
- ❌ NUNCA confiar apenas em cookies para autenticação
- ❌ NUNCA usar hash simples para senhas

### 8. **Validação de Input**
**REGRAS:**
- ✅ Sempre validar e sanitizar input do usuário
- ✅ Usar validação server-side (não apenas client-side)
- ✅ Validar tipos de dados, tamanhos e formatos
- ✅ Sanitizar dados antes de exibir em templates
- ❌ NUNCA confiar apenas em validação client-side
- ❌ NUNCA executar input do usuário

### 9. **Tratamento de Erros**
**REGRAS:**
- ✅ Logs de erro sem expor informações sensíveis
- ✅ Mensagens de erro genéricas para usuários
- ✅ Não expor stack traces em produção
- ✅ Tratamento graceful de erros
- ❌ NUNCA expor detalhes internos em erros
- ❌ NUNCA usar panic em produção

### 10. **Configuração de Produção**
**REGRAS:**
- ✅ HTTPS obrigatório em produção
- ✅ Cookies seguros em produção
- ✅ Headers de segurança em produção
- ✅ Logs estruturados em produção
- ❌ NUNCA usar configurações de desenvolvimento em produção
- ❌ NUNCA expor informações de debug em produção

## 🔧 **IMPLEMENTAÇÃO DE NOVAS FUNCIONALIDADES**

### Checklist Obrigatório:
- [ ] Cookies configurados corretamente
- [ ] Headers de segurança aplicados
- [ ] Rate limiting implementado
- [ ] Validação de input completa
- [ ] Logs sem informações sensíveis
- [ ] Tratamento de erros adequado
- [ ] Testes de segurança incluídos

### Validação Automática:
```bash
# Verificar configurações de segurança
make security-check

# Verificar headers de segurança
make security-headers-test

# Verificar rate limiting
make rate-limit-test
```

## 🚨 **PENALIDADES**

### Violações Críticas (Bloqueio de Merge):
- Credenciais hardcoded
- Cookies inseguros em produção
- Logs de informações sensíveis
- Falta de validação de arquivos

### Violações Moderadas (Aviso):
- Headers de segurança ausentes
- Rate limiting não implementado
- Validação de input insuficiente

### Violações Menores (Sugestão):
- Logs muito verbosos
- Mensagens de erro muito específicas
- Configurações não otimizadas

## 📋 **REVISÃO DE CÓDIGO**

### Checklist para Code Review:
1. **Segurança de Cookies** ✅
2. **Headers de Segurança** ✅
3. **Rate Limiting** ✅
4. **Validação de Input** ✅
5. **Logs Seguros** ✅
6. **Tratamento de Erros** ✅
7. **Configuração de Produção** ✅

### Comandos de Verificação:
```bash
# Verificar configurações
grep -r "Secure.*false" internal/
grep -r "HttpOnly.*false" internal/
grep -r "password.*=" internal/config/

# Verificar headers
grep -r "X-Content-Type-Options" internal/
grep -r "X-Frame-Options" internal/

# Verificar rate limiting
grep -r "RateLimit" internal/
```

## ✅ **MELHORIAS IMPLEMENTADAS**

### 1. **Cookies Seguros** ✅
- `Secure: config.AppConfig.IsProduction()` em todos os cookies
- `HttpOnly: true` para tokens CSRF e sessão
- `SameSite: http.SameSiteStrictMode` configurado

### 2. **Headers de Segurança** ✅
- Content Security Policy implementado
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Referrer-Policy: strict-origin-when-cross-origin
- HSTS configurado para produção

### 3. **Rate Limiting** ✅
- Rate limiting implementado para diferentes endpoints
- Limpeza automática de dados
- Limites configurados:
  - Autenticação: 5 requests/minuto
  - Upload: 10 requests/minuto
  - API: 100 requests/minuto

### 4. **Validação de Arquivos Melhorada** ✅
- Validação de extensão E MIME type
- `http.DetectContentType()` implementado
- Lista de MIME types permitidos atualizada

### 5. **Logs Seguros** ✅
- Tokens removidos dos logs
- Informações sensíveis não expostas
- Logs genéricos para auditoria

### 6. **Credenciais Seguras** ✅
- Credenciais hardcoded removidas
- Apenas variáveis de ambiente
- Valores vazios como fallback

### 7. **Tratamento de Erros** ✅
- Panic removido do código
- Tratamento graceful de erros
- Logs de erro sem informações sensíveis

## 🔄 **ATUALIZAÇÕES**

Esta documentação deve ser revisada e atualizada:
- A cada nova funcionalidade de segurança implementada
- Após incidentes de segurança
- A cada 6 meses para revisão geral
- Quando novas vulnerabilidades são descobertas

---

**Última atualização:** Dezembro 2024
**Responsável:** Equipe de Segurança
**Versão:** 2.0
**Status:** Implementado ✅ 