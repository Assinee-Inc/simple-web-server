package service

import (
	"os"
	"path/filepath"
	"testing"
)

// changeToProjectRoot é uma função helper para testes que precisam acessar
// recursos do projeto (como templates) que estão localizados na raiz.
//
// Problema: Quando executamos `go test`, o working directory é o diretório
// do pacote sendo testado (ex: internal/service), mas recursos como templates
// estão em web/mails/ na raiz do projeto.
//
// Solução: Esta função temporarily muda o working directory para a raiz
// do projeto durante o teste, e restaura o diretório original no final.
//
// Uso:
//
//	func TestMyFunction(t *testing.T) {
//	    cleanup := changeToProjectRoot(t)
//	    defer cleanup()
//	    // Agora você pode acessar web/mails/template.html
//	}
//
// A função retorna uma função cleanup que DEVE ser chamada com defer.
func changeToProjectRoot(t *testing.T) func() {
	t.Helper() // Marca como função helper para melhor stack trace

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Navegar para o diretório raiz do projeto (2 níveis acima de internal/service)
	projectRoot := filepath.Join(originalWd, "..", "..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}

	// Verificar se conseguimos acessar os templates (validação básica)
	templateDir := filepath.Join("web", "mails")
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		t.Fatalf("Template directory not found at %s. Working directory: %s", templateDir, projectRoot)
	}

	// Retornar função cleanup que restaura o diretório original
	return func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}
}
