package service_test

import (
	"strings"
	"testing"
	"time"

	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenService_GenerateAndValidateToken(t *testing.T) {
	secret := "test-jwt-hmac-secret-key-256-bits"
	svc := authsvc.NewTokenService(secret, 24*time.Hour)

	user := &authmodel.User{
		PublicID: "usr_abc123",
		Username: "Maria Silva",
		Email:    "maria@example.com",
		GoogleID: "google-sub-123",
	}
	user.ID = 42

	t.Run("emite e valida JWT HS256 com claims corretas", func(t *testing.T) {
		tokenStr, expiresIn, err := svc.GenerateToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenStr)
		assert.Equal(t, int64(86400), expiresIn)

		parts := strings.Split(tokenStr, ".")
		assert.Len(t, parts, 3)

		claims, err := svc.ValidateToken(tokenStr)
		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "usr_abc123", claims.Sub)
		assert.Equal(t, uint(42), claims.UID)
		assert.Equal(t, "maria@example.com", claims.Email)
		assert.Greater(t, claims.Exp, claims.Iat)
	})

	t.Run("rejeita token com assinatura adulterada ou segredo diferente", func(t *testing.T) {
		tokenStr, _, err := svc.GenerateToken(user)
		require.NoError(t, err)

		otherSvc := authsvc.NewTokenService("wrong-secret-key", 24*time.Hour)
		claims, err := otherSvc.ValidateToken(tokenStr)
		assert.Error(t, err)
		assert.Nil(t, claims)

		tampered := tokenStr + "tampered"
		claims, err = svc.ValidateToken(tampered)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("rejeita token expirado", func(t *testing.T) {
		expiredSvc := authsvc.NewTokenService(secret, -1*time.Hour)
		tokenStr, _, err := expiredSvc.GenerateToken(user)
		require.NoError(t, err)

		claims, err := svc.ValidateToken(tokenStr)
		assert.ErrorIs(t, err, authsvc.ErrExpiredToken)
		assert.Nil(t, claims)
	})

	t.Run("rejeita token malformado ou vazio", func(t *testing.T) {
		invalidTokens := []string{
			"",
			"not-a-jwt",
			"a.b",
			"a.b.c.d",
			"invalid_base64.invalid_base64.invalid_base64",
		}
		for _, tok := range invalidTokens {
			claims, err := svc.ValidateToken(tok)
			assert.Error(t, err)
			assert.Nil(t, claims)
		}
	})
}
