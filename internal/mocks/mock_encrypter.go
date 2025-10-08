package mocks

import "github.com/stretchr/testify/mock"

type MockEncrypter struct {
	mock.Mock
}

func (m *MockEncrypter) HashPassword(password string) string {
	args := m.Called(password)
	return args.String(0)
}

func (m *MockEncrypter) CheckPasswordHash(hashedPassword, password string) bool {
	args := m.Called(hashedPassword, password)
	return args.Bool(0)
}

func (m *MockEncrypter) GenerateToken(length int) string {
	args := m.Called(length)
	return args.String(0)
}
