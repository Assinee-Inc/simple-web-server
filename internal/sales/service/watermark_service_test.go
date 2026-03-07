package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFilename(t *testing.T) {
	// Teste com nome simples
	result := getFilename("test.pdf")
	assert.Contains(t, result, "test_")
	assert.Contains(t, result, ".pdf")
	assert.NotEqual(t, "test.pdf", result) // Deve ter timestamp

	// Teste com caminho complexo
	result = getFilename("files/1/Example Document-fce33251.pdf")
	assert.Contains(t, result, "Example Document-fce33251_")
	assert.Contains(t, result, ".pdf")
	assert.NotContains(t, result, "files/1/") // Não deve manter o caminho
}

func TestApplyWatermarkToLocalFile_FileNotFound(t *testing.T) {
	// Testar com arquivo que não existe
	outputPath, err := ApplyWatermarkToLocalFile("nonexistent.pdf", "Test Watermark", "test.pdf")

	// Deve retornar erro
	assert.Error(t, err)
	assert.Empty(t, outputPath)
}

func TestApplyWatermarkToLocalFile_EmptyContent(t *testing.T) {
	// Criar um arquivo de teste simples (não PDF)
	testFile := createTestFile(t)
	defer os.Remove(testFile)

	// Testar com conteúdo vazio
	outputPath, err := ApplyWatermarkToLocalFile(testFile, "", "test.pdf")

	// Deve retornar erro porque não é um PDF válido
	assert.Error(t, err)
	assert.Empty(t, outputPath)
}

// Função auxiliar para criar um arquivo de teste simples
func createTestFile(t *testing.T) string {
	// Criar diretório temp se não existir
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Erro ao criar diretório temp: %v", err)
	}

	// Criar arquivo de teste simples
	testFilePath := filepath.Join(tempDir, "test.txt")

	content := []byte("Test file content")

	err := os.WriteFile(testFilePath, content, 0644)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo de teste: %v", err)
	}

	return testFilePath
}

func TestApplyWatermark_Integration(t *testing.T) {
	// Este teste simula o fluxo completo
	// Primeiro criamos um arquivo local
	testFile := createTestFile(t)
	defer os.Remove(testFile)

	// Simulamos o que acontece quando baixamos do S3
	// (neste caso, usamos o arquivo local diretamente)
	outputPath, err := ApplyWatermarkToLocalFile(testFile, "Integration Test", "test.pdf")

	// Deve retornar erro porque não é um PDF válido
	assert.Error(t, err)
	assert.Empty(t, outputPath)
}
