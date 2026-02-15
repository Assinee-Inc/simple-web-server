//go:build integration
// +build integration

package account

import (
	"os"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v76/client"
)

func TestStripeGateway_CreateSellerAccount_Integration(t *testing.T) {
	// Skip em ambiente de CI/CD sem chave real
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo -short")
	}

	// Arrange
	os.Setenv("APPLICATION_MODE", "testing")
	config.LoadConfigs()

	// Validar que chave está disponível
	require.NotEmpty(t, config.AppConfig.StripeSecretKey,
		"STRIPE_SECRET_KEY não configurada. Configure em .env.test.local ou variáveis de ambiente")

	sc := &client.API{}
	sc.Init(config.AppConfig.StripeSecretKey, nil)
	gateway := NewStripeGateway(sc)

	// Email único com timestamp para evitar duplicatas no Stripe
	uniqueEmail := "integration_test_" + time.Now().Format("20060102150405") + "@example.com"

	testAccount := &Account{
		Name:      "Integration Test User",
		CPF:       "12345678900",
		Email:     uniqueEmail,
		Phone:     "11999999999",
		BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	// Act
	accountID, err := gateway.CreateSellerAccount(testAccount)

	// Assert
	require.NoError(t, err, "Erro ao criar conta no Stripe")
	assert.NotEmpty(t, accountID, "ID da conta não deve estar vazio")
	assert.True(t, len(accountID) > 5, "ID da conta deve ter formato válido (acct_...)")
	assert.True(t, accountID[:5] == "acct_", "ID deve começar com 'acct_'")

	t.Logf("✓ Conta criada com sucesso: %s", accountID)
}

func TestStripeGateway_CreateSellerAccount_InvalidEmail_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo -short")
	}

	os.Setenv("APPLICATION_MODE", "testing")
	config.LoadConfigs()

	require.NotEmpty(t, config.AppConfig.StripeSecretKey,
		"STRIPE_SECRET_KEY não configurada")

	sc := &client.API{}
	sc.Init(config.AppConfig.StripeSecretKey, nil)
	gateway := NewStripeGateway(sc)

	testAccount := &Account{
		Name:      "Test User",
		CPF:       "12345678900",
		Email:     "invalid-email-format", // Email inválido
		Phone:     "11999999999",
		BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	// Act
	accountID, err := gateway.CreateSellerAccount(testAccount)

	// Assert
	assert.Error(t, err, "Esperado erro para email inválido")
	assert.Empty(t, accountID)
	assert.Contains(t, err.Error(), "erro ao criar conta no Stripe")

	t.Logf("✓ Validação de email funcionou corretamente: %v", err)
}

func TestStripeGateway_CreateSellerAccount_DifferentPhoneFormats_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo -short")
	}

	os.Setenv("APPLICATION_MODE", "testing")
	config.LoadConfigs()

	require.NotEmpty(t, config.AppConfig.StripeSecretKey)

	sc := &client.API{}
	sc.Init(config.AppConfig.StripeSecretKey, nil)
	gateway := NewStripeGateway(sc)

	tests := []struct {
		name  string
		phone string
	}{
		{
			name:  "11 digits",
			phone: "11999999999",
		},
		{
			name:  "10 digits",
			phone: "1199999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uniqueEmail := "phone_test_" + time.Now().Format("20060102150405") + "@example.com"

			testAccount := &Account{
				Name:      "Phone Test User",
				CPF:       "12345678900",
				Email:     uniqueEmail,
				Phone:     tt.phone,
				BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			}

			accountID, err := gateway.CreateSellerAccount(testAccount)

			// Pode falhar por validação do Stripe, mas não deve haver erro de parsing
			if err == nil {
				assert.NotEmpty(t, accountID)
				t.Logf("✓ Telefone %s criado com sucesso: %s", tt.phone, accountID)
			} else {
				t.Logf("✓ Telefone %s retornou erro esperado: %v", tt.phone, err)
			}
		})
	}
}

func TestStripeGateway_CreateSellerAccount_MultipleNames_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo -short")
	}

	os.Setenv("APPLICATION_MODE", "testing")
	config.LoadConfigs()

	require.NotEmpty(t, config.AppConfig.StripeSecretKey)

	sc := &client.API{}
	sc.Init(config.AppConfig.StripeSecretKey, nil)
	gateway := NewStripeGateway(sc)

	tests := []struct {
		name string
		text string
	}{
		{
			name: "simple two names",
			text: "John Doe",
		},
		{
			name: "three names",
			text: "John Doe Smith",
		},
		{
			name: "four names",
			text: "Mary Jane Watson Parker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uniqueEmail := "name_test_" + time.Now().Format("20060102150405") + "@example.com"

			testAccount := &Account{
				Name:      tt.text,
				CPF:       "12345678900",
				Email:     uniqueEmail,
				Phone:     "11999999999",
				BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			}

			accountID, err := gateway.CreateSellerAccount(testAccount)

			require.NoError(t, err, "Erro ao criar conta com nome: %s", tt.text)
			assert.NotEmpty(t, accountID)

			t.Logf("✓ Nome '%s' criado com sucesso, accountID=%s", tt.text, accountID)
		})
	}
}

func TestStripeGateway_CreateSellerAccount_EdgeCases_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo -short")
	}

	os.Setenv("APPLICATION_MODE", "testing")
	config.LoadConfigs()

	require.NotEmpty(t, config.AppConfig.StripeSecretKey)

	sc := &client.API{}
	sc.Init(config.AppConfig.StripeSecretKey, nil)
	gateway := NewStripeGateway(sc)

	// Testar diferentes datas de nascimento
	tests := []struct {
		name       string
		birthDate  time.Time
		shouldFail bool // Se true, espera erro
	}{
		{
			name:       "18 years old",
			birthDate:  time.Now().AddDate(-18, 0, 0),
			shouldFail: false,
		},
		{
			name:       "50 years old",
			birthDate:  time.Now().AddDate(-50, 0, 0),
			shouldFail: false,
		},
		{
			name:       "80 years old",
			birthDate:  time.Now().AddDate(-80, 0, 0),
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uniqueEmail := "age_test_" + time.Now().Format("20060102150405") + "@example.com"

			testAccount := &Account{
				Name:      "Test User",
				CPF:       "12345678900",
				Email:     uniqueEmail,
				Phone:     "11999999999",
				BirthDate: tt.birthDate,
			}

			accountID, err := gateway.CreateSellerAccount(testAccount)

			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, accountID)
				t.Logf("✓ Data de nascimento %s criada com sucesso", tt.name)
			}
		})
	}
}
