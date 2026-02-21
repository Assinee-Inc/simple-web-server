package service

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
)

// Exemplos práticos de como acessar stripe.Account.Requirements corretamente
// Este arquivo contém exemplos de como trabalhar com erros de Requirements do Stripe

// Exemplo 1: Iteração simples sobre erros
func exampleIterateErrors(account *stripe.Account) {
	if account == nil || account.Requirements == nil {
		fmt.Println("Account ou Requirements é nil")
		return
	}

	// Verificar se há erros
	if len(account.Requirements.Errors) == 0 {
		fmt.Println("Sem erros de requirement")
		return
	}

	// Iterar sobre os erros - FORMA CORRETA
	for _, reqErr := range account.Requirements.Errors {
		// reqErr é do tipo *stripe.AccountRequirementsError
		fmt.Printf("Código: %s\n", reqErr.Code)
		fmt.Printf("Motivo: %s\n", reqErr.Reason)
		fmt.Printf("Campo: %s\n\n", reqErr.Requirement)
	}
}

// Exemplo 2: Classificar e processar erros
func exampleClassifyErrors(account *stripe.Account) {
	if account == nil || account.Requirements == nil {
		return
	}

	documentErrors := []string{}
	verificationErrors := []string{}
	otherErrors := []string{}

	for _, reqErr := range account.Requirements.Errors {
		switch {
		case contains(reqErr.Code, "document"):
			documentErrors = append(documentErrors, reqErr.Requirement)
		case contains(reqErr.Code, "verification"):
			verificationErrors = append(verificationErrors, reqErr.Requirement)
		default:
			otherErrors = append(otherErrors, reqErr.Requirement)
		}
	}

	fmt.Printf("Erros de documento: %v\n", documentErrors)
	fmt.Printf("Erros de verificação: %v\n", verificationErrors)
	fmt.Printf("Outros erros: %v\n", otherErrors)
}

// Exemplo 3: Compilar mensagens para o usuário
func exampleCompileErrorMessages(account *stripe.Account) []string {
	messages := []string{}

	if account == nil || account.Requirements == nil {
		return messages
	}

	// Tradução de códigos de erro para português
	translations := map[string]string{
		"invalid_address_city_state_postal_code": "O endereço está incompleto",
		"verification_document_expired":          "O documento enviado expirou",
		"verification_document_corrupt":          "O documento não está legível",
		"verification_failed_document_match":     "Os dados do documento não correspondem",
	}

	for _, reqErr := range account.Requirements.Errors {
		var message string

		// Buscar tradução
		if trans, ok := translations[reqErr.Code]; ok {
			message = trans
		} else {
			// Se não tiver tradução, usar o motivo da Stripe ou fallback
			message = reqErr.Reason
			if message == "" {
				message = "Há um problema com seu requisito de verificação"
			}
		}

		// Adicionar campo se disponível
		if reqErr.Requirement != "" {
			message = fmt.Sprintf("%s (%s)", message, humanizeField(reqErr.Requirement))
		}

		messages = append(messages, message)
	}

	return messages
}

// Exemplo 4: Verificar se há erros críticos
func exampleCheckCriticalErrors(account *stripe.Account) bool {
	if account == nil || account.Requirements == nil {
		return false
	}

	criticalCodes := map[string]bool{
		"verification_document_expired":      true,
		"verification_document_corrupt":      true,
		"verification_failed_document_match": true,
		"verification_failed_keyed_identity": true,
	}

	for _, reqErr := range account.Requirements.Errors {
		if criticalCodes[reqErr.Code] {
			return true
		}
	}

	return false
}

// Exemplo 5: Usar com FutureRequirements também
func exampleFutureRequirementsErrors(account *stripe.Account) {
	if account == nil || account.FutureRequirements == nil {
		return
	}

	// FutureRequirements usa AccountFutureRequirementsError
	for _, futureErr := range account.FutureRequirements.Errors {
		fmt.Printf("Erro futuro - Código: %s, Campo: %s\n",
			futureErr.Code,
			futureErr.Requirement)
	}
}

// Exemplo 6: Completo - análise estruturada
type RequirementAnalysis struct {
	HasErrors            bool
	ErrorCount           int
	CriticalErrorCount   int
	ErrorsByType         map[string][]string
	UserFriendlyMessages []string
}

func exampleAnalyzeRequirements(account *stripe.Account) *RequirementAnalysis {
	analysis := &RequirementAnalysis{
		HasErrors:            false,
		ErrorCount:           0,
		CriticalErrorCount:   0,
		ErrorsByType:         make(map[string][]string),
		UserFriendlyMessages: []string{},
	}

	if account == nil || account.Requirements == nil {
		return analysis
	}

	if len(account.Requirements.Errors) == 0 {
		return analysis
	}

	analysis.HasErrors = true

	criticalCodes := map[string]bool{
		"verification_document_expired":      true,
		"verification_document_corrupt":      true,
		"verification_failed_document_match": true,
	}

	for _, reqErr := range account.Requirements.Errors {
		analysis.ErrorCount++

		// Classificar por tipo
		errType := classifyErrorType(reqErr.Code)
		analysis.ErrorsByType[errType] = append(
			analysis.ErrorsByType[errType],
			reqErr.Requirement,
		)

		// Contar erros críticos
		if criticalCodes[reqErr.Code] {
			analysis.CriticalErrorCount++
		}

		// Gerar mensagens amigáveis
		msg := generateUserMessage(reqErr)
		analysis.UserFriendlyMessages = append(analysis.UserFriendlyMessages, msg)
	}

	return analysis
}

// Funções auxiliares

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func humanizeField(field string) string {
	fieldNames := map[string]string{
		"individual.address.postal_code":   "CEP/Código Postal",
		"individual.email":                 "Email",
		"individual.ssn_last_4":            "CPF",
		"individual.verification.document": "Documento de Identificação",
		"bank_account_verification":        "Verificação da Conta Bancária",
	}

	if name, ok := fieldNames[field]; ok {
		return name
	}
	return field
}

func classifyErrorType(code string) string {
	switch {
	case contains(code, "document"):
		return "document"
	case contains(code, "verification"):
		return "verification"
	case contains(code, "address"):
		return "address"
	default:
		return "other"
	}
}

func generateUserMessage(reqErr *stripe.AccountRequirementsError) string {
	translations := map[string]string{
		"verification_document_expired":          "Seu documento expirou. Por favor, atualize com um documento válido.",
		"verification_document_corrupt":          "Não conseguimos ler seu documento. Tente enviá-lo novamente.",
		"verification_failed_document_match":     "As informações do documento não correspondem aos seus dados. Por favor, verifique.",
		"invalid_address_city_state_postal_code": "Endereço incompleto. Por favor, forneça uma cidade, estado e CEP válidos.",
	}

	if msg, ok := translations[reqErr.Code]; ok {
		return msg
	}

	if reqErr.Reason != "" {
		return fmt.Sprintf("Por favor, corrija: %s", reqErr.Reason)
	}

	return "Por favor, atualize suas informações de verificação."
}

// ExampleMain pode ser usado para testar localmente
// Para executar: go run ./internal/service/stripe_requirements_examples.go -run ExampleMain
func ExampleAnalyzeRequirementsUsage() {
	// Criar conta de teste com erro
	account := &stripe.Account{
		Requirements: &stripe.AccountRequirements{
			Errors: []*stripe.AccountRequirementsError{
				{
					Code:        "verification_document_expired",
					Reason:      "The document has expired",
					Requirement: "individual.verification.document",
				},
				{
					Code:        "invalid_address_city_state_postal_code",
					Reason:      "Invalid postal code",
					Requirement: "individual.address.postal_code",
				},
			},
		},
	}

	fmt.Println("=== Exemplo 1: Iteração Simples ===")
	exampleIterateErrors(account)

	fmt.Println("\n=== Exemplo 3: Mensagens para Usuário ===")
	messages := exampleCompileErrorMessages(account)
	for i, msg := range messages {
		fmt.Printf("%d. %s\n", i+1, msg)
	}

	fmt.Println("\n=== Exemplo 4: Verificar Erros Críticos ===")
	hasCritical := exampleCheckCriticalErrors(account)
	fmt.Printf("Tem erros críticos? %v\n", hasCritical)

	fmt.Println("\n=== Exemplo 6: Análise Completa ===")
	analysis := exampleAnalyzeRequirements(account)
	fmt.Printf("Has Errors: %v\n", analysis.HasErrors)
	fmt.Printf("Error Count: %d\n", analysis.ErrorCount)
	fmt.Printf("Critical Errors: %d\n", analysis.CriticalErrorCount)
	fmt.Printf("Tipos de erros: %v\n", analysis.ErrorsByType)
}
