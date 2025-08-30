package service

import (
	"errors"

	"github.com/anglesson/simple-web-server/internal/models"
)

// validateUserInput validates all user input fields
func validateUserInput(input models.InputCreateUser) error {
	if err := validateUsername(input.Username); err != nil {
		return err
	}

	if err := validateEmail(input.Email); err != nil {
		return err
	}

	if err := validatePassword(input.Password); err != nil {
		return err
	}

	if input.Password != input.PasswordConfirmation {
		return errors.New("as senhas não coincidem")
	}

	return nil
}

// validateUsername validates the username
func validateUsername(username string) error {
	if username == "" {
		return errors.New("nome de usuário é obrigatório")
	}

	if len(username) > 50 {
		return errors.New("nome de usuário muito longo (máximo 50 caracteres)")
	}

	return nil
}

// validatePassword validates the password
func validatePassword(password string) error {
	if password == "" {
		return errors.New("senha é obrigatória")
	}

	if len(password) < 8 {
		return errors.New("a senha deve ter pelo menos 8 caracteres")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*' || char == '(' || char == ')' || char == '-' || char == '_' || char == '+' || char == '=' || char == '[' || char == ']' || char == '{' || char == '}' || char == '|' || char == '\\' || char == ':' || char == ';' || char == '"' || char == '\'' || char == '<' || char == '>' || char == ',' || char == '.' || char == '?' || char == '/':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("a senha deve conter pelo menos uma letra maiúscula")
	}

	if !hasLower {
		return errors.New("a senha deve conter pelo menos uma letra minúscula")
	}

	if !hasDigit {
		return errors.New("a senha deve conter pelo menos um dígito")
	}

	if !hasSpecial {
		return errors.New("a senha deve conter pelo menos um caractere especial")
	}

	return nil
}
