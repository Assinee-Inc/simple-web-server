# 🔒 Análise de Segurança - Upload Múltiplo de Arquivos

## 🚨 VULNERABILIDADES CRÍTICAS IDENTIFICADAS

### 1. **Limitação de Arquivos Simultâneos**
**PROBLEMA:** Não há limite de quantos arquivos podem ser enviados simultaneamente
**RISCO:** DoS, sobrecarga do servidor, consumo excessivo de banda
**SOLUÇÃO:**
```javascript
// Em ebook_handler.go - adicionar constante
const MAX_FILES_PER_UPLOAD = 10

// Em validateSelectedFiles()
if len(selectedFiles) > MAX_FILES_PER_UPLOAD {
    return fmt.Errorf("máximo %d arquivos por upload", MAX_FILES_PER_UPLOAD)
}
```

### 2. **Sanitização de Nome de Arquivo**
**PROBLEMA:** Nomes de arquivo não são sanitizados contra path traversal
**RISCO:** Path traversal, overwrite de arquivos importantes
**SOLUÇÃO:**
```go
// Função de sanitização segura
func sanitizeFilename(filename string) string {
    // Remover path separators
    filename = filepath.Base(filename)
    
    // Remover caracteres perigosos
    reg := regexp.MustCompile(`[<>:"/\\|?*]`)
    filename = reg.ReplaceAllString(filename, "_")
    
    // Limitar tamanho
    if len(filename) > 255 {
        ext := filepath.Ext(filename)
        filename = filename[:255-len(ext)] + ext
    }
    
    return filename
}
```

### 3. **Rate Limiting para Upload**
**PROBLEMA:** Não há rate limiting específico para uploads múltiplos
**RISCO:** DoS através de uploads massivos
**SOLUÇÃO:**
```go
// Em middleware de rate limiting
rateLimiter.AddRule("/dashboard/ebook/*/create", 5, time.Minute) // 5 uploads/minuto
rateLimiter.AddRule("/dashboard/ebook/*/update", 5, time.Minute) // 5 updates/minuto
```

### 4. **XSS em Templates**
**PROBLEMA:** Dados de arquivo não são escapados nos templates
**RISCO:** XSS através de nomes de arquivo maliciosos
**SOLUÇÃO:**
```html
<!-- Trocar de: -->
<div class="fw-semibold">{{.filename}}</div>

<!-- Para: -->
<div class="fw-semibold">{{.filename | html}}</div>
```

### 5. **Validação de Cota de Armazenamento**
**PROBLEMA:** Usuários podem fazer upload ilimitado
**RISCO:** Consumo excessivo de espaço de armazenamento
**SOLUÇÃO:**
```go
// Em file_service.go
const MAX_STORAGE_PER_USER = 1024 * 1024 * 1024 // 1GB

func (s *fileService) CheckUserStorageQuota(userID uint, newFileSize int64) error {
    var totalSize int64
    err := s.db.Model(&models.File{}).
        Where("creator_id = ?", userID).
        Select("COALESCE(SUM(size), 0)").
        Scan(&totalSize).Error
    
    if err != nil {
        return err
    }
    
    if totalSize + newFileSize > MAX_STORAGE_PER_USER {
        return fmt.Errorf("cota de armazenamento excedida. Limite: 1GB")
    }
    
    return nil
}
```

## ✅ PONTOS FORTES IDENTIFICADOS

### 1. **Validação Dupla de Arquivos**
- ✅ Validação de extensão + MIME type
- ✅ Uso de `http.DetectContentType()` para detecção real
- ✅ Lista de tipos permitidos bem definida

### 2. **Controle de Propriedade**
- ✅ Validação de propriedade de arquivos
- ✅ Verificação de permissões antes de adicionar arquivos ao ebook
- ✅ Isolamento entre usuários

### 3. **Validação de Tamanho**
- ✅ Limite de 50MB por arquivo
- ✅ Validação no backend (não só frontend)

### 4. **Tratamento de Erros**
- ✅ Mensagens de erro não expõem detalhes internos
- ✅ Logs seguros sem informações sensíveis

## 🔧 IMPLEMENTAÇÃO PRIORITÁRIA

### Prioridade ALTA (Implementar Imediatamente):
1. **Limitação de arquivos simultâneos** - Previne DoS
2. **Sanitização de nomes** - Previne path traversal  
3. **Rate limiting para uploads** - Previne abuso

### Prioridade MÉDIA (Implementar em próxima iteração):
4. **Escape XSS em templates** - Previne XSS
5. **Cota de armazenamento** - Controla uso de recursos

### Prioridade BAIXA (Melhorias futuras):
6. **Validação de tipos de arquivo mais rigorosa** - Magic bytes específicos
7. **Logs de auditoria detalhados** - Monitoramento aprimorado

## 📋 CHECKLIST DE VALIDAÇÃO

Antes do deploy em produção, verificar:

- [ ] Limitação máxima de arquivos por upload implementada
- [ ] Sanitização de nomes de arquivo funcionando
- [ ] Rate limiting específico para uploads configurado
- [ ] Templates escapam dados de arquivo adequadamente
- [ ] Cota de armazenamento por usuário implementada
- [ ] Testes de segurança passando
- [ ] Headers de segurança aplicados nas rotas de upload
- [ ] Logs de upload não expõem informações sensíveis

## 🧪 TESTES DE SEGURANÇA SUGERIDOS

```go
func TestUploadSecurity(t *testing.T) {
    // Teste path traversal
    t.Run("RejectPathTraversal", func(t *testing.T) {
        maliciousName := "../../../etc/passwd"
        // Verificar se é rejeitado
    })
    
    // Teste limite de arquivos
    t.Run("RejectTooManyFiles", func(t *testing.T) {
        files := make([]File, 20) // Mais que o limite
        // Verificar se é rejeitado
    })
    
    // Teste XSS em nome
    t.Run("SanitizeXSSInFilename", func(t *testing.T) {
        xssName := "<script>alert('xss')</script>.pdf"
        // Verificar se é sanitizado
    })
}
```

---

**⚠️ ATENÇÃO:** Esta análise identifica vulnerabilidades que devem ser corrigidas antes do deploy em produção. As soluções propostas seguem as diretrizes estabelecidas no `SECURITY_RULES.md` do projeto.
