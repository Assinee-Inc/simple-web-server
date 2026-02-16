package account

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Model) BeforeCreate(*gorm.DB) error {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	m.ID = newUUID
	return nil
}

type Account struct {
	Model
	Name                 string    `json:"name"`
	CPF                  string    `json:"cpf"`
	Email                string    `json:"email"`
	Phone                string    `json:"phone"`
	BirthDate            time.Time `json:"birth_date"`
	UserID               uuid.UUID `json:"user_id"`
	Origin               string    `json:"origin,omitempty"`
	ExternalAccountID    string    `json:"external_account_id"`
	OnboardingCompleted  bool      `json:"onboarding_completed" gorm:"default:false"`
	OnboardingRefreshURL string    `json:"onboarding_refresh_url"`
	OnboardingReturnURL  string    `json:"onboarding_return_url"`
	PayoutsEnabled       bool      `json:"payouts_enabled" gorm:"default:false"`
	ChargesEnabled       bool      `json:"charges_enabled" gorm:"default:false"`
}

// Validate realiza validações de segurança em todos os campos
func (a *Account) Validate() error {
	if err := a.validateName(); err != nil {
		return err
	}
	if err := a.validateEmail(); err != nil {
		return err
	}
	if err := a.validateCPF(); err != nil {
		return err
	}
	if err := a.validatePhone(); err != nil {
		return err
	}
	if err := a.validateBirthDate(); err != nil {
		return err
	}
	return nil
}

// validateName valida o nome do usuário
func (a *Account) validateName() error {
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("nome não pode estar vazio")
	}
	if len(a.Name) > 255 {
		return fmt.Errorf("nome muito longo (máximo 255 caracteres)")
	}
	if len(a.Name) < 3 {
		return fmt.Errorf("nome deve ter pelo menos 3 caracteres")
	}
	// Verificar se contém apenas letras, espaços e alguns caracteres especiais
	for _, r := range a.Name {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) && r != '-' && r != '\'' {
			return fmt.Errorf("nome contém caracteres inválidos")
		}
	}
	return nil
}

// validateEmail valida o formato do email
func (a *Account) validateEmail() error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(a.Email) {
		return fmt.Errorf("email inválido")
	}
	if len(a.Email) > 254 {
		return fmt.Errorf("email muito longo (máximo 254 caracteres)")
	}
	return nil
}

// validateCPF valida se o CPF tem 11 dígitos
func (a *Account) validateCPF() error {
	if len(a.CPF) != 11 {
		return fmt.Errorf("CPF deve ter exatamente 11 dígitos")
	}
	for _, ch := range a.CPF {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("CPF deve conter apenas dígitos")
		}
	}
	return nil
}

// validatePhone valida o telefone
func (a *Account) validatePhone() error {
	// Remover caracteres especiais e contar apenas dígitos
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, a.Phone)

	if len(cleaned) < 10 || len(cleaned) > 11 {
		return fmt.Errorf("telefone inválido (deve ter 10 ou 11 dígitos)")
	}
	return nil
}

// validateBirthDate valida a data de nascimento
func (a *Account) validateBirthDate() error {
	if a.BirthDate.IsZero() {
		return fmt.Errorf("data de nascimento não pode estar vazia")
	}

	today := time.Now()

	// Verificar primeiro se a data é no futuro
	if a.BirthDate.After(today) {
		return fmt.Errorf("data de nascimento não pode ser no futuro")
	}

	age := today.Year() - a.BirthDate.Year()

	// Ajustar idade se ainda não fez aniversário neste ano
	if today.Month() < a.BirthDate.Month() ||
		(today.Month() == a.BirthDate.Month() && today.Day() < a.BirthDate.Day()) {
		age--
	}

	if age < 18 {
		return fmt.Errorf("deve ter pelo menos 18 anos")
	}
	if age > 120 {
		return fmt.Errorf("data de nascimento inválida")
	}

	return nil
}

// splitName divides a full name into first and last name. It assumes the first word is the first name and the rest is the last name.
func (s *Account) SplitName() (string, string) {
	names := strings.Fields(s.Name)
	firstName := names[0]
	lastName := ""
	if len(names) > 1 {
		lastName = strings.Join(names[1:], " ")
	}
	return firstName, lastName
}
