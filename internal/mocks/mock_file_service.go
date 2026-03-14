package mocks

import (
	"mime/multipart"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"
	"github.com/stretchr/testify/mock"
)

type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) UploadFile(file *multipart.FileHeader, name, description string, creatorID uint) (*librarymodel.File, error) {
	args := m.Called(file, name, description, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetFilesByCreator(creatorID uint) ([]*librarymodel.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetFilesByCreatorPaginated(creatorID uint, query libraryrepo.FileQuery) ([]*librarymodel.File, int64, error) {
	args := m.Called(creatorID, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*librarymodel.File), args.Get(1).(int64), args.Error(2)
}

func (m *MockFileService) GetActiveByCreator(creatorID uint) ([]*librarymodel.File, error) {
	args := m.Called(creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetFileByID(id uint) (*librarymodel.File, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileService) GetFileByPublicID(publicID string) (*librarymodel.File, error) {
	args := m.Called(publicID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*librarymodel.File), args.Error(1)
}

func (m *MockFileService) UpdateFile(id uint, name, description string) error {
	args := m.Called(id, name, description)
	return args.Error(0)
}

func (m *MockFileService) DeleteFile(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockFileService) GetFilesByType(creatorID uint, fileType string) ([]*librarymodel.File, error) {
	args := m.Called(creatorID, fileType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*librarymodel.File), args.Error(1)
}

func (m *MockFileService) ValidateFile(file *multipart.FileHeader) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *MockFileService) GetFileType(ext string) string {
	args := m.Called(ext)
	return args.String(0)
}
