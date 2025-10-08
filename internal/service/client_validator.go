package service

import "github.com/anglesson/simple-web-server/internal/models"

// validateClientInput validates all client input fields
func validateClientInput(input models.CreateClientInput) error {
	if err := validateName(input.Name); err != nil {
		return err
	}

	if err := validateEmail(input.Email); err != nil {
		return err
	}

	if err := validatePhone(input.Phone); err != nil {
		return err
	}

	if err := validateCPF(input.CPF); err != nil {
		return err
	}

	if err := validateBirthDate(input.BirthDate); err != nil {
		return err
	}

	return nil
}
