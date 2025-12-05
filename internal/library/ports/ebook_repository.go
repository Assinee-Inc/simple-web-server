package ports

import "github.com/anglesson/simple-web-server/internal/library/models"

type EbookRepositoryPort interface {
	Save(ebook *models.Ebook) error
	FindByParams(params ...any) ([]*models.Ebook, error)
}
