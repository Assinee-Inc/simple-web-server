package account

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccount_Validate_Success(t *testing.T) {
	testAccount := &Account{
		Name:      "John Doe",
		CPF:       "12345678900",
		Email:     "john.doe@example.com",
		Phone:     "11999999999",
		BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	err := testAccount.Validate()

	assert.NoError(t, err)
}

func TestAccount_ValidateName(t *testing.T) {
	tests := []struct {
		name        string
		accountName string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid name",
			accountName: "John Doe",
			shouldError: false,
		},
		{
			name:        "valid name with multiple words",
			accountName: "Mary Jane Watson Parker",
			shouldError: false,
		},
		{
			name:        "valid name with hyphen",
			accountName: "Mary-Jane Watson",
			shouldError: false,
		},
		{
			name:        "empty name",
			accountName: "",
			shouldError: true,
			errorMsg:    "nome não pode estar vazio",
		},
		{
			name:        "name too short",
			accountName: "Jo",
			shouldError: true,
			errorMsg:    "nome deve ter pelo menos 3 caracteres",
		},
		{
			name:        "name too long",
			accountName: string(make([]byte, 256)),
			shouldError: true,
			errorMsg:    "nome muito longo",
		},
		{
			name:        "name with numbers",
			accountName: "John123 Doe",
			shouldError: true,
			errorMsg:    "nome contém caracteres inválidos",
		},
		{
			name:        "name with special chars",
			accountName: "John@Doe",
			shouldError: true,
			errorMsg:    "nome contém caracteres inválidos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Name:      tt.accountName,
				CPF:       "12345678900",
				Email:     "test@example.com",
				Phone:     "11999999999",
				BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			}

			err := account.validateName()

			if tt.shouldError {
				assert.Error(t, err, "esperado erro para: %s", tt.name)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "não esperado erro para: %s", tt.name)
			}
		})
	}
}

func TestAccount_ValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid email",
			email:       "user@example.com",
			shouldError: false,
		},
		{
			name:        "valid email with plus",
			email:       "user+tag@example.com",
			shouldError: false,
		},
		{
			name:        "valid email with numbers",
			email:       "user123@example.co.uk",
			shouldError: false,
		},
		{
			name:        "empty email",
			email:       "",
			shouldError: true,
			errorMsg:    "email inválido",
		},
		{
			name:        "email without @",
			email:       "userexample.com",
			shouldError: true,
			errorMsg:    "email inválido",
		},
		{
			name:        "email without domain",
			email:       "user@",
			shouldError: true,
			errorMsg:    "email inválido",
		},
		{
			name:        "email without TLD",
			email:       "user@example",
			shouldError: true,
			errorMsg:    "email inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Name:      "Test User",
				CPF:       "12345678900",
				Email:     tt.email,
				Phone:     "11999999999",
				BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			}

			err := account.validateEmail()

			if tt.shouldError {
				assert.Error(t, err, "esperado erro para: %s", tt.name)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "não esperado erro para: %s", tt.name)
			}
		})
	}
}

func TestAccount_ValidateCPF(t *testing.T) {
	tests := []struct {
		name        string
		cpf         string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid CPF",
			cpf:         "12345678900",
			shouldError: false,
		},
		{
			name:        "valid CPF all zeros",
			cpf:         "00000000000",
			shouldError: false,
		},
		{
			name:        "valid CPF all nines",
			cpf:         "99999999999",
			shouldError: false,
		},
		{
			name:        "CPF too short",
			cpf:         "123456789",
			shouldError: true,
			errorMsg:    "CPF deve ter exatamente 11 dígitos",
		},
		{
			name:        "CPF too long",
			cpf:         "123456789001",
			shouldError: true,
			errorMsg:    "CPF deve ter exatamente 11 dígitos",
		},
		{
			name:        "CPF with letters",
			cpf:         "1234567890A",
			shouldError: true,
			errorMsg:    "CPF deve conter apenas dígitos",
		},
		{
			name:        "CPF with special chars",
			cpf:         "123.456.789-00",
			shouldError: true,
			errorMsg:    "CPF deve ter exatamente 11 dígitos",
		},
		{
			name:        "empty CPF",
			cpf:         "",
			shouldError: true,
			errorMsg:    "CPF deve ter exatamente 11 dígitos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Name:      "Test User",
				CPF:       tt.cpf,
				Email:     "test@example.com",
				Phone:     "11999999999",
				BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			}

			err := account.validateCPF()

			if tt.shouldError {
				assert.Error(t, err, "esperado erro para: %s", tt.name)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "não esperado erro para: %s", tt.name)
			}
		})
	}
}

func TestAccount_ValidatePhone(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid phone 11 digits",
			phone:       "11999999999",
			shouldError: false,
		},
		{
			name:        "valid phone 10 digits",
			phone:       "1199999999",
			shouldError: false,
		},
		{
			name:        "valid phone with formatting",
			phone:       "(11) 99999-9999",
			shouldError: false,
		},
		{
			name:        "phone too short",
			phone:       "119999999",
			shouldError: true,
			errorMsg:    "telefone inválido",
		},
		{
			name:        "phone too long",
			phone:       "119999999999",
			shouldError: true,
			errorMsg:    "telefone inválido",
		},
		{
			name:        "empty phone",
			phone:       "",
			shouldError: true,
			errorMsg:    "telefone inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Name:      "Test User",
				CPF:       "12345678900",
				Email:     "test@example.com",
				Phone:     tt.phone,
				BirthDate: time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			}

			err := account.validatePhone()

			if tt.shouldError {
				assert.Error(t, err, "esperado erro para: %s", tt.name)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "não esperado erro para: %s", tt.name)
			}
		})
	}
}

func TestAccount_ValidateBirthDate(t *testing.T) {
	tests := []struct {
		name        string
		birthDate   time.Time
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid age 18",
			birthDate:   time.Now().AddDate(-18, 0, 0),
			shouldError: false,
		},
		{
			name:        "valid age 30",
			birthDate:   time.Now().AddDate(-30, 0, 0),
			shouldError: false,
		},
		{
			name:        "valid age 80",
			birthDate:   time.Now().AddDate(-80, 0, 0),
			shouldError: false,
		},
		{
			name:        "age too young",
			birthDate:   time.Now().AddDate(-17, 0, 0),
			shouldError: true,
			errorMsg:    "deve ter pelo menos 18 anos",
		},
		{
			name:        "age too old",
			birthDate:   time.Now().AddDate(-121, 0, 0),
			shouldError: true,
			errorMsg:    "data de nascimento inválida",
		},
		{
			name:        "future date",
			birthDate:   time.Now().AddDate(0, 0, 1), // 1 dia no futuro
			shouldError: true,
			errorMsg:    "futuro",
		},
		{
			name:        "zero date",
			birthDate:   time.Time{},
			shouldError: true,
			errorMsg:    "data de nascimento não pode estar vazia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Name:      "Test User",
				CPF:       "12345678900",
				Email:     "test@example.com",
				Phone:     "11999999999",
				BirthDate: tt.birthDate,
			}

			err := account.validateBirthDate()

			if tt.shouldError {
				assert.Error(t, err, "esperado erro para: %s", tt.name)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "não esperado erro para: %s", tt.name)
			}
		})
	}
}

func TestAccount_SplitName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFirst string
		expectedLast  string
	}{
		{
			name:          "full name with multiple words",
			input:         "John Doe Smith",
			expectedFirst: "John",
			expectedLast:  "Doe Smith",
		},
		{
			name:          "first name only",
			input:         "John",
			expectedFirst: "John",
			expectedLast:  "",
		},
		{
			name:          "multiple spaces",
			input:         "Mary Jane Watson Parker",
			expectedFirst: "Mary",
			expectedLast:  "Jane Watson Parker",
		},
		{
			name:          "two names",
			input:         "John Doe",
			expectedFirst: "John",
			expectedLast:  "Doe",
		},
	}

	account := &Account{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account.Name = tt.input
			first, last := account.SplitName()

			assert.Equal(t, tt.expectedFirst, first)
			assert.Equal(t, tt.expectedLast, last)
		})
	}
}
