package repository_test

import (
	"testing"
	"time"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountrepo "github.com/Assinee-Inc/docffy/internal/account/repository"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	"github.com/Assinee-Inc/docffy/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CreatorGormTestSuite struct {
	suite.Suite
	db          *gorm.DB
	creatorRepo *accountrepo.GormCreatorRepository
}

func TestCreatorGormTestSuite(t *testing.T) {
	suite.Run(t, new(CreatorGormTestSuite))
}

func (suite *CreatorGormTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), db.AutoMigrate(&authmodel.User{}, &accountmodel.Creator{}))
	suite.db = db
	database.DB = db
	suite.creatorRepo = accountrepo.NewGormCreatorRepository(suite.db)
}

func (suite *CreatorGormTestSuite) SetupTest() {
	suite.db.Exec("DELETE FROM creators")
	suite.db.Exec("DELETE FROM users")
}

func (suite *CreatorGormTestSuite) TearDownSuite() {
	suite.db.Exec("DELETE FROM creators")
	suite.db.Exec("DELETE FROM users")
}

func (suite *CreatorGormTestSuite) TestFindCreatorByUserEmail_Success_Aligned() {
	// 1. Criar Usuário
	user := authmodel.NewUser("testuser", "password123", "aligned@example.com")
	err := suite.db.Create(user).Error
	suite.Require().NoError(err)

	// 2. Criar Creator com user_id correto
	birthDate, _ := time.Parse("2006-01-02", "1990-01-01")
	creator := accountmodel.NewCreator(
		"Test Creator",
		"Test Social",
		"aligned@example.com",
		"11999999999",
		"12345678901",
		birthDate,
		user.ID,
	)
	err = suite.creatorRepo.Create(creator)
	suite.Require().NoError(err)

	// 3. Buscar Creator por email
	foundCreator, err := suite.creatorRepo.FindCreatorByUserEmail("aligned@example.com")
	suite.Require().NoError(err)
	suite.Require().NotNil(foundCreator)
	assert.Equal(suite.T(), creator.ID, foundCreator.ID)
	assert.Equal(suite.T(), user.ID, foundCreator.UserID)
}

func (suite *CreatorGormTestSuite) TestFindCreatorByUserEmail_Success_AutoHealMismatched() {
	// 1. Criar Usuário
	user := authmodel.NewUser("testuser", "password123", "mismatched@example.com")
	err := suite.db.Create(user).Error
	suite.Require().NoError(err)

	// 2. Criar Creator com user_id desalinhado (por exemplo, 9999)
	birthDate, _ := time.Parse("2006-01-02", "1990-01-01")
	creator := accountmodel.NewCreator(
		"Test Creator",
		"Test Social",
		"mismatched@example.com",
		"11999999999",
		"12345678901",
		birthDate,
		9999, // UserID incorreto
	)
	err = suite.creatorRepo.Create(creator)
	suite.Require().NoError(err)

	// 3. Buscar Creator por email do usuário (deve disparar auto-healing)
	foundCreator, err := suite.creatorRepo.FindCreatorByUserEmail("mismatched@example.com")
	suite.Require().NoError(err)
	suite.Require().NotNil(foundCreator)
	assert.Equal(suite.T(), creator.ID, foundCreator.ID)
	assert.Equal(suite.T(), user.ID, foundCreator.UserID) // Deve ter sido atualizado para user.ID

	// 4. Confirmar que o registro no banco de dados foi realmente atualizado
	var updatedCreator accountmodel.Creator
	err = suite.db.First(&updatedCreator, creator.ID).Error
	suite.Require().NoError(err)
	assert.Equal(suite.T(), user.ID, updatedCreator.UserID)
}

func (suite *CreatorGormTestSuite) TestFindCreatorByUserEmail_NotFound() {
	// Buscar por um email que não possui Creator
	foundCreator, err := suite.creatorRepo.FindCreatorByUserEmail("nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundCreator)
	assert.Equal(suite.T(), "creator not found", err.Error())
}
