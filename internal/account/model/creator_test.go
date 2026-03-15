package model_test

import (
	"testing"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	"github.com/stretchr/testify/assert"
)

func TestCreator_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		creator  accountmodel.Creator
		expected string
	}{
		{
			name:     "retorna SocialName quando preenchido",
			creator:  accountmodel.Creator{Name: "João Silva Santos", SocialName: "João das Letras"},
			expected: "João das Letras",
		},
		{
			name:     "retorna primeiro e último nome quando SocialName vazio",
			creator:  accountmodel.Creator{Name: "João Silva Santos"},
			expected: "João Santos",
		},
		{
			name:     "retorna o nome inteiro quando há apenas um token",
			creator:  accountmodel.Creator{Name: "Monônimo"},
			expected: "Monônimo",
		},
		{
			name:     "retorna ambos os nomes quando há exatamente dois tokens",
			creator:  accountmodel.Creator{Name: "João Santos"},
			expected: "João Santos",
		},
		{
			name:     "SocialName tem precedência mesmo com nome completo",
			creator:  accountmodel.Creator{Name: "Maria Aparecida Ferreira", SocialName: "Mari F."},
			expected: "Mari F.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.creator.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}
