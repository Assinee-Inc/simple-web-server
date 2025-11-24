package domain_test

import (
	"strings"
	"testing"

	"github.com/anglesson/simple-web-server/internal/library/domain"
)

func TestEbook_Validate(t *testing.T) {
	t.Run("Promotional value is greater than price", func(t *testing.T) {
		ebook := &domain.Ebook{
			Price:            100,
			PromotionalPrice: 200,
		}

		err := ebook.Validate()

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err != nil && err.Error() != "promotional Price cannot be greater than value" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Title has more than 50 characters", func(t *testing.T) {
		ebook := &domain.Ebook{
			Title: strings.Repeat("a", 51),
		}

		err := ebook.Validate()

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err != nil && err.Error() != "title cannot be longer than 50 characters" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Description has more than 120 characters", func(t *testing.T) {
		ebook := &domain.Ebook{
			Description: strings.Repeat("a", 121),
		}

		err := ebook.Validate()

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err != nil && err.Error() != "description cannot be longer than 120 characters" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Sales description has more than 120 characters", func(t *testing.T) {
		ebook := &domain.Ebook{
			SalesDescription: strings.Repeat("a", 121),
		}
		err := ebook.Validate()
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err != nil && err.Error() != "sales Description cannot be longer than 120 characters" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

}
