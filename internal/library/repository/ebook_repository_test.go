package repository_test

import (
	"testing"
	"time"

	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"
	"github.com/anglesson/simple-web-server/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type EbookRepositoryTestSuite struct {
	suite.Suite
	db              *gorm.DB
	ebookRepository libraryrepo.EbookRepository
	creator         *accountmodel.Creator
}

func (suite *EbookRepositoryTestSuite) SetupSuite() {
	// Configurar banco de teste
	database.Connect()
	suite.db = database.DB

	// Auto-migrate
	suite.db.AutoMigrate(&librarymodel.Ebook{}, &accountmodel.Creator{}, &authmodel.User{}, &librarymodel.File{})
}

func (suite *EbookRepositoryTestSuite) SetupTest() {
	// Limpar dados antes de cada teste
	suite.db.Exec("DELETE FROM ebook_files")
	suite.db.Exec("DELETE FROM ebooks")
	suite.db.Exec("DELETE FROM files")
	suite.db.Exec("DELETE FROM creators")
	suite.db.Exec("DELETE FROM users")

	// Criar criador de teste
	user := &authmodel.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	suite.db.Create(user)

	birthDate, _ := time.Parse("2006-01-02", "1990-01-01")
	suite.creator = &accountmodel.Creator{
		Name:      "Test Creator",
		Email:     "creator@example.com",
		CPF:       "12345678901",
		Phone:     "11999999999",
		BirthDate: birthDate,
		UserID:    user.ID,
	}
	suite.db.Create(suite.creator)

	// Inicializar repositório
	suite.ebookRepository = libraryrepo.NewGormEbookRepository(suite.db)
}

func (suite *EbookRepositoryTestSuite) TearDownSuite() {
	// Limpar após todos os testes
	suite.db.Exec("DELETE FROM ebook_files")
	suite.db.Exec("DELETE FROM ebooks")
	suite.db.Exec("DELETE FROM files")
	suite.db.Exec("DELETE FROM creators")
	suite.db.Exec("DELETE FROM users")
}

func (suite *EbookRepositoryTestSuite) TestCreate() {
	// Arrange
	ebook := &librarymodel.Ebook{
		Title:       "Test Ebook",
		Description: "Test description",
		SalesPage:   "Sales page content",
		Value:       29.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}

	// Act
	err := suite.ebookRepository.Create(ebook)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), ebook.ID)

	// Verificar se foi salvo no banco
	var savedEbook librarymodel.Ebook
	suite.db.First(&savedEbook, ebook.ID)
	assert.Equal(suite.T(), ebook.Title, savedEbook.Title)
}

func (suite *EbookRepositoryTestSuite) TestFindByID() {
	// Arrange
	ebook := &librarymodel.Ebook{
		Title:       "Test Ebook",
		Description: "Test description",
		SalesPage:   "Sales page content",
		Value:       29.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	suite.db.Create(ebook)

	// Act
	foundEbook, err := suite.ebookRepository.FindByID(ebook.ID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundEbook)
	assert.Equal(suite.T(), ebook.Title, foundEbook.Title)
	assert.Equal(suite.T(), ebook.CreatorID, foundEbook.CreatorID)
}

func (suite *EbookRepositoryTestSuite) TestFindByCreator() {
	// Arrange
	ebook1 := &librarymodel.Ebook{
		Title:       "First Ebook",
		Description: "First description",
		SalesPage:   "First sales page",
		Value:       29.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	ebook2 := &librarymodel.Ebook{
		Title:       "Second Ebook",
		Description: "Second description",
		SalesPage:   "Second sales page",
		Value:       39.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	suite.db.Create(ebook1)
	suite.db.Create(ebook2)

	// Act
	ebooks, err := suite.ebookRepository.FindByCreator(suite.creator.ID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), ebooks, 2)
	assert.Equal(suite.T(), ebook2.Title, ebooks[0].Title) // Ordenado por created_at DESC
	assert.Equal(suite.T(), ebook1.Title, ebooks[1].Title)
}

func (suite *EbookRepositoryTestSuite) TestUpdate() {
	// Arrange
	ebook := &librarymodel.Ebook{
		Title:       "Original Title",
		Description: "Original description",
		SalesPage:   "Original sales page",
		Value:       29.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	suite.db.Create(ebook)

	// Act
	ebook.Title = "Updated Title"
	ebook.Description = "Updated description"
	err := suite.ebookRepository.Update(ebook)

	// Assert
	assert.NoError(suite.T(), err)

	// Verificar se foi atualizado no banco
	var updatedEbook librarymodel.Ebook
	suite.db.First(&updatedEbook, ebook.ID)
	assert.Equal(suite.T(), "Updated Title", updatedEbook.Title)
	assert.Equal(suite.T(), "Updated description", updatedEbook.Description)
}

func (suite *EbookRepositoryTestSuite) TestDelete() {
	// Arrange
	ebook := &librarymodel.Ebook{
		Title:       "Test Ebook",
		Description: "Test description",
		SalesPage:   "Sales page content",
		Value:       29.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	suite.db.Create(ebook)

	// Act
	err := suite.ebookRepository.Delete(ebook.ID)

	// Assert
	assert.NoError(suite.T(), err)

	// Verificar se foi deletado do banco
	var deletedEbook librarymodel.Ebook
	result := suite.db.First(&deletedEbook, ebook.ID)
	assert.Error(suite.T(), result.Error) // Deve retornar erro pois não existe mais
}

func (suite *EbookRepositoryTestSuite) TestFindAll() {
	// Arrange
	ebook1 := &librarymodel.Ebook{
		Title:       "First Ebook",
		Description: "First description",
		SalesPage:   "First sales page",
		Value:       29.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	ebook2 := &librarymodel.Ebook{
		Title:       "Second Ebook",
		Description: "Second description",
		SalesPage:   "Second sales page",
		Value:       39.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	suite.db.Create(ebook1)
	suite.db.Create(ebook2)

	// Act
	ebooks, err := suite.ebookRepository.FindAll()

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), ebooks, 2)
}

func (suite *EbookRepositoryTestSuite) TestFindActive() {
	// Arrange
	activeEbook := &librarymodel.Ebook{
		Title:       "Active Ebook",
		Description: "Active description",
		SalesPage:   "Active sales page",
		Value:       29.90,
		Status:      true,
		CreatorID:   suite.creator.ID,
	}
	inactiveEbook := &librarymodel.Ebook{
		Title:       "Inactive Ebook",
		Description: "Inactive description",
		SalesPage:   "Inactive sales page",
		Value:       39.90,
		Status:      false,
		CreatorID:   suite.creator.ID,
	}
	suite.db.Create(activeEbook)
	suite.db.Create(inactiveEbook)

	// Act
	activeEbooks, err := suite.ebookRepository.FindActive()

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), activeEbooks, 1)
	assert.Equal(suite.T(), activeEbook.Title, activeEbooks[0].Title)
	assert.True(suite.T(), activeEbooks[0].Status)
}

func TestEbookRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(EbookRepositoryTestSuite))
}
