//go:build integration
// +build integration

package repository

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/anglesson/simple-web-server/internal/account"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: conn,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to initialize GORM: %v", err)
	}

	err = gormDB.AutoMigrate(&account.Account{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return gormDB
}

func TestGormAccountRepository_Create_Nil_Account(t *testing.T) {
	gormDB := setupTestDB(t)
	repo := NewGormAccountRepository(gormDB)
	assert.Error(t, repo.Create(nil), "Expected error when creating nil account")
}

func TestGormAccountRepository_Create_Valid_Account(t *testing.T) {
	gormDB := setupTestDB(t)

	repo := NewGormAccountRepository(gormDB)

	testAccount := &account.Account{
		Name:                 "Test User",
		CPF:                  "12345678900",
		Email:                "test@example.com",
		Phone:                "11999999999",
		BirthDate:            time.Now(),
		UserID:               uuid.New(),
		Origin:               "any_origin",
		ExternalAccountID:    "any_external_account_id",
		OnboardingCompleted:  true,
		OnboardingRefreshURL: "http://example.com/refresh",
		OnboardingReturnURL:  "http://example.com/return",
		PayoutsEnabled:       true,
		ChargesEnabled:       true,
	}
	err := repo.Create(testAccount)
	log.Printf("Create account error: %v", err)
	assert.NoError(t, err, "Expected no error when creating valid account")

	var retrieved account.Account
	err = gormDB.First(&retrieved, "email = ?", testAccount.Email).Error

	assert.NoError(t, err, "Expected to find created account in database")
	assert.IsType(t, uuid.UUID{}, retrieved.ID, "Expected account ID field to have type uuid.UUID, got %T", retrieved.ID)
	assert.Equal(t, testAccount.Name, retrieved.Name, "Expected names to match")
	assert.Equal(t, testAccount.CPF, retrieved.CPF, "Expected CPFs to match")
	assert.Equal(t, testAccount.Email, retrieved.Email, "Expected emails to match")
	assert.Equal(t, testAccount.Phone, retrieved.Phone, "Expected phones to match")
	assert.Equal(t, testAccount.BirthDate.Unix(), retrieved.BirthDate.Unix(), "Expected birth dates to match")
	assert.Equal(t, testAccount.UserID, retrieved.UserID, "Expected user IDs to match")
	assert.Equal(t, testAccount.Origin, retrieved.Origin, "Expected origins to match")
	assert.Equal(t, testAccount.ExternalAccountID, retrieved.ExternalAccountID, "Expected external account IDs to match")
	assert.Equal(t, testAccount.OnboardingCompleted, retrieved.OnboardingCompleted, "Expected onboarding completed flags to match")
	assert.Equal(t, testAccount.OnboardingRefreshURL, retrieved.OnboardingRefreshURL, "Expected onboarding refresh URLs to match")
	assert.Equal(t, testAccount.OnboardingReturnURL, retrieved.OnboardingReturnURL, "Expected onboarding return URLs to match")
	assert.Equal(t, testAccount.PayoutsEnabled, retrieved.PayoutsEnabled, "Expected payouts enabled flags to match")
	assert.Equal(t, testAccount.ChargesEnabled, retrieved.ChargesEnabled, "Expected charges enabled flags to match")
}
