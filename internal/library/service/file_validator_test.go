package service_test

import (
	"testing"

	librarysvc "github.com/Assinee-Inc/docffy/internal/library/service"
	"github.com/stretchr/testify/assert"
)

func TestValidatePresignedUploadRequest(t *testing.T) {
	tests := []struct {
		name        string
		req         librarysvc.PresignedUploadRequest
		expectError bool
		errMsg      string
	}{
		{
			name: "Arquivo PDF válido",
			req: librarysvc.PresignedUploadRequest{
				FileName:    "manual_do_usuario.pdf",
				FileSize:    1024 * 1024, // 1MB
				ContentType: "application/pdf",
			},
			expectError: false,
		},
		{
			name: "Nome de arquivo vazio",
			req: librarysvc.PresignedUploadRequest{
				FileName:    "",
				FileSize:    1024,
				ContentType: "application/pdf",
			},
			expectError: true,
			errMsg:      "Apenas arquivos PDF são permitidos ou tamanho excede 100MB",
		},
		{
			name: "Extensão inválida (não-PDF)",
			req: librarysvc.PresignedUploadRequest{
				FileName:    "virus.exe",
				FileSize:    1024,
				ContentType: "application/x-msdownload",
			},
			expectError: true,
			errMsg:      "Apenas arquivos PDF são permitidos ou tamanho excede 100MB",
		},
		{
			name: "Tamanho menor ou igual a zero",
			req: librarysvc.PresignedUploadRequest{
				FileName:    "documento.pdf",
				FileSize:    0,
				ContentType: "application/pdf",
			},
			expectError: true,
			errMsg:      "Apenas arquivos PDF são permitidos ou tamanho excede 100MB",
		},
		{
			name: "Tamanho negativo",
			req: librarysvc.PresignedUploadRequest{
				FileName:    "documento.pdf",
				FileSize:    -10,
				ContentType: "application/pdf",
			},
			expectError: true,
			errMsg:      "Apenas arquivos PDF são permitidos ou tamanho excede 100MB",
		},
		{
			name: "Tamanho superior a 100MB",
			req: librarysvc.PresignedUploadRequest{
				FileName:    "arquivo_pesado.pdf",
				FileSize:    104857601, // 100MB + 1 byte
				ContentType: "application/pdf",
			},
			expectError: true,
			errMsg:      "Apenas arquivos PDF são permitidos ou tamanho excede 100MB",
		},
		{
			name: "ContentType diferente de application/pdf",
			req: librarysvc.PresignedUploadRequest{
				FileName:    "relatorio.pdf",
				FileSize:    2048,
				ContentType: "text/plain",
			},
			expectError: true,
			errMsg:      "Apenas arquivos PDF são permitidos ou tamanho excede 100MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := librarysvc.ValidatePresignedUploadRequest(tt.req)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Nome normal",
			input:    "apostila.pdf",
			expected: "apostila.pdf",
		},
		{
			name:     "Nome com caracteres especiais e espaços",
			input:    "meu arquivo @ 2026! #1.pdf",
			expected: "meu_arquivo___2026___1.pdf",
		},
		{
			name:     "Caminho com directory traversal",
			input:    "../../secret.pdf",
			expected: "secret.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := librarysvc.SanitizeFileName(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
