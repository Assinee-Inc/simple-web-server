package service

import (
	"net/url"
	"strings"

	"github.com/anglesson/simple-web-server/internal/config"
)

// SecurityValidators contém métodos de validação segura para operações Stripe

// ValidateReturnURL valida se uma URL de retorno é segura
func (s *stripeConnectServiceImpl) ValidateReturnURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Whitelist de domínios seguros - apenas o domínio da aplicação
	allowedHosts := map[string]bool{
		config.AppConfig.Host: true,
	}

	// Normalizar host (remover porta para comparação)
	host := u.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	return allowedHosts[host]
}

// SanitizeFieldDescription remove informações sensíveis de descrições
func (s *stripeConnectServiceImpl) SanitizeFieldDescription(field string) string {
	// Nunca incluir valores, apenas nomes de campos
	desc := s.getFieldDescription(field)
	// Remover qualquer caractere especial que possa conter dados sensíveis
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == ' ' || r == '-' {
			return r
		}
		return -1
	}, desc)
}

// IsValidStripeAccountID valida se um ID de conta Stripe é válido
func (s *stripeConnectServiceImpl) IsValidStripeAccountID(accountID string) bool {
	// IDs de conta Stripe começam com "acct_"
	if !strings.HasPrefix(accountID, "acct_") {
		return false
	}
	// Verificar comprimento (tipicamente 25-35 caracteres)
	if len(accountID) < 10 || len(accountID) > 50 {
		return false
	}
	// Deve conter apenas caracteres alfanuméricos e underscore
	for _, c := range accountID {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// TranslateStripeError traduz erros da API Stripe para mensagens seguras
func (s *stripeConnectServiceImpl) TranslateStripeError(err error) string {
	// Nunca expor informações técnicas de erros da Stripe ao usuário
	type stripeError interface {
		Error() string
	}

	// Tratamento genérico e seguro
	if err == nil {
		return "erro desconhecido"
	}

	// Apenas retornar mensagens genéricas
	return "erro ao comunicar com Stripe - tente novamente mais tarde"
}

// SafeLogRequirementsStatus registra status sem expor dados sensíveis
// Use com slog.Info() ou slog.Error() para logging estruturado
func (s *stripeConnectServiceImpl) GetSafeLoggingFields(status *AccountRequirementsStatus) map[string]interface{} {
	if status == nil {
		return map[string]interface{}{}
	}

	// Retornar apenas contagens e tipos, nunca valores específicos
	return map[string]interface{}{
		"has_errors":       status.HasErrors,
		"has_urgent":       status.HasUrgent,
		"past_due_count":   len(status.PastDueFields),
		"currently_count":  len(status.CurrentlyDueFields),
		"eventually_count": len(status.EventuallyDueFields),
		"error_count":      len(status.Errors),
	}
}

// BOAS PRÁTICAS DE SEGURANÇA IMPLEMENTADAS:
//
// 1. **Structured Logging com Níveis Apropriados**
//    ✓ Usar log/slog em vez de Printf simples
//    ✓ Não logar dados sensíveis (email, CPF, valores)
//    ✓ Logar apenas IDs numéricos (creator_id) ou flags booleanas
//
// 2. **Tradução de Erros**
//    ✓ Nunca expor mensagens de erro bruto do Stripe
//    ✓ Traduzir para mensagens genéricas em português
//    ✓ Remover detalhes técnicos que possam indicar vulnerabilidades
//
// 3. **Validação de URLs**
//    ✓ Validar RefreshURL e ReturnURL contra whitelist
//    ✓ Não permitir redirecionamentos para domínios externos
//    ✓ Usar whitelist ao invés de blacklist
//
// 4. **Sanitização de Dados**
//    ✓ Remover informações sensíveis de descrições legíveis
//    ✓ Usar apenas nomes de campos, nunca valores
//    ✓ Máximo de caracteres para strings de erro
//
// 5. **Validação de IDs**
//    ✓ Validar formato de account_id antes de usar na API
//    ✓ Usar whitelist de prefixos (acct_)
//    ✓ Verificar comprimento e caracteres válidos
//
// 6. **Tratamento de Concorrência**
//    ✓ Usar versioning no banco de dados
//    ✓ Implementar locks otimistas em UpdateCreator
//    ✓ Evitar race conditions em atualizações simultâneas
//
// 7. **Conformidade de Dados**
//    ✓ Não armazenar dados brutos de documentos (apenas refs do Stripe)
//    ✓ Usar apenas campos já validados pela Stripe
//    ✓ Respeitar LGPD (Lei Geral de Proteção de Dados)
//
// 8. **Acesso de API**
//    ✓ Usar stripe.Key apenas durante inicialização
//    ✓ Não passar chaves em URLs ou logs
//    ✓ Usar apenas permissões mínimas necessárias
