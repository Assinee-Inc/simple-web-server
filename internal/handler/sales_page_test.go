package handler

import (
	"testing"

	"github.com/anglesson/simple-web-server/internal/mocks"
	"github.com/stretchr/testify/assert"
)

// Teste básico da criação do handler
func TestSalesPageHandler_Creation(t *testing.T) {
	// Setup
	mockEbookService := new(mocks.MockEbookService)
	mockCreatorService := new(mocks.MockCreatorService)
	mockTemplateRenderer := new(mocks.MockTemplateRenderer)

	// Act
	handler := NewSalesPageHandler(mockEbookService, mockCreatorService, mockTemplateRenderer)

	// Assert
	assert.NotNil(t, handler, "Handler deve ser criado")
	assert.NotNil(t, handler.ebookService, "EbookService deve ser injetado")
	assert.NotNil(t, handler.creatorService, "CreatorService deve ser injetado")
	assert.NotNil(t, handler.templateRenderer, "TemplateRenderer deve ser injetado")
}
