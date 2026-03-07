package repository

import (
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
)

type ClientRepository interface {
	Save(client *salesmodel.Client) error
	FindClientsByCreator(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error)
	FindByIDAndCreators(client *salesmodel.Client, clientID uint, creator string) error
	FindByClientsWhereEbookNotSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error)
	FindByClientsWhereEbookWasSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error)
	InsertBatch(clients []*salesmodel.Client) error
	FindByEmail(email string) (*salesmodel.Client, error)
}
