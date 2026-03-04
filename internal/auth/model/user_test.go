package model

import (
	"testing"
	"time"
)

func TestNewUser_TermsAcceptedAt(t *testing.T) {
	user := NewUser("testuser", "password123", "test@example.com")

	if user.TermsAcceptedAt == nil {
		t.Error("TermsAcceptedAt should be set when creating a new user")
	}

	if !user.HasAcceptedTerms() {
		t.Error("HasAcceptedTerms() should return true when TermsAcceptedAt is set")
	}
}

func TestUser_HasAcceptedTerms(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "should return true when TermsAcceptedAt is set",
			user: &User{
				TermsAcceptedAt: &time.Time{},
			},
			expected: true,
		},
		{
			name: "should return false when TermsAcceptedAt is nil",
			user: &User{
				TermsAcceptedAt: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.HasAcceptedTerms()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
