package mocks

import "github.com/stretchr/testify/mock"

type MockUUID struct {
	mock.Mock
}

func (m *MockUUID) GenerateUUID() string {
	args := m.Called()
	return args.String(0)
}
