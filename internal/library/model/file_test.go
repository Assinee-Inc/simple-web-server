package model_test

import (
	"testing"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
)

func TestFile_GetFileSizeFormatted(t *testing.T) {
	tests := []struct {
		name     string
		file     *librarymodel.File
		expected string
	}{
		{
			name:     "Small file (bytes)",
			file:     &librarymodel.File{FileSize: 512},
			expected: "512 B",
		},
		{
			name:     "Medium file (KB)",
			file:     &librarymodel.File{FileSize: 1024},
			expected: "1.0 KB",
		},
		{
			name:     "Large file (MB)",
			file:     &librarymodel.File{FileSize: 1048576},
			expected: "1.0 MB",
		},
		{
			name:     "Very large file (GB)",
			file:     &librarymodel.File{FileSize: 1073741824},
			expected: "1.0 GB",
		},
		{
			name:     "Zero size",
			file:     &librarymodel.File{FileSize: 0},
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.GetFileSizeFormatted()
			if result != tt.expected {
				t.Errorf("GetFileSizeFormatted() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFile_IsPDF(t *testing.T) {
	tests := []struct {
		name     string
		file     *librarymodel.File
		expected bool
	}{
		{
			name:     "PDF file",
			file:     &librarymodel.File{FileType: "pdf"},
			expected: true,
		},
		{
			name:     "Document file",
			file:     &librarymodel.File{FileType: "document"},
			expected: false,
		},
		{
			name:     "Image file",
			file:     &librarymodel.File{FileType: "image"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.IsPDF()
			if result != tt.expected {
				t.Errorf("IsPDF() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFile_IsImage(t *testing.T) {
	tests := []struct {
		name     string
		file     *librarymodel.File
		expected bool
	}{
		{
			name:     "Image file",
			file:     &librarymodel.File{FileType: "image"},
			expected: true,
		},
		{
			name:     "PDF file",
			file:     &librarymodel.File{FileType: "pdf"},
			expected: false,
		},
		{
			name:     "Document file",
			file:     &librarymodel.File{FileType: "document"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.IsImage()
			if result != tt.expected {
				t.Errorf("IsImage() = %v, want %v", result, tt.expected)
			}
		})
	}
}
