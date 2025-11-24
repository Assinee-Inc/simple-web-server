package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/library/domain"
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

func (m *MockEbookRepository) Save(ebook *domain.Ebook) error {
	args := m.Called(ebook)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func (m *MockEbookRepository) FindByParams(params ...interface{}) ([]*domain.Ebook, error) {
	args := m.Called(params...)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Ebook), args.Error(1)
}

func TestEbookService_Create(t *testing.T) {
	t.Run("Create ebook with success", func(t *testing.T) {
		expectedID := "any_id"
		input := domain.Ebook{}
		mockUUID := &MockUUID{}

		mockEbookRepo := new(MockEbookRepository)
		mockEbookRepo.On("FindByParams", "", "").Return(nil, nil)
		mockEbookRepo.On("Save", &domain.Ebook{
			ID: expectedID,
		}).Return(nil)

		ebooksService := NewEbookService(mockUUID, mockEbookRepo)

		newEbook, err := ebooksService.CreateEbook(input)
		if err != nil {
			t.Errorf("Error when creating ebook: %v", err)
		}

		if mockUUID.GenerateUUIDCall == false {
			t.Errorf("GenerateUUID was not called")
		}

		if newEbook.ID != expectedID {
			t.Errorf("Ebook created is nil")
		}
	})

	t.Run("Should return error if ebook already exists", func(t *testing.T) {
		mockUUID := &MockUUID{}

		mockEbookRepo := new(MockEbookRepository)
		ebooks := []*domain.Ebook{}
		ebooks = append(ebooks, &domain.Ebook{})
		mockEbookRepo.On("FindByParams", "any_title", "any_info_producer_id").Return(ebooks, nil)
		mockEbookRepo.On("Save", &domain.Ebook{
			ID:             "any_id",
			Title:          "any_title",
			InfoProducerID: "any_info_producer_id",
		}).Return(nil)

		ebooksService := NewEbookService(mockUUID, mockEbookRepo)
		newEbook, err := ebooksService.CreateEbook(domain.Ebook{
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
		mockUUID := &MockUUID{}
		mockEbookRepo := new(MockEbookRepository)
		mockEbookRepo.On("FindByParams", "", "").Return(nil, nil)
		mockEbookRepo.On("Save", &domain.Ebook{
			ID: "any_id",
		}).Return(nil)

		ebooksService := NewEbookService(mockUUID, mockEbookRepo)

		_, err := ebooksService.CreateEbook(domain.Ebook{
			SalesDescription: strings.Repeat("a", 121), // Invalid title,
		})
		if err == nil {
			t.Errorf("Error was expected")
		}
	})
}
