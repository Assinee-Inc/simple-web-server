package repository

import (
	"github.com/anglesson/simple-web-server/internal/models"
)

type CreatorRepository interface {
	FindCreatorByUserID(userID uint) (*models.Creator, error)
	FindCreatorByUserEmail(email string) (*models.Creator, error)
	FindByCPF(cpf string) (*models.Creator, error)
	Create(creator *models.Creator) error
}
