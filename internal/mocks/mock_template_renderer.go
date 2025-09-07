package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockTemplateRenderer implements template.TemplateRenderer for testing
type MockTemplateRenderer struct {
	mock.Mock
}

func (m *MockTemplateRenderer) View(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}, layout string) {
	m.Called(w, r, page, data, layout)
}

func (m *MockTemplateRenderer) ViewWithoutLayout(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	m.Called(w, r, page, data)
}
