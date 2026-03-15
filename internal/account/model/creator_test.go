package model_test

import (
	"testing"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	"github.com/stretchr/testify/assert"
)

func TestValidateFacebookPixelID(t *testing.T) {
	tests := []struct {
		name    string
		pixelID string
		wantErr bool
	}{
		{name: "vazio é válido (remove pixel)", pixelID: "", wantErr: false},
		{name: "15 dígitos válido", pixelID: "123456789012345", wantErr: false},
		{name: "10 dígitos válido (mínimo)", pixelID: "1234567890", wantErr: false},
		{name: "20 dígitos válido (máximo)", pixelID: "12345678901234567890", wantErr: false},
		{name: "9 dígitos muito curto", pixelID: "123456789", wantErr: true},
		{name: "21 dígitos muito longo", pixelID: "123456789012345678901", wantErr: true},
		{name: "contém letras", pixelID: "12345678901234a", wantErr: true},
		{name: "contém caracteres especiais", pixelID: "123456789<script>", wantErr: true},
		{name: "contém espaços", pixelID: "123456 789012345", wantErr: true},
		{name: "tentativa de XSS", pixelID: `<script>alert(1)</script>`, wantErr: true},
		{name: "tentativa de injeção JS", pixelID: "');alert(1);//", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := accountmodel.ValidateFacebookPixelID(tt.pixelID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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
