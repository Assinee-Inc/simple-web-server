package service

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

// validateUserInput valida os campos básicos de usuário (nome e e-mail), sem senha.
func validateUserInput(input InputCreateUser) error {
	if err := validateUsername(input.Username); err != nil {
		return err
	}

	if err := validateEmail(input.Email); err != nil {
		return err
	}

	return nil
}

// validateUsername validates the username
func validateUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return errors.New("nome de usuário é obrigatório")
	}

	if len(username) > 100 {
		return errors.New("nome de usuário muito longo (máximo 100 caracteres)")
	}

	return nil
}

// validateEmail validates the email format
func validateEmail(value string) error {
	value = strings.TrimSpace(strings.ToLower(value))

	addr, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("formato de e-mail inválido: %w", err)
	}

	if len(addr.Address) > 254 {
		return fmt.Errorf("endereço de e-mail muito longo")
	}

	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 {
		return fmt.Errorf("formato de e-mail inválido")
	}

	if len(parts[0]) > 64 {
		return fmt.Errorf("parte local do e-mail muito longa")
	}

	if len(parts[1]) > 255 {
		return fmt.Errorf("domínio do e-mail muito longo")
	}

	return nil
}
