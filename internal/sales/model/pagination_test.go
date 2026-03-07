package model_test

import (
	"testing"

	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
)

func TestPagination_NewPagination(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expected *salesmodel.Pagination
	}{
		{
			name:  "Valid pagination",
			page:  1,
			limit: 10,
			expected: &salesmodel.Pagination{
				Page:       1,
				Limit:      10,
				PageSize:   10,
				Start:      1,
				End:        10,
				PrevPage:   1,
				NextPage:   2,
				Total:      0,
				TotalPages: 0,
			},
		},
		{
			name:  "Zero page should default to 1",
			page:  0,
			limit: 10,
			expected: &salesmodel.Pagination{
				Page:       1,
				Limit:      10,
				PageSize:   10,
				Start:      1,
				End:        10,
				PrevPage:   1,
				NextPage:   2,
				Total:      0,
				TotalPages: 0,
			},
		},
		{
			name:  "Zero limit should default to 10",
			page:  1,
			limit: 0,
			expected: &salesmodel.Pagination{
				Page:       1,
				Limit:      10,
				PageSize:   10,
				Start:      1,
				End:        10,
				PrevPage:   1,
				NextPage:   2,
				Total:      0,
				TotalPages: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := salesmodel.NewPagination(tt.page, tt.limit)

			if result.Page != tt.expected.Page {
				t.Errorf("Page = %v, want %v", result.Page, tt.expected.Page)
			}
			if result.Limit != tt.expected.Limit {
				t.Errorf("Limit = %v, want %v", result.Limit, tt.expected.Limit)
			}
			if result.PageSize != tt.expected.PageSize {
				t.Errorf("PageSize = %v, want %v", result.PageSize, tt.expected.PageSize)
			}
			if result.Start != tt.expected.Start {
				t.Errorf("Start = %v, want %v", result.Start, tt.expected.Start)
			}
			if result.End != tt.expected.End {
				t.Errorf("End = %v, want %v", result.End, tt.expected.End)
			}
			if result.PrevPage != tt.expected.PrevPage {
				t.Errorf("PrevPage = %v, want %v", result.PrevPage, tt.expected.PrevPage)
			}
			if result.NextPage != tt.expected.NextPage {
				t.Errorf("NextPage = %v, want %v", result.NextPage, tt.expected.NextPage)
			}
		})
	}
}

func TestPagination_SetTotal(t *testing.T) {
	pagination := salesmodel.NewPagination(1, 10)

	tests := []struct {
		name     string
		total    int64
		expected int
	}{
		{
			name:     "Small total",
			total:    25,
			expected: 3,
		},
		{
			name:     "Exact total",
			total:    10,
			expected: 1,
		},
		{
			name:     "Zero total",
			total:    0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pagination.SetTotal(tt.total)

			if pagination.Total != tt.total {
				t.Errorf("Total = %v, want %v", pagination.Total, tt.total)
			}
			if pagination.TotalPages != tt.expected {
				t.Errorf("TotalPages = %v, want %v", pagination.TotalPages, tt.expected)
			}
		})
	}
}
