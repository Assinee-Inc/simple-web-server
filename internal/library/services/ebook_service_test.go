package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/library/mocks"
	"github.com/anglesson/simple-web-server/internal/library/models"
)

var ebookRepoMock *mocks.MockEbookRepository
var validatorMock *mocks.MockValidator
var sut *EbookService
var uuidMock *mocks.MockUUID

func setUpMocks() {
	ebookRepoMock = new(mocks.MockEbookRepository)
	validatorMock = new(mocks.MockValidator)
	uuidMock = new(mocks.MockUUID)
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
		uuidMock.On("GenerateUUID").Return(expectedID, nil)
		ebookRepoMock.On("FindByParams", "any_title", "any_info_producer_id").Return(nil, nil)
		ebookRepoMock.On("Save", &input).Return(nil)

		validatorMock.On("Validate", input).Return(nil)

		newEbook, err := sut.CreateEbook(input)
		if err != nil {
			t.Errorf("Error when creating ebook: %v", err)
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

		uuidMock.On("GenerateUUID").Return("any_id", nil)

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
