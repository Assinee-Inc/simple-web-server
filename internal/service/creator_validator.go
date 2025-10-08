package service

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/internal/models"
)

// validateCreatorInput validates all creator input fields
func validateCreatorInput(input models.InputCreateCreator) error {
	if err := validateName(input.Name); err != nil {
		return err
	}

	if err := validateEmail(input.Email); err != nil {
		return err
	}

	if err := validatePhone(input.PhoneNumber); err != nil {
		return err
	}

	if err := validateCPF(input.CPF); err != nil {
		return err
	}

	if err := validateBirthDate(input.BirthDate); err != nil {
		return err
	}

	if err := validateTermsAccepted(input.TermsAccepted); err != nil {
		return err
	}

	return nil
}

// validateName validates the creator name
func validateName(name string) error {
	if name == "" {
		return errors.New("nome inválido")
	}

	if len(name) > 255 {
		return errors.New("nome muito longo")
	}

	return nil
}

// validateEmail validates the email format
func validateEmail(value string) error {
	// Trim spaces and convert to lowercase
	value = strings.TrimSpace(strings.ToLower(value))

	// Validate email format
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("formato de e-mail inválido: %w", err)
	}

	// Additional validation rules
	if len(addr.Address) > 254 { // RFC 5321
		return fmt.Errorf("endereço de e-mail muito longo")
	}

	// Split email into local and domain parts
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 {
		return fmt.Errorf("formato de e-mail inválido")
	}

	localPart := parts[0]
	domain := parts[1]

	// Validate local part length (RFC 5321)
	if len(localPart) > 64 {
		return fmt.Errorf("parte local do e-mail muito longa")
	}

	// Validate domain length (RFC 5321)
	if len(domain) > 255 {
		return fmt.Errorf("domínio do e-mail muito longo")
	}

	return nil
}

// validatePhone validates the phone number format
func validatePhone(value string) error {
	// Clean the phone number (remove all non-digit characters)
	cleanNumber := cleanPhone(value)

	// Check if the phone number has the correct length
	if len(cleanNumber) != 11 {
		return fmt.Errorf("número de telefone inválido: deve ter 11 dígitos")
	}

	// Validate area code (DDD)
	areaCode := cleanNumber[0:2]
	if !isValidAreaCode(areaCode) {
		return fmt.Errorf("DDD inválido: %s", areaCode)
	}

	// Validate if it's a mobile number (must start with 9)
	if cleanNumber[2] != '9' {
		return fmt.Errorf("número de telefone inválido: celulares devem começar com 9")
	}

	// Validate if all digits are the same (invalid number)
	if allPhoneDigitsSame(cleanNumber) {
		return fmt.Errorf("número de telefone inválido: todos os dígitos são iguais")
	}

	return nil
}

// validateCPF validates the CPF format
func validateCPF(value string) error {
	cpf := cleanCPF(value)

	// Check if CPF has 11 digits
	if len(cpf) != 11 {
		return fmt.Errorf("CPF inválido: deve ter 11 dígitos")
	}

	// Check if all digits are the same (invalid CPF)
	if allDigitsSame(cpf) {
		return fmt.Errorf("CPF inválido: todos os dígitos são iguais")
	}

	// Validate first digit
	digit1 := calculateDigit(cpf[:9], 10)
	if digit1 != int(cpf[9]-'0') {
		return fmt.Errorf("CPF inválido: primeiro dígito verificador incorreto")
	}

	// Validate second digit
	digit2 := calculateDigit(cpf[:10], 11)
	if digit2 != int(cpf[10]-'0') {
		return fmt.Errorf("CPF inválido: segundo dígito verificador incorreto")
	}

	return nil
}

// validateBirthDate validates the birth date
func validateBirthDate(birthDateStr string) error {
	// First try to parse as DD/MM/YYYY format (from jmask)
	parsedDate, err := time.Parse("02/01/2006", birthDateStr)
	if err != nil {
		// If that fails, try YYYY-MM-DD format (from HTML date input)
		parsedDate, err = time.Parse("2006-01-02", birthDateStr)
		if err != nil {
			return fmt.Errorf("formato de data de nascimento inválido: %w", err)
		}
	}

	year := parsedDate.Year()
	month := int(parsedDate.Month())
	day := parsedDate.Day()

	// Validate year
	currentYear := time.Now().Year()
	if year < 1900 || year > currentYear {
		return fmt.Errorf("ano inválido: deve estar entre 1900 e %d", currentYear)
	}

	// Validate if the date is valid (e.g., February 30th would be invalid)
	if parsedDate.Year() != year || int(parsedDate.Month()) != month || parsedDate.Day() != day {
		return fmt.Errorf("data inválida: %d-%02d-%02d", year, month, day)
	}

	// Validate if the date is not in the future
	if parsedDate.After(time.Now()) {
		return fmt.Errorf("data de nascimento não pode ser no futuro")
	}

	// Validate if person is adult (18 years or older)
	age := time.Now().Year() - year
	if time.Now().Month() < parsedDate.Month() || (time.Now().Month() == parsedDate.Month() && time.Now().Day() < parsedDate.Day()) {
		age--
	}

	if age < 18 {
		return errors.New("o criador deve ter 18 anos ou mais")
	}

	return nil
}

// Helper functions for CPF validation
func cleanCPF(cpf string) string {
	re := regexp.MustCompile(`[^\d]`)
	return re.ReplaceAllString(cpf, "")
}

func allDigitsSame(cpf string) bool {
	first := cpf[0]
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != first {
			return false
		}
	}
	return true
}

func calculateDigit(cpf string, factor int) int {
	var sum int
	for _, digit := range cpf {
		num, _ := strconv.Atoi(string(digit))
		sum += num * factor
		factor--
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}

// Helper functions for phone validation
func cleanPhone(phone string) string {
	re := regexp.MustCompile(`[^\d]`)
	return re.ReplaceAllString(phone, "")
}

func isValidAreaCode(areaCode string) bool {
	// List of valid Brazilian area codes
	validAreaCodes := map[string]bool{
		"11": true, "12": true, "13": true, "14": true, "15": true, "16": true, "17": true, "18": true, "19": true,
		"21": true, "22": true, "24": true, "27": true, "28": true,
		"31": true, "32": true, "33": true, "34": true, "35": true, "37": true, "38": true,
		"41": true, "42": true, "43": true, "44": true, "45": true, "46": true, "47": true, "48": true, "49": true,
		"51": true, "53": true, "54": true, "55": true,
		"61": true, "62": true, "63": true, "64": true, "65": true, "66": true, "67": true, "68": true, "69": true,
		"71": true, "73": true, "74": true, "75": true, "77": true, "79": true,
		"81": true, "82": true, "83": true, "84": true, "85": true, "86": true, "87": true, "88": true, "89": true,
		"91": true, "92": true, "93": true, "94": true, "95": true, "96": true, "97": true, "98": true, "99": true,
	}

	return validAreaCodes[areaCode]
}

func allPhoneDigitsSame(phone string) bool {
	first := phone[0]
	for i := 1; i < len(phone); i++ {
		if phone[i] != first {
			return false
		}
	}
	return true
}

// validateTermsAccepted validates that the user accepted the terms
func validateTermsAccepted(termsAccepted string) error {
	if termsAccepted != "on" {
		return errors.New("você deve aceitar os Termos de Serviço e Política de Privacidade")
	}
	return nil
}
