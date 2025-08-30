package mocks

import (
	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/internal/repository"
	"github.com/stretchr/testify/mock"
)

type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(file *models.File) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *MockFileRepository) FindByID(id uint) (*models.File, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.File), args.Error(1)
}

func (m *MockFileRepository) FindByCreator(creatorID uint) ([]*models.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

func (m *MockFileRepository) FindByCreatorPaginated(query repository.FileQuery) ([]*models.File, int64, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.File), args.Get(1).(int64), args.Error(2)
}

func (m *MockFileRepository) Update(file *models.File) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFileRepository) FindByType(creatorID uint, fileType string) ([]*models.File, error) {
	args := m.Called(creatorID, fileType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}

func (m *MockFileRepository) FindActiveByCreator(creatorID uint) ([]*models.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.File), args.Error(1)
}
