package uuid_mocks

import "github.com/stretchr/testify/mock"

type MockUUID struct {
	mock.Mock
}

func (m *MockUUID) Generate() string {
	args := m.Called()
	return args.String(0)
}
