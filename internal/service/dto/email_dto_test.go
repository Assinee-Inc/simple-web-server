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

func TestFileDTO_GetFileSizeFormatted(t *testing.T) {
	tests := []struct {
		name     string
		fileDTO  FileDTO
		expected string
	}{
		{
			name: "Deve retornar tamanho formatado quando Size não é vazio",
			fileDTO: FileDTO{
				OriginalName: "arquivo.pdf",
				Size:         "2.5 MB",
			},
			expected: "2.5 MB",
		},
		{
			name: "Deve retornar mensagem padrão quando Size é vazio",
			fileDTO: FileDTO{
				OriginalName: "arquivo.pdf",
				Size:         "",
			},
			expected: "Tamanho desconhecido",
		},
		{
			name: "Deve funcionar com diferentes formatos de tamanho",
			fileDTO: FileDTO{
				OriginalName: "documento.docx",
				Size:         "1.8 KB",
			},
			expected: "1.8 KB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fileDTO.GetFileSizeFormatted()
			assert.Equal(t, tt.expected, result)
		})
	}
}
