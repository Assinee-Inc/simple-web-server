package mocks

import (
	"context"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

type MockS3Storage struct {
	mock.Mock
}

func (m *MockS3Storage) UploadFile(file *multipart.FileHeader, key, cacheControl string) (string, error) {
	args := m.Called(file, key, cacheControl)
	return args.String(0), args.Error(1)
}

func (m *MockS3Storage) DeleteFile(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockS3Storage) GenerateDownloadLink(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockS3Storage) GenerateDownloadLinkWithExpiration(key string, expirationSeconds int) string {
	args := m.Called(key, expirationSeconds)
	return args.String(0)
}

func (m *MockS3Storage) GeneratePreviewLinkWithExpiration(key, contentType string, expirationSeconds int) string {
	args := m.Called(key, contentType, expirationSeconds)
	return args.String(0)
}

func (m *MockS3Storage) GeneratePresignedPutURL(ctx context.Context, key, contentType string, expirationSeconds int) (string, error) {
	args := m.Called(ctx, key, contentType, expirationSeconds)
	return args.String(0), args.Error(1)
}
