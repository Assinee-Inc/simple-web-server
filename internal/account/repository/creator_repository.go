package repository

import (
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
)

type CreatorRepository interface {
	FindCreatorByUserID(userID uint) (*accountmodel.Creator, error)
	FindCreatorByUserEmail(email string) (*accountmodel.Creator, error)
	FindByCPF(cpf string) (*accountmodel.Creator, error)
	Update(creator *accountmodel.Creator) error
	FindByID(id uint) (*accountmodel.Creator, error)
	FindByPublicID(publicID string) (*accountmodel.Creator, error)
	Create(creator *accountmodel.Creator) error
}
