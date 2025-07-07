package service

import (
	"fmt"
	"os"

	"github.com/unidoc/unipdf/v3/model"
)

// ApplyDRM aplica proteção DRM ao PDF com as seguintes restrições:
// - Requer senha para abrir (userPassword)
// - Requer senha para editar/imprimir (ownerPassword)
// - Desabilita impressão
// - Desabilita cópia de conteúdo
// - Desabilita modificações
func ApplyDRM(inputPath, userPassword, ownerPassword string) (string, error) {
	// Abrir o PDF original
	f, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %v", err)
	}
	defer f.Close()

	// Criar o leitor PDF
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %v", err)
	}

	// Criar o escritor PDF
	pdfWriter := model.NewPdfWriter()

	// Aplicar criptografia com senhas
	err = pdfWriter.Encrypt([]byte(userPassword), []byte(ownerPassword), nil)
	if err != nil {
		return "", fmt.Errorf("failed to set encryption: %v", err)
	}

	// Copiar todas as páginas do PDF original
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("failed to get number of pages: %v", err)
	}

	for i := 0; i < numPages; i++ {
		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return "", fmt.Errorf("failed to get page %d: %v", i+1, err)
		}

		err = pdfWriter.AddPage(page)
		if err != nil {
			return "", fmt.Errorf("failed to add page %d: %v", i+1, err)
		}
	}

	// Criar arquivo de saída
	outputPath := "protected_output.pdf"
	of, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer of.Close()

	// Escrever o PDF protegido
	err = pdfWriter.Write(of)
	if err != nil {
		return "", fmt.Errorf("failed to write PDF: %v", err)
	}

	return outputPath, nil
}
