package mocks

import "github.com/stretchr/testify/mock"

// MockMailer é um mock do Mailer para este teste
type MockMailerSimple struct {
	mock.Mock
}

func (m *MockMailerSimple) From(email string) {
	m.Called(email)
}

func (m *MockMailerSimple) To(email string) {
	m.Called(email)
}

func (m *MockMailerSimple) Subject(subject string) {
	m.Called(subject)
}

func (m *MockMailerSimple) Body(body string) {
	m.Called(body)
}

func (m *MockMailerSimple) Send() {
	m.Called()
}
