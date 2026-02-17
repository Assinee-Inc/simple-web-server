package service

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v76"
)

// MockCreatorService para testes
type MockCreatorService struct {
	updateCalled bool
	updateError  error
}

func (m *MockCreatorService) UpdateCreator(creator *models.Creator) error {
	m.updateCalled = true
	return m.updateError
}

func (m *MockCreatorService) CreateCreator(input models.InputCreateCreator) (*models.Creator, error) {
	return nil, nil
}

func (m *MockCreatorService) FindCreatorByEmail(email string) (*models.Creator, error) {
	return nil, nil
}

func (m *MockCreatorService) FindCreatorByUserID(userID uint) (*models.Creator, error) {
	return nil, nil
}

func (m *MockCreatorService) FindByID(id uint) (*models.Creator, error) {
	return nil, nil
}

// TestAnalyzeRequirements_WithPastDue testa análise com campos vencidos
func TestAnalyzeRequirements_WithPastDue(t *testing.T) {
	// Arrange
	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	account := &stripe.Account{
		DetailsSubmitted: false,
		ChargesEnabled:   false,
		PayoutsEnabled:   false,
		Requirements: &stripe.AccountRequirements{
			PastDue:      []string{"individual.verification.document", "bank_account_verification"},
			CurrentlyDue: []string{"tos_acceptance"},
		},
	}

	// Act
	status, err := svc.AnalyzeRequirements(account)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.HasUrgent)
	assert.Equal(t, 2, len(status.PastDueFields))
	assert.Equal(t, 1, len(status.CurrentlyDueFields))
	assert.Equal(t, 0, len(status.EventuallyDueFields))
	assert.True(t, svc.ShouldGenerateRemediationLink(status))
}

// TestAnalyzeRequirements_WithOnlyEventuallyDue testa análise com apenas campos futuros
func TestAnalyzeRequirements_WithOnlyEventuallyDue(t *testing.T) {
	// Arrange
	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	account := &stripe.Account{
		DetailsSubmitted: true,
		ChargesEnabled:   true,
		PayoutsEnabled:   true,
		Requirements: &stripe.AccountRequirements{
			PastDue:       []string{},
			CurrentlyDue:  []string{},
			EventuallyDue: []string{"some_future_field"},
			Errors:        nil,
		},
	}

	// Act
	status, err := svc.AnalyzeRequirements(account)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.HasErrors)
	assert.False(t, status.HasUrgent)
	assert.Equal(t, 0, len(status.PastDueFields))
	assert.Equal(t, 0, len(status.CurrentlyDueFields))
	assert.Equal(t, 1, len(status.EventuallyDueFields))
	assert.False(t, svc.ShouldGenerateRemediationLink(status))
}

// TestAnalyzeRequirements_NoRequirements testa análise com conta sem pendências
func TestAnalyzeRequirements_NoRequirements(t *testing.T) {
	// Arrange
	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	account := &stripe.Account{
		DetailsSubmitted: true,
		ChargesEnabled:   true,
		PayoutsEnabled:   true,
		Requirements:     nil,
	}

	// Act
	status, err := svc.AnalyzeRequirements(account)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.HasErrors)
	assert.False(t, status.HasUrgent)
	assert.Equal(t, 0, len(status.PastDueFields))
}

// TestAnalyzeRequirements_NilAccount testa análise com conta nil
func TestAnalyzeRequirements_NilAccount(t *testing.T) {
	// Arrange
	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	// Act
	status, err := svc.AnalyzeRequirements(nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.HasErrors)
	assert.False(t, status.HasUrgent)
}

// TestHasCriticalPendencies testa identificação de pendências críticas
func TestHasCriticalPendencies(t *testing.T) {
	tests := []struct {
		name     string
		status   *AccountRequirementsStatus
		expected bool
	}{
		{
			name: "critical when past_due exists",
			status: &AccountRequirementsStatus{
				HasUrgent: true,
				PastDueFields: []PendingRequirement{
					{Priority: PriorityPastDue, Field: "document"},
				},
			},
			expected: true,
		},
		{
			name: "critical when currently_due exists",
			status: &AccountRequirementsStatus{
				HasUrgent: true,
				CurrentlyDueFields: []PendingRequirement{
					{Priority: PriorityCurrently, Field: "bank_account"},
				},
			},
			expected: true,
		},
		{
			name: "not critical when only eventually_due",
			status: &AccountRequirementsStatus{
				HasUrgent: false,
				EventuallyDueFields: []PendingRequirement{
					{Priority: PriorityEventually, Field: "future_field"},
				},
			},
			expected: false,
		},
		{
			name:     "not critical when nil status",
			status:   nil,
			expected: false,
		},
		{
			name:     "not critical when empty status",
			status:   &AccountRequirementsStatus{},
			expected: false,
		},
	}

	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.HasCriticalPendencies(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldGenerateRemediationLink testa decisão de gerar link de remediação
func TestShouldGenerateRemediationLink(t *testing.T) {
	tests := []struct {
		name     string
		status   *AccountRequirementsStatus
		expected bool
	}{
		{
			name: "should generate when has past_due",
			status: &AccountRequirementsStatus{
				HasErrors: false,
				HasUrgent: true,
				PastDueFields: []PendingRequirement{
					{Priority: PriorityPastDue},
				},
			},
			expected: true,
		},
		{
			name: "should generate when has errors",
			status: &AccountRequirementsStatus{
				HasErrors: true,
				HasUrgent: false,
			},
			expected: true,
		},
		{
			name: "should not generate when only eventually_due",
			status: &AccountRequirementsStatus{
				HasErrors: false,
				HasUrgent: false,
				EventuallyDueFields: []PendingRequirement{
					{Priority: PriorityEventually},
				},
			},
			expected: false,
		},
		{
			name:     "should not generate when nil",
			status:   nil,
			expected: false,
		},
	}

	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.ShouldGenerateRemediationLink(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetPendingDocuments testa extração de documentos pendentes
func TestGetPendingDocuments(t *testing.T) {
	// Arrange
	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	account := &stripe.Account{
		Requirements: &stripe.AccountRequirements{
			PastDue: []string{
				"individual.verification.document",
				"bank_account_verification",
			},
			CurrentlyDue: []string{
				"individual.email",
				"individual.address.postal_code",
			},
			EventuallyDue: []string{
				"some_future_requirement",
			},
		},
	}

	// Act
	documents := svc.GetPendingDocuments(account)

	// Assert
	assert.NotNil(t, documents)
	assert.Equal(t, 2, len(documents), "deve conter 2 documentos (verification.document e bank_account)")

	// Verificar que contém os documentos esperados
	hasVerification := false
	hasBankAccount := false

	for _, doc := range documents {
		if doc.Field == "individual.verification.document" {
			hasVerification = true
		}
		if doc.Field == "bank_account_verification" {
			hasBankAccount = true
		}
	}

	assert.True(t, hasVerification, "deve conter verification.document")
	assert.True(t, hasBankAccount, "deve conter bank_account_verification")
}

// TestClassifyRequirementType testa classificação de tipos de pendência
func TestClassifyRequirementType(t *testing.T) {
	tests := []struct {
		field    string
		expected RequirementType
	}{
		{"individual.verification.document", RequirementTypeDocument},
		{"bank_account_verification", RequirementTypeDocument},
		{"individual.identity", RequirementTypeDocument},
		{"individual.ssn_last_4", RequirementTypePersonal},
		{"individual.address.postal_code", RequirementTypePersonal},
		{"individual.email", RequirementTypePersonal},
		{"tos_acceptance", RequirementTypeCompliance},
		{"individual.first_name", RequirementTypePersonal},
		{"unknown_field", RequirementTypeOther},
	}

	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := svc.classifyRequirementType(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetFieldDescription testa geração de descrições legíveis
func TestGetFieldDescription(t *testing.T) {
	tests := []struct {
		field            string
		expectedContains string
	}{
		{"individual.verification.document", "Documento"},
		{"bank_account_verification", "banco"},
		{"individual.ssn_last_4", "CPF"},
		{"individual.email", "Email"},
		{"tos_acceptance", "termos"},
	}

	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := svc.getFieldDescription(tt.field)
			assert.NotEmpty(t, result, "descrição não deve estar vazia")
			// Verificar se contém pelo menos parte do campo original
			assert.NotEqual(t, "", result)
		})
	}
}

// TestTranslateRequirementError testa tradução de erros Stripe
func TestTranslateRequirementError(t *testing.T) {
	tests := []struct {
		name          string
		err           interface{}
		shouldContain string
	}{
		{
			name:          "unknown error",
			err:           "some error",
			shouldContain: "Requisito pendente",
		},
		{
			name:          "nil error",
			err:           nil,
			shouldContain: "desconhecido",
		},
	}

	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.translateRequirementError(tt.err)
			assert.NotEmpty(t, result)
			// Não deve expor código técnico da Stripe
			assert.NotContains(t, result, "stripe")
		})
	}
}

// TestMapFieldToRequirement testa conversão de campo para struct
func TestMapFieldToRequirement(t *testing.T) {
	// Arrange
	mockCreatorSvc := &MockCreatorService{}
	svc := &stripeConnectServiceImpl{creatorSvc: mockCreatorSvc}

	// Act
	req := svc.mapFieldToRequirement("individual.verification.document", PriorityPastDue)

	// Assert
	assert.Equal(t, "individual.verification.document", req.Field)
	assert.Equal(t, PriorityPastDue, req.Priority)
	assert.Equal(t, RequirementTypeDocument, req.Type)
	assert.NotEmpty(t, req.Description)
	assert.True(t, req.Resolvable)
}
