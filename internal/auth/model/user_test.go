package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewGoogleUser(t *testing.T) {
	user := NewGoogleUser("Maria Silva", "maria@example.com", "google-sub-123")

	assert.Equal(t, "Maria Silva", user.Username)
	assert.Equal(t, "maria@example.com", user.Email)
	assert.Equal(t, "google-sub-123", user.GoogleID)
	assert.NotNil(t, user.EmailVerifiedAt, "EmailVerifiedAt deve ser preenchido na criação via Google")
	assert.True(t, user.IsEmailVerified())
	assert.True(t, user.IsConfirmed())
	assert.Nil(t, user.TermsAcceptedAt, "TermsAcceptedAt só deve ser preenchido na Etapa 2 de onboarding")
	assert.False(t, user.HasAcceptedTerms())
}

func TestUser_NoPasswordOrLegacyTokenFields(t *testing.T) {
	userType := reflect.TypeOf(User{})

	removedFields := []string{
		"Password",
		"PasswordResetToken",
		"PasswordResetAt",
		"EmailVerificationToken",
		"CSRFToken",
		"SessionToken",
	}

	for _, fieldName := range removedFields {
		_, found := userType.FieldByName(fieldName)
		assert.Falsef(t, found, "Campo legado %q não deve existir na entidade User", fieldName)
	}

	googleField, found := userType.FieldByName("GoogleID")
	assert.True(t, found, "Campo GoogleID deve existir na entidade User")
	assert.Contains(t, googleField.Tag.Get("json"), "google_id")
	assert.Contains(t, googleField.Tag.Get("gorm"), "uniqueIndex")
}

func TestUser_HasAcceptedTerms(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "should return true when TermsAcceptedAt is set",
			user: &User{
				TermsAcceptedAt: &now,
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
			assert.Equal(t, tt.expected, tt.user.HasAcceptedTerms())
		})
	}
}

func TestUser_IsConfirmedAndVerified(t *testing.T) {
	now := time.Now()
	verified := &User{EmailVerifiedAt: &now}
	unverified := &User{EmailVerifiedAt: nil}

	assert.True(t, verified.IsEmailVerified())
	assert.True(t, verified.IsConfirmed())
	assert.False(t, unverified.IsEmailVerified())
	assert.False(t, unverified.IsConfirmed())
}
