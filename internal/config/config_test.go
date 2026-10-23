package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigs_GoogleAndJWT(t *testing.T) {
	t.Run("carrega GOOGLE_CLIENT_ID, JWT_SECRET e JWT_EXPIRATION_HOURS do ambiente", func(t *testing.T) {
		t.Setenv("APPLICATION_MODE", "test")
		t.Setenv("GOOGLE_CLIENT_ID", "test-google-client.apps.googleusercontent.com")
		t.Setenv("JWT_SECRET", "super-secret-jwt-key")
		t.Setenv("JWT_EXPIRATION_HOURS", "48")

		LoadConfigs()

		assert.Equal(t, "test-google-client.apps.googleusercontent.com", AppConfig.GoogleClientID)
		assert.Equal(t, "super-secret-jwt-key", AppConfig.JWTSecret)
		assert.Equal(t, 48, AppConfig.JWTExpirationHours)
	})

	t.Run("usa fallback de JWT_SECRET para SESSION_AUTH_KEY ou APP_KEY e default 24h", func(t *testing.T) {
		t.Setenv("APPLICATION_MODE", "test")
		_ = os.Unsetenv("JWT_SECRET")
		_ = os.Unsetenv("JWT_EXPIRATION_HOURS")
		t.Setenv("SESSION_AUTH_KEY", "session-fallback-secret")
		t.Setenv("APP_KEY", "app-key-fallback")

		LoadConfigs()

		assert.Equal(t, "session-fallback-secret", AppConfig.JWTSecret)
		assert.Equal(t, 24, AppConfig.JWTExpirationHours)

		_ = os.Unsetenv("SESSION_AUTH_KEY")
		LoadConfigs()
		assert.Equal(t, "app-key-fallback", AppConfig.JWTSecret)
	})
}

func TestLoadConfigs_FrontendURL(t *testing.T) {
	t.Run("carrega FRONTEND_URL do ambiente quando definido", func(t *testing.T) {
		t.Setenv("APPLICATION_MODE", "test")
		t.Setenv("FRONTEND_URL", "https://app.docffy.com")

		LoadConfigs()

		assert.Equal(t, "https://app.docffy.com", AppConfig.FrontendURL)
	})

	t.Run("usa default http://localhost:5173 quando FRONTEND_URL não definido", func(t *testing.T) {
		t.Setenv("APPLICATION_MODE", "test")
		_ = os.Unsetenv("FRONTEND_URL")

		LoadConfigs()

		assert.Equal(t, "http://localhost:5173", AppConfig.FrontendURL)
	})
}
