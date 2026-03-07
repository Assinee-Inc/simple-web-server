package service

import (
	"os"
	"path/filepath"
	"testing"
)

// changeToProjectRoot é uma função helper para testes que precisam acessar
// recursos do projeto (como templates) que estão localizados na raiz.
func changeToProjectRoot(t *testing.T) func() {
	t.Helper()

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Navegar para o diretório raiz do projeto (3 níveis acima de internal/sales/service)
	projectRoot := filepath.Join(originalWd, "..", "..", "..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	templateDir := filepath.Join("web", "mails")
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		t.Fatalf("Template directory not found at %s. Working directory: %s", templateDir, projectRoot)
	}

	return func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}
}
