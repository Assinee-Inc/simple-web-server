package services

import (
	"errors"

	"github.com/anglesson/simple-web-server/internal/library/models"
	"github.com/anglesson/simple-web-server/internal/library/ports"
)

var ErrEbookAlreadyExist = errors.New("ebook already exists")

type EbookService struct {
	uuid      ports.UUIDGeneratorPort
	repo      ports.EbookRepositoryPort
	validator ports.Validator
}

func NewEbookService(uuid ports.UUIDGeneratorPort, ebookRepo ports.EbookRepositoryPort, validator ports.Validator) *EbookService {
	return &EbookService{uuid, ebookRepo, validator}
}

func (s *EbookService) CreateEbook(input models.Ebook) (*models.Ebook, error) {
	newEbook := new(models.Ebook)

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
