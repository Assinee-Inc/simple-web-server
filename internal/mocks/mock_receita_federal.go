package mocks

import (
	"github.com/anglesson/simple-web-server/pkg/gov"
	"github.com/stretchr/testify/mock"
)

type MockRFService struct {
	mock.Mock
}

func (m *MockRFService) ConsultaCPF(cpf, dataNascimento string) (*gov.ReceitaFederalResponse, error) {
	args := m.Called(cpf, dataNascimento)
	return args.Get(0).(*gov.ReceitaFederalResponse), args.Error(1)
}
