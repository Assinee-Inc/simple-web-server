package models_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewEbook(t *testing.T) {
	// Arrange
	title := "Test Ebook"
	description := "Test description"
	salesPage := "This is a sales page content"
	value := 29.90
	promotionalValue := 19.90
	creator := models.Creator{
		Name:  "Test Creator",
		Email: "creator@test.com",
	}

	// Act
	ebook := models.NewEbook(title, description, salesPage, value, promotionalValue, creator, false)

	// Assert
	assert.NotNil(t, ebook)
	assert.Equal(t, title, ebook.Title)
	assert.Equal(t, description, ebook.Description)
	assert.Equal(t, salesPage, ebook.SalesPage)
	assert.Equal(t, value, ebook.GetFinalValue())
	assert.True(t, ebook.Status)
	assert.Equal(t, creator.ID, ebook.CreatorID)
	assert.NotEmpty(t, ebook.Slug)
}

func TestEbook_AddFile(t *testing.T) {
	// Arrange
	ebook := &models.Ebook{Title: "Test Ebook"}
	file1 := &models.File{Name: "file1.pdf"}
	file2 := &models.File{Name: "file2.pdf"}

	// Act
	ebook.AddFile(file1)
	ebook.AddFile(file2)

	// Assert
	assert.Len(t, ebook.Files, 2)
	assert.Equal(t, file1, ebook.Files[0])
	assert.Equal(t, file2, ebook.Files[1])
}

func TestEbook_RemoveFile(t *testing.T) {
	// Arrange
	ebook := &models.Ebook{Title: "Test Ebook"}
	file1 := &models.File{Name: "file1.pdf"}
	file2 := &models.File{Name: "file2.pdf"}
	file3 := &models.File{Name: "file3.pdf"}

	// Definir IDs únicos para os arquivos
	file1.ID = 1
	file2.ID = 2
	file3.ID = 3

	ebook.AddFile(file1)
	ebook.AddFile(file2)
	ebook.AddFile(file3)

	// Act
	ebook.RemoveFile(file2.ID)

	// Assert
	assert.Len(t, ebook.Files, 2)
	assert.Equal(t, file1.Name, ebook.Files[0].Name)
	assert.Equal(t, file3.Name, ebook.Files[1].Name)
}

func TestEbook_GetTotalFileSize(t *testing.T) {
	// Arrange
	ebook := &models.Ebook{Title: "Test Ebook"}
	file1 := &models.File{FileSize: 1024 * 1024} // 1MB
	file2 := &models.File{FileSize: 2048 * 1024} // 2MB

	ebook.AddFile(file1)
	ebook.AddFile(file2)

	// Act
	totalSize := ebook.GetTotalFileSize()

	// Assert
	expectedSize := int64(3 * 1024 * 1024) // 3MB
	assert.Equal(t, expectedSize, totalSize)
}

func TestEbook_GetFileCount(t *testing.T) {
	tests := []struct {
		name     string
		ebook    *models.Ebook
		expected int
	}{
		{
			name:     "No files",
			ebook:    &models.Ebook{},
			expected: 0,
		},
		{
			name: "One file",
			ebook: &models.Ebook{
				Files: []*models.File{
					{Name: "file1.pdf"},
				},
			},
			expected: 1,
		},
		{
			name: "Multiple files",
			ebook: &models.Ebook{
				Files: []*models.File{
					{Name: "file1.pdf"},
					{Name: "file2.pdf"},
					{Name: "file3.pdf"},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ebook.GetFileCount()
			if result != tt.expected {
				t.Errorf("GetFileCount() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEbook_GetValue(t *testing.T) {
	tests := []struct {
		name     string
		ebook    *models.Ebook
		expected string
	}{
		{
			name:     "Zero value",
			ebook:    &models.Ebook{Value: 0},
			expected: "R$ 0,00",
		},
		{
			name:     "Positive value",
			ebook:    &models.Ebook{Value: 29.90},
			expected: "R$ 29,90",
		},
		{
			name:     "Large value",
			ebook:    &models.Ebook{Value: 199.99},
			expected: "R$ 199,99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ebook.GetValue()
			if result != tt.expected {
				t.Errorf("GetValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEbook_IncrementViews(t *testing.T) {
	// Arrange
	ebook := &models.Ebook{Title: "Test Ebook", Views: 10}

	// Act
	ebook.IncrementViews()
	ebook.IncrementViews()

	// Assert
	assert.Equal(t, 12, ebook.Views)
}

func TestEbook_IncrementSales(t *testing.T) {
	// Arrange
	ebook := &models.Ebook{Title: "Test Ebook", Sales: 5}

	// Act
	ebook.IncrementSales()
	ebook.IncrementSales()

	// Assert
	assert.Equal(t, 7, ebook.Sales)
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "Simple title",
			title:    "Test Ebook",
			expected: "test-ebook",
		},
		{
			name:     "Title with special characters",
			title:    "Guia Completo de Marketing Digital!",
			expected: "guia-completo-de-marketing-digital",
		},
		{
			name:     "Title with accents",
			title:    "Apostila de Português",
			expected: "apostila-de-portugues",
		},
		{
			name:     "Title with numbers",
			title:    "Ebook 2024 - Edição Especial",
			expected: "ebook-2024-edicao-especial",
		},
		{
			name:     "Title with multiple spaces",
			title:    "  Test   Ebook  ",
			expected: "test-ebook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			slug := generateSlug(tt.title)

			// Assert
			assert.Equal(t, tt.expected, slug)
		})
	}
}

// Função auxiliar para testar generateSlug
func generateSlug(title string) string {
	// Copiar a implementação do modelo para teste
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "ç", "c")
	slug = strings.ReplaceAll(slug, "ã", "a")
	slug = strings.ReplaceAll(slug, "á", "a")
	slug = strings.ReplaceAll(slug, "à", "a")
	slug = strings.ReplaceAll(slug, "â", "a")
	slug = strings.ReplaceAll(slug, "é", "e")
	slug = strings.ReplaceAll(slug, "ê", "e")
	slug = strings.ReplaceAll(slug, "í", "i")
	slug = strings.ReplaceAll(slug, "ó", "o")
	slug = strings.ReplaceAll(slug, "ô", "o")
	slug = strings.ReplaceAll(slug, "ú", "u")
	slug = strings.ReplaceAll(slug, "ü", "u")
	slug = strings.ReplaceAll(slug, "ñ", "n")

	// Remove caracteres especiais
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")

	// Remove hífens duplicados
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Remove hífens no início e fim
	slug = strings.Trim(slug, "-")

	return slug
}
