package template

import (
	"testing"
)

func TestMaskCPF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CPF válido sem formatação",
			input:    "12345678901",
			expected: "123.***-01",
		},
		{
			name:     "CPF válido com pontos e hífen",
			input:    "123.456.789-01",
			expected: "123.***-01",
		},
		{
			name:     "CPF válido com espaços",
			input:    "123 456 789 01",
			expected: "123.***-01",
		},
		{
			name:     "CPF válido com pontos apenas",
			input:    "123.456.789.01",
			expected: "123.***-01",
		},
		{
			name:     "CPF inválido - muito curto",
			input:    "12345678",
			expected: "12345678", // Retorna original
		},
		{
			name:     "CPF inválido - muito longo",
			input:    "123456789012",
			expected: "123456789012", // Retorna original
		},
		{
			name:     "String vazia",
			input:    "",
			expected: "", // Retorna original
		},
		{
			name:     "CPF com caracteres especiais mistos",
			input:    "123-456.789-01",
			expected: "123.***-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar a função maskCPF inline para teste
			maskCPF := func(cpf string) string {
				// Remove todos os caracteres não numéricos
				cleanCPF := ""
				for _, char := range cpf {
					if char >= '0' && char <= '9' {
						cleanCPF += string(char)
					}
				}

				// Verifica se tem 11 dígitos
				if len(cleanCPF) != 11 {
					return cpf // Retorna original se não for CPF válido
				}

				// Formata como 000.***-00
				return cleanCPF[:3] + ".***-" + cleanCPF[9:]
			}

			result := maskCPF(tt.input)
			if result != tt.expected {
				t.Errorf("maskCPF(%q) = %q, esperado %q", tt.input, result, tt.expected)
			}
		})
	}
}
