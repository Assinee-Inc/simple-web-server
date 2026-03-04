package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	libraryrepo "github.com/anglesson/simple-web-server/internal/library/repository"
	"github.com/anglesson/simple-web-server/pkg/storage"
	"github.com/anglesson/simple-web-server/pkg/utils"
)

type EbookService interface {
	ListEbooksForUser(UserID uint, query libraryrepo.EbookQuery) (*[]librarymodel.Ebook, error)
	FindByID(id uint) (*librarymodel.Ebook, error)
	FindBySlug(slug string) (*librarymodel.Ebook, error)
	Update(ebook *librarymodel.Ebook) error
	Create(ebook *librarymodel.Ebook) error
	Delete(id uint) error
	GetEbooksByCreatorID(creatorID uint) ([]*librarymodel.Ebook, error)
}

type EbookServiceImpl struct {
	ebookRepository libraryrepo.EbookRepository
	s3Storage       storage.S3Storage
}

func NewEbookService(ebookRepository libraryrepo.EbookRepository, s3Storage storage.S3Storage) EbookService {
	return &EbookServiceImpl{
		ebookRepository: ebookRepository,
		s3Storage:       s3Storage,
	}
}

func (s *EbookServiceImpl) ListEbooksForUser(UserID uint, query libraryrepo.EbookQuery) (*[]librarymodel.Ebook, error) {
	ebooks, err := s.ebookRepository.ListEbooksForUser(UserID, query)
	if err != nil {
		return nil, err
	}

	for i := range *ebooks {
		if (*ebooks)[i].Image != "" {
			(*ebooks)[i].Image = s.generatePresignedImageURL((*ebooks)[i].Image)
		}
	}

	return ebooks, nil
}

func (s *EbookServiceImpl) FindByID(id uint) (*librarymodel.Ebook, error) {
	ebook, err := s.ebookRepository.FindByID(id)
	if err != nil {
		return nil, err
	}

	if ebook.Image != "" {
		ebook.Image = s.generatePresignedImageURL(ebook.Image)
	}

	return ebook, nil
}

func (s *EbookServiceImpl) FindBySlug(slug string) (*librarymodel.Ebook, error) {
	ebook, err := s.ebookRepository.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	if ebook != nil && ebook.Image != "" {
		ebook.Image = s.generatePresignedImageURL(ebook.Image)
	}

	return ebook, nil
}

func (s *EbookServiceImpl) Update(ebook *librarymodel.Ebook) error {
	return s.ebookRepository.Update(ebook)
}

func (s *EbookServiceImpl) Create(ebook *librarymodel.Ebook) error {
	existsEbook, err := s.ebookRepository.FindBySlug(ebook.Slug)
	if err != nil {
		slog.Error("Erro ao buscar ebook por slug %v. Detalhes: %s", ebook.Slug, err)
		return errors.New("erro ao criar ebook")
	}

	if existsEbook != nil {
		return errors.New(fmt.Sprintf("O título %s não pode ser utilizado. Tente outro.", ebook.Title))
	}

	ebook.TitleNormalized = utils.NormalizeText(ebook.Title)
	ebook.DescriptionNormalized = utils.NormalizeText(ebook.Description)
	return s.ebookRepository.Create(ebook)
}

func (s *EbookServiceImpl) Delete(id uint) error {
	err := s.ebookRepository.Delete(id)
	if err != nil {
		slog.Error("Erro ao remover ebook", "id", id, "error", err)
		return errors.New("não foi possível remover o ebook. Tente novamente mais tarde")
	}
	return nil
}

func (s *EbookServiceImpl) GetEbooksByCreatorID(creatorID uint) ([]*librarymodel.Ebook, error) {
	ebooks, err := s.ebookRepository.FindByCreator(creatorID)
	if err != nil {
		return nil, err
	}
	return ebooks, nil
}

func (s *EbookServiceImpl) generatePresignedImageURL(imageURL string) string {
	if imageURL == "" {
		return ""
	}

	if s.isS3PublicURL(imageURL) {
		key := s.extractS3Key(imageURL)
		return s.s3Storage.GenerateDownloadLink(key)
	}

	return imageURL
}

func (s *EbookServiceImpl) isS3PublicURL(url string) bool {
	return len(url) > 0 && (url[0:8] == "https://" || url[0:7] == "http://")
}

func (s *EbookServiceImpl) extractS3Key(url string) string {
	if len(url) > 8 && url[0:8] == "https://" {
		url = url[8:]
	} else if len(url) > 7 && url[0:7] == "http://" {
		url = url[7:]
	}

	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	firstSlash := -1
	for i, char := range url {
		if char == '/' {
			firstSlash = i
			break
		}
	}

	if firstSlash == -1 {
		return ""
	}

	amazonawsIndex := strings.Index(url, "amazonaws.com/")
	if amazonawsIndex != -1 {
		return url[amazonawsIndex+14:]
	}

	secondSlash := -1
	for i := firstSlash + 1; i < len(url); i++ {
		if url[i] == '/' {
			secondSlash = i
			break
		}
	}

	if secondSlash == -1 {
		return url[firstSlash+1:]
	}

	return url[secondSlash+1:]
}
