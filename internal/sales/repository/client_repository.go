package repository

import (
	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
)

type ClientRepository interface {
	Save(client *salesmodel.Client) error
	FindClientsByCreator(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error)
	FindByIDAndCreators(client *salesmodel.Client, clientID uint, creator string) error
	FindByPublicID(publicID string) (*salesmodel.Client, error)
	FindByClientsWhereEbookNotSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error)
	FindByClientsWhereEbookWasSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error)
	FindClientsByPurchasesFromCreator(creator *accountmodel.Creator) (*[]salesmodel.Client, error)
	FindByEmail(email string) (*salesmodel.Client, error)
	FindByCPF(cpf string) (*salesmodel.Client, error)
}
