package mocks

import (
	"github.com/anglesson/simple-web-server/internal/library/models"
	"github.com/stretchr/testify/mock"
)

type EbookServiceMock struct {
	mock.Mock
}

func (e *EbookServiceMock) CreateEbook(ebook models.Ebook) (*models.Ebook, error) {
	args := e.Called(ebook)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Ebook), nil
}
