package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/library/models"
	"github.com/stretchr/testify/mock"
)

type MockUUID struct {
	GenerateUUIDCall bool
}

func (m *MockUUID) GenerateUUID() string {
	m.GenerateUUIDCall = true
	return "any_id"
}

type MockEbookRepository struct {
	mock.Mock
}

func (m *MockEbookRepository) Save(ebook *models.Ebook) error {
	args := m.Called(ebook)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockEbookRepository) FindByParams(params ...interface{}) ([]*models.Ebook, error) {
	args := m.Called(params...)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Ebook), args.Error(1)
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(input interface{}) error {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

var ebookRepoMock *MockEbookRepository
var validatorMock *MockValidator
var sut *EbookService
var uuidMock *MockUUID

func setUpMocks() {
	ebookRepoMock = new(MockEbookRepository)
	validatorMock = new(MockValidator)
	uuidMock = new(MockUUID)
	sut = NewEbookService(uuidMock, ebookRepoMock, validatorMock)
}

func TestEbookService_Create(t *testing.T) {
	t.Run("Create ebook with success", func(t *testing.T) {
		setUpMocks()
		expectedID := "any_id"
		input := models.Ebook{
			ID:               expectedID,
			Title:            "any_title",
			Price:            100,
			PromotionalPrice: 10,
			InfoProducerID:   "any_info_producer_id",
		}
		ebookRepoMock.On("FindByParams", "any_title", "any_info_producer_id").Return(nil, nil)
		ebookRepoMock.On("Save", &input).Return(nil)

		validatorMock.On("Validate", input).Return(nil)

		newEbook, err := sut.CreateEbook(input)
		if err != nil {
			t.Errorf("Error when creating ebook: %v", err)
		}

		if uuidMock.GenerateUUIDCall == false {
			t.Errorf("GenerateUUID was not called")
		}

		if newEbook.ID != expectedID {
			t.Errorf("Ebook created is nil")
		}
	})

	t.Run("Should return error if ebook already exists", func(t *testing.T) {
		setUpMocks()
		ebooks := []*models.Ebook{}
		ebooks = append(ebooks, &models.Ebook{})
		input := &models.Ebook{
			ID:             "any_id",
			Title:          "any_title",
			InfoProducerID: "any_info_producer_id",
		}

		ebookRepoMock.On("FindByParams", "any_title", "any_info_producer_id").Return(ebooks, nil)
		ebookRepoMock.On("Save", input).Return(nil)
		validatorMock.On("Validate", input).Return(nil)

		newEbook, err := sut.CreateEbook(models.Ebook{
			Title:          "any_title",
			InfoProducerID: "any_info_producer_id",
		})

		if newEbook != nil {
			t.Errorf("Ebook created is not nil")
		}

		if err == nil || !errors.Is(err, ErrEbookAlreadyExist) {
			t.Errorf("Error was expected")
		}
	})

	t.Run("Should return error if ebook is not valid", func(t *testing.T) {
		input := &models.Ebook{
			ID:             "any_id",
			Title:          "any_title",
			InfoProducerID: "any_info_producer_id",
		}

		ebookRepoMock.On("FindByParams", "any_title", "any_info_producer_id").Return(nil, nil)
		ebookRepoMock.On("Save", input).Return(nil)
		validatorMock.On("Validate", input).Return(nil)

		_, err := sut.CreateEbook(models.Ebook{
			Title:            "any_title",
			InfoProducerID:   "any_info_producer_id",
			SalesDescription: strings.Repeat("a", 121), // Invalid sales description
		})
		if err == nil {
			t.Errorf("Error was expected")
		}
	})
}
