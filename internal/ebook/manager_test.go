package ebook_test

import (
	"context"
	"errors"
	"testing"

	"github.com/anglesson/simple-web-server/internal/ebook"
	uuid_mocks "github.com/anglesson/simple-web-server/internal/platform/uuid/mocks"
	"github.com/stretchr/testify/mock"
)

var ebookRepoMock *MockEbookRepository
var manager *ebook.Manager
var uuidMock *uuid_mocks.MockUUID

func setUpMocks() {
	ebookRepoMock = new(MockEbookRepository)
	uuidMock = new(uuid_mocks.MockUUID)
	manager = ebook.NewManager(uuidMock, ebookRepoMock)
}

func TestEbookManager_Create(t *testing.T) {
	t.Run("Create ebook with success", func(t *testing.T) {
		setUpMocks()
		expectedID := "any_id"
		input := &ebook.Ebook{
			Title:            "any_title",
			Price:            100,
			PromotionalPrice: 10,
			InfoProducerID:   "any_info_producer_id",
		}

		uuidMock.On("Generate").Return(expectedID)
		ebookRepoMock.On("FindByParams", mock.Anything, "any_title", "any_info_producer_id").Return(nil, nil)
		ebookRepoMock.On("Save", mock.Anything, mock.AnythingOfType("*ebook.Ebook")).Return(nil)

		newEbook, err := manager.CreateEbook(context.Background(), input)
		if err != nil {
			t.Errorf("Error when creating ebook: %v", err)
		}

		if newEbook == nil {
			t.Fatal("Ebook created is nil")
		}

		if newEbook.ID != expectedID {
			t.Errorf("Expected ID %s, got %s", expectedID, newEbook.ID)
		}
	})

	t.Run("Should return an error if fails", func(t *testing.T) {
		setUpMocks()

		input := &ebook.Ebook{
			Title:            "any_title",
			Price:            100,
			PromotionalPrice: 10,
			InfoProducerID:   "any_info_producer_id",
		}

		ebookRepoMock.On("Save", mock.Anything, mock.AnythingOfType("*ebook.Ebook")).Return(errors.New("Error on repository"))
		uuidMock.On("Generate").Return("any_uuid")

		_, err := manager.CreateEbook(context.Background(), input)
		if err == nil {
			t.Error("Expected error, got 'nil'")
		}
	})
}
