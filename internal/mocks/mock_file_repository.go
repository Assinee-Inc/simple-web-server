package mocks

import (
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"
	"github.com/stretchr/testify/mock"
)

type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(file *librarymodel.File) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *MockFileRepository) FindByID(id uint) (*librarymodel.File, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileRepository) FindByPublicID(publicID string) (*librarymodel.File, error) {
	args := m.Called(publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileRepository) FindByCreator(creatorID uint) ([]*librarymodel.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileRepository) FindByCreatorPaginated(query libraryrepo.FileQuery) ([]*librarymodel.File, int64, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*librarymodel.File), args.Get(1).(int64), args.Error(2)
}

func (m *MockFileRepository) Update(file *librarymodel.File) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFileRepository) FindByType(creatorID uint, fileType string) ([]*librarymodel.File, error) {
	args := m.Called(creatorID, fileType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileRepository) FindActiveByCreator(creatorID uint) ([]*librarymodel.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}
