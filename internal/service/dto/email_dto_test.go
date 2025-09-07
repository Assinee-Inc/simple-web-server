package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResendDownloadLinkDTO_Validate_Success(t *testing.T) {
	// Arrange
	dto := &ResendDownloadLinkDTO{
		ClientName:  "João Silva",
		ClientEmail: "joao@teste.com",
		EbookTitle:  "Ebook de Teste",
		EbookFiles: []FileDTO{
			{OriginalName: "arquivo1.pdf", Size: "2.5 MB"},
		},
		DownloadLink: "https://example.com/download/123",
		AppName:      "MeuApp",
		ContactEmail: "contato@exemplo.com",
	}

	// Act
	err := dto.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestResendDownloadLinkDTO_Validate_EmptyEmail(t *testing.T) {
	// Arrange
	dto := &ResendDownloadLinkDTO{
		ClientName:  "João Silva",
		ClientEmail: "", // Email vazio
		EbookTitle:  "Ebook de Teste",
		EbookFiles: []FileDTO{
			{OriginalName: "arquivo1.pdf", Size: "2.5 MB"},
		},
		DownloadLink: "https://example.com/download/123",
		AppName:      "MeuApp",
		ContactEmail: "contato@exemplo.com",
	}

	// Act
	err := dto.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email do cliente é obrigatório")
}

func TestResendDownloadLinkDTO_Validate_NoFiles(t *testing.T) {
	// Arrange
	dto := &ResendDownloadLinkDTO{
		ClientName:   "João Silva",
		ClientEmail:  "joao@teste.com",
		EbookTitle:   "Ebook de Teste",
		EbookFiles:   []FileDTO{}, // Sem arquivos
		DownloadLink: "https://example.com/download/123",
		AppName:      "MeuApp",
		ContactEmail: "contato@exemplo.com",
	}

	// Act
	err := dto.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ebook deve ter pelo menos um arquivo")
}
