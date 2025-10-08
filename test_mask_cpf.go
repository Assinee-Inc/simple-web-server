package main

import (
	"fmt"
	"strings"
)

// maskCPF função para testar a máscara de CPF
func maskCPF(cpf string) string {
	// Remove todos os caracteres não numéricos
	cleanCPF := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(cpf, ".", ""), "-", ""), " ", "")

	// Verifica se tem 11 dígitos
	if len(cleanCPF) != 11 {
		return cpf // Retorna original se não for CPF válido
	}

	// Formata como 000.***-00
	return cleanCPF[:3] + ".***-" + cleanCPF[9:]
}

func main() {
	// Testes com diferentes formatos de CPF
	testCPFs := []string{
		"12345678901",
		"123.456.789-01",
		"123 456 789 01",
		"123.456.789.01",
		"12345678",     // CPF inválido
		"123456789012", // CPF inválido
	}

	fmt.Println("Teste da função maskCPF:")
	fmt.Println("========================")

	for _, cpf := range testCPFs {
		masked := maskCPF(cpf)
		fmt.Printf("Original: %-15s -> Mascarado: %s\n", cpf, masked)
	}
}
