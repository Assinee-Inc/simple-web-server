package model_test

import (
	"testing"

	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
)

func TestClient_GetInitials(t *testing.T) {
	tests := []struct {
		name     string
		client   *salesmodel.Client
		expected string
	}{
		{
			name:     "Single name",
			client:   &salesmodel.Client{Name: "João"},
			expected: "J",
		},
		{
			name:     "Two names",
			client:   &salesmodel.Client{Name: "João Silva"},
			expected: "JS",
		},
		{
			name:     "Three names",
			client:   &salesmodel.Client{Name: "João Pedro Silva"},
			expected: "JS",
		},
		{
			name:     "Empty name",
			client:   &salesmodel.Client{Name: ""},
			expected: "?",
		},
		{
			name:     "Multiple spaces",
			client:   &salesmodel.Client{Name: "João   Silva"},
			expected: "JS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.GetInitials()
			if result != tt.expected {
				t.Errorf("GetInitials() = %v, want %v", result, tt.expected)
			}
		})
	}
}
