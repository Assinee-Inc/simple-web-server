package ports

import "github.com/anglesson/simple-web-server/internal/library/models"

type EbookService interface {
	CreateEbook(input models.Ebook) (*models.Ebook, error)
}
