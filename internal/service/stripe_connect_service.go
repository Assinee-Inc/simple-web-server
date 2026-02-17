package service

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
	"github.com/stripe/stripe-go/v76/accountlink"
)

// RequirementPriority categoriza a urgência da pendência
type RequirementPriority string

const (
	PriorityPastDue    RequirementPriority = "past_due"       // Urgente - está vencido
	PriorityCurrently  RequirementPriority = "currently_due"  // Crítico - necessário agora
	PriorityEventually RequirementPriority = "eventually_due" // Baixa - será necessário depois
)

// RequirementType categoriza o tipo de pendência
type RequirementType string

const (
	RequirementTypeDocument   RequirementType = "document"   // Documentos
	RequirementTypePersonal   RequirementType = "personal"   // Dados pessoais
	RequirementTypeCompliance RequirementType = "compliance" // Compliance/Termos
	RequirementTypeOther      RequirementType = "other"      // Outros
)

// PendingRequirement estrutura cada pendência identificada
type PendingRequirement struct {
	Field       string              // Ex: "individual.verification.document"
	Type        RequirementType     // Categorização do tipo
	Priority    RequirementPriority // Urgência
	Description string              // Descrição legível para o usuário
	Resolvable  bool                // Se pode ser resolvido via onboarding link
}

// AccountRequirementsStatus consolida o estado de pendências da conta
type AccountRequirementsStatus struct {
	HasErrors           bool                 // Se há errors em Requirements.Errors
	HasUrgent           bool                 // Se há pendências currently_due ou past_due
	PastDueFields       []PendingRequirement // Campos vencidos (máxima urgência)
	CurrentlyDueFields  []PendingRequirement // Campos necessários agora
	EventuallyDueFields []PendingRequirement // Campos que serão necessários depois
	Errors              []string             // Mensagens de erro traduzidas
}

type StripeConnectService interface {
	CreateConnectAccount(creator *models.Creator) (string, error)
	CreateOnboardingLink(accountID string, refreshURL string, returnURL string) (string, error)
	GetAccountDetails(accountID string) (*stripe.Account, error)
	UpdateCreatorFromAccount(creator *models.Creator, account *stripe.Account) error
	GerarLinkRemediacao(creator *models.Creator) (string, error)
	AnalyzeRequirements(account *stripe.Account) (*AccountRequirementsStatus, error)
	GetPendingDocuments(account *stripe.Account) []PendingRequirement
	HasCriticalPendencies(status *AccountRequirementsStatus) bool
	ShouldGenerateRemediationLink(status *AccountRequirementsStatus) bool
}

type stripeConnectServiceImpl struct {
	creatorSvc CreatorService
}

func NewStripeConnectService(creatorSvc CreatorService) StripeConnectService {
	stripe.Key = config.AppConfig.StripeSecretKey
	return &stripeConnectServiceImpl{
		creatorSvc,
	}
}

// CreateConnectAccount creates a new Stripe Connect account for the creator
func (s *stripeConnectServiceImpl) CreateConnectAccount(creator *models.Creator) (string, error) {
	// Split name into first and last name (basic approach)
	names := strings.Fields(creator.Name)
	firstName := names[0]
	lastName := ""
	if len(names) > 1 {
		lastName = strings.Join(names[1:], " ")
	}

	params := &stripe.AccountParams{
		Type:         stripe.String("express"),
		Country:      stripe.String("BR"), // Brazil
		Email:        stripe.String(creator.Email),
		BusinessType: stripe.String("individual"),
		Individual: &stripe.PersonParams{
			FirstName: stripe.String(firstName),
			LastName:  stripe.String(lastName),
			Email:     stripe.String(creator.Email),
			Phone:     stripe.String("+55" + creator.Phone),
			IDNumber:  stripe.String(creator.CPF),
			DOB: &stripe.PersonDOBParams{
				Day:   stripe.Int64(int64(creator.BirthDate.Day())),
				Month: stripe.Int64(int64(creator.BirthDate.Month())),
				Year:  stripe.Int64(int64(creator.BirthDate.Year())),
			},
		},
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
	}

	acc, err := account.New(params)
	if err != nil {
		log.Printf("Error creating Stripe Connect account: %v", err)
		return "", fmt.Errorf("erro ao criar conta no Stripe: %v", err)
	}

	return acc.ID, nil
}

// CreateOnboardingLink creates an onboarding link for the creator to complete their Stripe setup
func (s *stripeConnectServiceImpl) CreateOnboardingLink(accountID string, refreshURL string, returnURL string) (string, error) {
	params := &stripe.AccountLinkParams{
		Account:    stripe.String(accountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(params)
	if err != nil {
		log.Printf("Error creating onboarding link: %v", err)
		return "", fmt.Errorf("erro ao criar link de onboarding: %v", err)
	}

	return link.URL, nil
}

// GetAccountDetails retrieves account details from Stripe
func (s *stripeConnectServiceImpl) GetAccountDetails(accountID string) (*stripe.Account, error) {
	acc, err := account.GetByID(accountID, nil)
	if err != nil {
		log.Printf("Error retrieving account details: %v", err)
		return nil, fmt.Errorf("erro ao buscar detalhes da conta: %v", err)
	}

	return acc, nil
}

// UpdateCreatorFromAccount updates creator with account status from Stripe
func (s *stripeConnectServiceImpl) UpdateCreatorFromAccount(creator *models.Creator, account *stripe.Account) error {
	slog.Info("updating creator stripe status",
		slog.Any("creator_id", creator.ID),
		slog.Bool("details_submitted", account.DetailsSubmitted),
		slog.Bool("charges_enabled", account.ChargesEnabled),
		slog.Bool("payouts_enabled", account.PayoutsEnabled))

	// Analisar requirements estruturadamente
	reqStatus, err := s.AnalyzeRequirements(account)
	if err != nil {
		slog.Error("failed to analyze requirements",
			slog.Any("creator_id", creator.ID),
			slog.String("error", err.Error()))
		return err
	}

	// Atualizar flags de status
	creator.OnboardingCompleted = account.DetailsSubmitted
	creator.PayoutsEnabled = account.PayoutsEnabled
	creator.ChargesEnabled = account.ChargesEnabled

	// Determinar se há urgência de remediação
	if s.ShouldGenerateRemediationLink(reqStatus) {
		slog.Info("generating remediation link",
			slog.Any("creator_id", creator.ID),
			slog.Int("urgent_requirements", len(reqStatus.PastDueFields)+len(reqStatus.CurrentlyDueFields)))

		urlRefresh, err := s.GerarLinkRemediacao(creator)
		if urlRefresh != "" && err == nil {
			slog.Info("remediation link generated",
				slog.Any("creator_id", creator.ID))
			creator.OnboardingRefreshURL = urlRefresh
		} else if err != nil {
			slog.Error("failed to generate remediation link",
				slog.Any("creator_id", creator.ID),
				slog.String("error", err.Error()))
		}
	} else if reqStatus.HasUrgent {
		slog.Info("account has non-critical pending requirements",
			slog.Any("creator_id", creator.ID))
	}

	// Persistir no banco
	err = s.creatorSvc.UpdateCreator(creator)
	if err != nil {
		slog.Error("failed to update creator",
			slog.Any("creator_id", creator.ID),
			slog.String("error", err.Error()))
		return fmt.Errorf("erro ao atualizar criador: %v", err)
	}

	return nil
}

func (s *stripeConnectServiceImpl) GerarLinkRemediacao(creator *models.Creator) (string, error) {
	if creator.StripeConnectAccountID == "" {
		return "", fmt.Errorf("creator não possui conta Stripe Connect")
	}

	params := &stripe.AccountLinkParams{
		Account:    stripe.String(creator.StripeConnectAccountID),
		Type:       stripe.String("account_onboarding"),
		ReturnURL:  stripe.String(config.AppConfig.Host + "/stripe-connect/status"),
		RefreshURL: stripe.String(config.AppConfig.Host + "/stripe-connect/status"),
	}
	result, err := accountlink.New(params)
	if err != nil {
		slog.Error("failed to create remediation link",
			slog.String("account_id", creator.StripeConnectAccountID),
			slog.String("error", err.Error()))
		return "", fmt.Errorf("erro ao criar link de remediação: %v", err)
	}

	return result.URL, err
}

// AnalyzeRequirements analisa os requirements de uma conta Stripe de forma estruturada
func (s *stripeConnectServiceImpl) AnalyzeRequirements(account *stripe.Account) (*AccountRequirementsStatus, error) {
	if account == nil || account.Requirements == nil {
		return &AccountRequirementsStatus{
			HasErrors:           false,
			HasUrgent:           false,
			PastDueFields:       []PendingRequirement{},
			CurrentlyDueFields:  []PendingRequirement{},
			EventuallyDueFields: []PendingRequirement{},
			Errors:              []string{},
		}, nil
	}

	status := &AccountRequirementsStatus{
		HasErrors:           len(account.Requirements.PastDue) > 0 || len(account.Requirements.CurrentlyDue) > 0,
		HasUrgent:           false,
		PastDueFields:       []PendingRequirement{},
		CurrentlyDueFields:  []PendingRequirement{},
		EventuallyDueFields: []PendingRequirement{},
		Errors:              []string{},
	}

	// Processar campos past_due (máxima urgência)
	for _, field := range account.Requirements.PastDue {
		req := s.mapFieldToRequirement(field, PriorityPastDue)
		status.PastDueFields = append(status.PastDueFields, req)
		status.HasUrgent = true
	}

	// Processar campos currently_due (urgência alta)
	for _, field := range account.Requirements.CurrentlyDue {
		req := s.mapFieldToRequirement(field, PriorityCurrently)
		status.CurrentlyDueFields = append(status.CurrentlyDueFields, req)
		status.HasUrgent = true
	}

	// Processar campos eventually_due (urgência baixa)
	for _, field := range account.Requirements.EventuallyDue {
		req := s.mapFieldToRequirement(field, PriorityEventually)
		status.EventuallyDueFields = append(status.EventuallyDueFields, req)
	}

	return status, nil
}

// GetPendingDocuments retorna apenas as pendências relacionadas a documentos
func (s *stripeConnectServiceImpl) GetPendingDocuments(account *stripe.Account) []PendingRequirement {
	var documents []PendingRequirement
	if account == nil || account.Requirements == nil {
		return documents
	}

	documentKeywords := []string{
		"verification",
		"document",
		"bank_account",
		"identity",
	}

	// Coletar todos os campos de pendências
	allFields := make([]string, 0)
	allFields = append(allFields, account.Requirements.PastDue...)
	allFields = append(allFields, account.Requirements.CurrentlyDue...)
	allFields = append(allFields, account.Requirements.EventuallyDue...)

	// Filtrar apenas documentos
	for _, field := range allFields {
		if s.isDocumentField(field, documentKeywords) {
			priority := s.getPriorityForField(field, account.Requirements)
			doc := s.mapFieldToRequirement(field, priority)
			documents = append(documents, doc)
		}
	}

	return documents
}

// HasCriticalPendencies verifica se há pendências críticas (past_due ou currently_due)
func (s *stripeConnectServiceImpl) HasCriticalPendencies(status *AccountRequirementsStatus) bool {
	if status == nil {
		return false
	}
	return len(status.PastDueFields) > 0 || len(status.CurrentlyDueFields) > 0
}

// ShouldGenerateRemediationLink determina se deve gerar link de remediação
func (s *stripeConnectServiceImpl) ShouldGenerateRemediationLink(status *AccountRequirementsStatus) bool {
	if status == nil {
		return false
	}
	// Gerar link apenas se há pendências urgentes (not just eventually_due)
	return s.HasCriticalPendencies(status) || status.HasErrors
}

// mapFieldToRequirement converte um campo de pendência para struct estruturada
func (s *stripeConnectServiceImpl) mapFieldToRequirement(field string, priority RequirementPriority) PendingRequirement {
	reqType := s.classifyRequirementType(field)
	description := s.getFieldDescription(field)

	return PendingRequirement{
		Field:       field,
		Type:        reqType,
		Priority:    priority,
		Description: description,
		Resolvable:  true, // Todas as pendências do Stripe podem ser resolvidas via onboarding link
	}
}

// classifyRequirementType categoriza o tipo de pendência
func (s *stripeConnectServiceImpl) classifyRequirementType(field string) RequirementType {
	// Documentos
	documentPatterns := []string{"verification", "document", "bank_account", "identity"}
	for _, pattern := range documentPatterns {
		if strings.Contains(strings.ToLower(field), pattern) {
			return RequirementTypeDocument
		}
	}

	// Dados pessoais
	personalPatterns := []string{"ssn", "address", "dob", "email", "phone", "name"}
	for _, pattern := range personalPatterns {
		if strings.Contains(strings.ToLower(field), pattern) {
			return RequirementTypePersonal
		}
	}

	// Compliance
	compliancePatterns := []string{"tos", "terms", "acceptance", "agreement"}
	for _, pattern := range compliancePatterns {
		if strings.Contains(strings.ToLower(field), pattern) {
			return RequirementTypeCompliance
		}
	}

	return RequirementTypeOther
}

// getFieldDescription retorna uma descrição legível para o campo
func (s *stripeConnectServiceImpl) getFieldDescription(field string) string {
	descriptions := map[string]string{
		"individual.verification.document": "Documento de identidade",
		"individual.verification.status":   "Status da verificação de identidade",
		"bank_account_verification":        "Verificação da conta bancária",
		"individual.ssn_last_4":            "Últimos 4 dígitos do CPF",
		"individual.address.postal_code":   "CEP do endereço",
		"individual.address.city":          "Cidade do endereço",
		"individual.address.line1":         "Endereço de residência",
		"individual.address.state":         "Estado do endereço",
		"individual.email":                 "Email de contato",
		"individual.phone":                 "Telefone de contato",
		"individual.dob":                   "Data de nascimento",
		"tos_acceptance":                   "Aceitação de termos de serviço",
		"individual.first_name":            "Primeiro nome",
		"individual.last_name":             "Sobrenome",
	}

	if desc, exists := descriptions[field]; exists {
		return desc
	}

	// Fallback: usar o nome do campo formatado
	return strings.ReplaceAll(strings.Title(strings.ReplaceAll(field, "_", " ")), ".", " - ")
}

// isDocumentField verifica se um campo é relacionado a documentos
func (s *stripeConnectServiceImpl) isDocumentField(field string, keywords []string) bool {
	fieldLower := strings.ToLower(field)
	for _, keyword := range keywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}
	return false
}

// getPriorityForField determina a prioridade de um campo
func (s *stripeConnectServiceImpl) getPriorityForField(field string, req *stripe.AccountRequirements) RequirementPriority {
	// Verificar em qual lista o campo aparece
	for _, pastDue := range req.PastDue {
		if pastDue == field {
			return PriorityPastDue
		}
	}

	for _, currently := range req.CurrentlyDue {
		if currently == field {
			return PriorityCurrently
		}
	}

	for _, eventually := range req.EventuallyDue {
		if eventually == field {
			return PriorityEventually
		}
	}

	return PriorityEventually // Padrão
}

// translateRequirementError traduz erros do Stripe para mensagens seguras
func (s *stripeConnectServiceImpl) translateRequirementError(err interface{}) string {
	if err == nil {
		return "erro desconhecido"
	}

	// Traduzir apenas o código e campo, sem valores ou detalhes técnicos
	translations := map[string]string{
		"invalid_address_city_state_postal_code": "Endereço inválido",
		"invalid_business_profile_name":          "Nome da empresa inválido",
		"invalid_dob":                            "Data de nascimento inválida",
		"verification_document_address_mismatch": "Endereço do documento não corresponde",
		"verification_document_expired":          "Documento expirado",
		"verification_document_corrupt":          "Documento corrompido",
		"verification_failed_document_match":     "Documento não verificado",
		"verification_document_not_uploaded":     "Documento não enviado",
		"verification_missing_owners":            "Proprietários não fornecidos",
		"verification_missing_directors":         "Diretores não fornecidos",
	}

	// Tentar extrair código e requirement se for um AccountRequirementsError
	var code, requirement string
	if reqErr, ok := err.(*stripe.AccountRequirementsError); ok {
		code = reqErr.Code
		requirement = reqErr.Requirement
	}

	if trans, exists := translations[code]; exists {
		if requirement != "" {
			return fmt.Sprintf("%s (%s)", trans, requirement)
		}
		return trans
	}

	// Fallback genérico sem expor detalhes técnicos
	return "Requisito pendente"
}
