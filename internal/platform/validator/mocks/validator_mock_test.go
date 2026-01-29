package validator_mocks

import "github.com/stretchr/testify/mock"

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(input any) error {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}
