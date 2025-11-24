package services

import (
	"errors"

	"github.com/anglesson/simple-web-server/internal/library/domain"
)

type UUIDGeneratorPort interface {
	GenerateUUID() string
}

type EbookServicePort interface {
	CreateEbook(input domain.Ebook) (*domain.Ebook, error)
}

type EbookRepositoryPort interface {
	Save(ebook *domain.Ebook) error
	FindByParams(params ...interface{}) ([]*domain.Ebook, error)
}

type EbookService struct {
	uuid UUIDGeneratorPort
	repo EbookRepositoryPort
}

func NewEbookService(uuid UUIDGeneratorPort, ebookRepo EbookRepositoryPort) *EbookService {
	return &EbookService{uuid, ebookRepo}
}

func (s *EbookService) CreateEbook(input domain.Ebook) (*domain.Ebook, error) {
	newEbook := new(domain.Ebook)

	newEbook = &input
	newEbook.ID = s.uuid.GenerateUUID()

	ebooks, _ := s.repo.FindByParams(newEbook.Title, newEbook.InfoProducerID)
	if len(ebooks) > 0 {
		return nil, ErrEbookAlreadyExist
	}

	err := newEbook.Validate()
	if err != nil {
		return nil, err
	}

	s.repo.Save(newEbook)

	return newEbook, nil
}

var ErrEbookAlreadyExist = errors.New("ebook already exists")
