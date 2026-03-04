package service

import (
	"mime/multipart"
	"testing"

	"github.com/anglesson/simple-web-server/internal/models"
	libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// MockS3Storage para testes
type MockS3Storage struct {
	generateDownloadLinkFunc func(key string) string
}

func (m *MockS3Storage) UploadFile(file *multipart.FileHeader, key, cacheControl string) (string, error) {
	return "", nil
}

func (m *MockS3Storage) DeleteFile(key string) error {
	return nil
}

func (m *MockS3Storage) GenerateDownloadLink(key string) string {
	if m.generateDownloadLinkFunc != nil {
		return m.generateDownloadLinkFunc(key)
	}
	return "presigned-url"
}

func (m *MockS3Storage) GeneratePreviewLinkWithExpiration(key, contentType string, expirationSeconds int) string {
	return "preview-url"
}

func (m *MockS3Storage) GenerateDownloadLinkWithExpiration(key string, expirationSeconds int) string {
	return "presigned-url"
}

// MockEbookRepository para testes
type MockEbookRepository struct {
	findByIDFunc          func(id uint) (*models.Ebook, error)
	listEbooksForUserFunc func(userID uint, query libraryrepo.EbookQuery) (*[]models.Ebook, error)
}

func (m *MockEbookRepository) FindByID(id uint) (*models.Ebook, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return nil, nil
}

func (m *MockEbookRepository) ListEbooksForUser(userID uint, query libraryrepo.EbookQuery) (*[]models.Ebook, error) {
	if m.listEbooksForUserFunc != nil {
		return m.listEbooksForUserFunc(userID, query)
	}
	return nil, nil
}

func (m *MockEbookRepository) FindBySlug(slug string) (*models.Ebook, error) {
	return nil, nil
}

func (m *MockEbookRepository) Update(ebook *models.Ebook) error {
	return nil
}

func (m *MockEbookRepository) Create(ebook *models.Ebook) error {
	return nil
}

func (m *MockEbookRepository) Delete(id uint) error {
	return nil
}

func (m *MockEbookRepository) FindByCreator(creatorID uint) ([]*models.Ebook, error) {
	return nil, nil
}

func (m *MockEbookRepository) FindAll() ([]*models.Ebook, error) {
	return nil, nil
}

func (m *MockEbookRepository) FindActive() ([]*models.Ebook, error) {
	return nil, nil
}

func TestEbookService_GeneratePresignedImageURL(t *testing.T) {
	mockRepo := &MockEbookRepository{}
	mockS3Storage := &MockS3Storage{}

	svc := &EbookServiceImpl{
		ebookRepository: mockRepo,
		s3Storage:       mockS3Storage,
	}

	tests := []struct {
		name     string
		imageURL string
		expected string
		setup    func()
	}{
		{
			name:     "URL vazia",
			imageURL: "",
			expected: "",
		},
		{
			name:     "URL já pré-assinada",
			imageURL: "https://s3.amazonaws.com/bucket/object?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
			expected: "presigned-url",
		},
		{
			name:     "URL pública do S3",
			imageURL: "https://bucket.s3.region.amazonaws.com/ebook-covers/filename.jpg",
			expected: "presigned-url",
		},
		{
			name:     "URL externa (não S3)",
			imageURL: "https://example.com/image.jpg",
			expected: "presigned-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result := svc.generatePresignedImageURL(tt.imageURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEbookService_ExtractS3Key(t *testing.T) {
	mockRepo := &MockEbookRepository{}
	mockS3Storage := &MockS3Storage{}

	svc := &EbookServiceImpl{
		ebookRepository: mockRepo,
		s3Storage:       mockS3Storage,
	}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL S3 padrão",
			url:      "https://bucket.s3.region.amazonaws.com/ebook-covers/1753917672-1.png",
			expected: "ebook-covers/1753917672-1.png",
		},
		{
			name:     "URL S3 com subdomínio",
			url:      "https://my-bucket.s3.us-east-1.amazonaws.com/ebook-covers/test.jpg",
			expected: "ebook-covers/test.jpg",
		},
		{
			name:     "URL S3 sem subdomínio",
			url:      "https://bucket.s3.amazonaws.com/ebook-covers/file.png",
			expected: "ebook-covers/file.png",
		},
		{
			name:     "URL externa",
			url:      "https://example.com/image.jpg",
			expected: "image.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.extractS3Key(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEbookService_FindByID_WithPresignedImage(t *testing.T) {
	mockRepo := &MockEbookRepository{}
	mockS3Storage := &MockS3Storage{}

	svc := &EbookServiceImpl{
		ebookRepository: mockRepo,
		s3Storage:       mockS3Storage,
	}

	ebookID := uint(1)
	creator := models.Creator{Model: gorm.Model{ID: 1}, Name: "Test Creator"}
	ebook := &models.Ebook{
		Model:     gorm.Model{ID: ebookID},
		Title:     "Test Ebook",
		Image:     "https://bucket.s3.region.amazonaws.com/ebook-covers/test.jpg",
		CreatorID: creator.ID,
	}

	mockRepo.findByIDFunc = func(id uint) (*models.Ebook, error) {
		return ebook, nil
	}

	result, err := svc.FindByID(ebookID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "presigned-url", result.Image)
}

func TestEbookService_ListEbooksForUser_WithPresignedImages(t *testing.T) {
	mockRepo := &MockEbookRepository{}
	mockS3Storage := &MockS3Storage{}

	svc := &EbookServiceImpl{
		ebookRepository: mockRepo,
		s3Storage:       mockS3Storage,
	}

	userID := uint(1)
	creator := models.Creator{Model: gorm.Model{ID: 1}, Name: "Test Creator"}
	ebooks := &[]models.Ebook{
		{
			Model:     gorm.Model{ID: 1},
			Title:     "Ebook 1",
			Image:     "https://bucket.s3.region.amazonaws.com/ebook-covers/ebook1.jpg",
			CreatorID: creator.ID,
		},
		{
			Model:     gorm.Model{ID: 2},
			Title:     "Ebook 2",
			Image:     "https://bucket.s3.region.amazonaws.com/ebook-covers/ebook2.jpg",
			CreatorID: creator.ID,
		},
	}

	query := libraryrepo.EbookQuery{}

	mockRepo.listEbooksForUserFunc = func(userID uint, query libraryrepo.EbookQuery) (*[]models.Ebook, error) {
		return ebooks, nil
	}

	result, err := svc.ListEbooksForUser(userID, query)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, 2)
	assert.Equal(t, "presigned-url", (*result)[0].Image)
	assert.Equal(t, "presigned-url", (*result)[1].Image)
}
