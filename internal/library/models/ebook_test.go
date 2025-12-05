package models_test

import (
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/library/models"
)

func TestEbook_Validate(t *testing.T) {
	t.Run("info_producer_id is required", func(t *testing.T) {
		ebook := &models.Ebook{
			Title: "any title",
		}

		err := ebook.Validate()

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Promotional value is greater than price", func(t *testing.T) {
		ebook := &models.Ebook{
			InfoProducerID:   "any-id",
			Title:            "any title",
			Price:            100,
			PromotionalPrice: 200,
		}

		err := ebook.Validate()

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err != nil && err.Error() != "O campo promotional_price deve ser menor que o campo Price" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Title has more than 50 characters", func(t *testing.T) {
		ebook := &models.Ebook{
			InfoProducerID: "any-id",
			Title:          strings.Repeat("a", 51),
		}

		err := ebook.Validate()

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Description has more than 120 characters", func(t *testing.T) {
		ebook := &models.Ebook{
			InfoProducerID: "any-id",
			Title:          "any title",
			Description:    strings.Repeat("a", 121),
		}

		err := ebook.Validate()

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Sales description has more than 120 characters", func(t *testing.T) {
		ebook := &models.Ebook{
			InfoProducerID:   "any-id",
			Title:            "any title",
			SalesDescription: strings.Repeat("a", 121),
		}
		err := ebook.Validate()
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
