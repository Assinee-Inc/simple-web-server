package gorm

import (
	"errors"
	"log"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/anglesson/simple-web-server/pkg/database"
	"gorm.io/gorm"
)

type ClientGormRepository struct {
}

func NewClientGormRepository() *ClientGormRepository {
	return &ClientGormRepository{}
}

func (cr *ClientGormRepository) Save(client *salesmodel.Client) error {
	var existingClient salesmodel.Client
	err := database.DB.Where("cpf = ?", client.CPF).First(&existingClient).Error

	if err != nil {
		err = database.DB.Create(client).Error
		if err != nil {
			log.Printf("Erro ao criar cliente: %s", err)
			return errors.New("erro ao salvar dados")
		}
		log.Printf("Novo cliente criado com ID=%d", client.ID)
	} else {
		originalClientID := client.ID
		client.ID = existingClient.ID
		client.CreatedAt = existingClient.CreatedAt

		isSimpleUpdate := originalClientID != 0 && originalClientID == existingClient.ID

		if isSimpleUpdate {
			err = database.DB.Save(client).Error
			if err != nil {
				log.Printf("Erro ao atualizar cliente: %s", err)
				return errors.New("erro ao salvar dados")
			}
			log.Printf("Cliente existente atualizado com ID=%d", client.ID)
		} else {
			err = database.DB.Save(client).Error
			if err != nil {
				log.Printf("Erro ao atualizar cliente: %s", err)
				return errors.New("erro ao salvar dados")
			}
			log.Printf("Cliente existente atualizado com ID=%d", client.ID)
		}
	}

	return nil
}

func (cr *ClientGormRepository) FindClientsByCreator(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error) {
	var clients []salesmodel.Client

	subquery := database.DB.Model(&salesmodel.Client{}).
		Select("clients.id").
		Distinct().
		Joins("JOIN purchases ON purchases.client_id = clients.id").
		Joins("JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Where("ebooks.creator_id = ?", creator.ID)

	err := database.DB.
		Offset(getOffset(query.Pagination)).
		Limit(getLimit(query.Pagination)).
		Model(&salesmodel.Client{}).
		Where("clients.id IN (?)", subquery).
		Preload("Purchases").
		Scopes(ContainsNameCpfEmailOrPhoneWith(query.Term)).
		Find(&clients).
		Error

	if err != nil {
		log.Printf("Erro na busca de clientes: %s", err)
		return nil, errors.New("erro na busca de clientes")
	}

	log.Printf("Encontrados %d clientes para o creator ID=%d", len(clients), creator.ID)
	return &clients, nil
}

func ContainsNameCpfEmailOrPhoneWith(term string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if term == "" {
			return db
		}
		searchTerm := "%" + term + "%"
		return db.Where("clients.name LIKE ? OR clients.cpf LIKE ? OR clients.email LIKE ? OR clients.phone LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}
}

func (cr *ClientGormRepository) FindByIDAndCreators(client *salesmodel.Client, clientID uint, creator string) error {
	if creator == "" {
		err := database.DB.
			First(client, clientID).
			Error
		if err != nil {
			log.Printf("Erro na busca do client por ID: %s", err)
			return errors.New("não foi possível recuperar as informações do cliente")
		}
		return nil
	}

	subquery := database.DB.Model(&salesmodel.Client{}).
		Select("clients.id").
		Distinct().
		Joins("JOIN purchases ON purchases.client_id = clients.id").
		Joins("JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("JOIN creators as ec ON ec.id = ebooks.creator_id").
		Where("ec.email = ? AND clients.id = ?", creator, clientID)

	err := database.DB.
		Preload("Purchases").
		Where("id IN (?)", subquery).
		First(client).
		Error
	if err != nil {
		log.Printf("Erro na busca do client: %s", err)
		return errors.New("não foi possível recuperar as informações do cliente")
	}
	return nil
}

func (cr *ClientGormRepository) FindByPublicID(publicID string) (*salesmodel.Client, error) {
	var client salesmodel.Client
	err := database.DB.Preload("Purchases").Where("public_id = ?", publicID).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cliente não encontrado")
		}
		log.Printf("Erro ao buscar cliente por PublicID: %s", err)
		return nil, errors.New("erro ao buscar cliente")
	}
	return &client, nil
}

func (cr *ClientGormRepository) FindByClientsWhereEbookNotSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error) {
	var clients []salesmodel.Client
	err := database.DB.Debug().
		Offset(getOffset(query.Pagination)).
		Limit(getLimit(query.Pagination)).
		Model(&salesmodel.Client{}).
		Joins("JOIN purchases ON purchases.client_id = clients.id").
		Joins("JOIN ebooks ON ebooks.id = purchases.ebook_id AND ebooks.creator_id = ?", creator.ID).
		Where("clients.id NOT IN (SELECT client_id FROM purchases WHERE ebook_id = ?)", query.EbookID).
		Preload("Purchases").
		Scopes(ContainsNameCpfEmailOrPhoneWith(query.Term)).
		Find(&clients).Error

	if err != nil {
		log.Printf("Erro na busca de clientes: %s", err)
		return nil, errors.New("erro na busca de clientes")
	}

	return &clients, nil
}

func (cr *ClientGormRepository) FindByClientsWhereEbookWasSend(creator *accountmodel.Creator, query salesmodel.ClientFilter) (*[]salesmodel.Client, error) {
	var clients []salesmodel.Client
	err := database.DB.Debug().
		Distinct("clients.*").
		Offset(getOffset(query.Pagination)).
		Limit(getLimit(query.Pagination)).
		Model(&salesmodel.Client{}).
		Joins("JOIN purchases ON purchases.client_id = clients.id").
		Joins("JOIN ebooks ON ebooks.id = purchases.ebook_id AND ebooks.creator_id = ?", creator.ID).
		Where("clients.id IN (SELECT client_id FROM purchases WHERE ebook_id = ?)", query.EbookID).
		Preload("Purchases").
		Scopes(ContainsNameCpfEmailOrPhoneWith(query.Term)).
		Find(&clients).Error

	if err != nil {
		log.Printf("Erro na busca de clientes: %s", err)
		return nil, errors.New("erro na busca de clientes")
	}

	return &clients, nil
}

func (cr *ClientGormRepository) FindClientsByPurchasesFromCreator(creator *accountmodel.Creator) (*[]salesmodel.Client, error) {
	var clients []salesmodel.Client
	err := database.DB.
		Model(&salesmodel.Client{}).
		Distinct("clients.*").
		Joins("JOIN purchases ON purchases.client_id = clients.id").
		Joins("JOIN ebooks ON ebooks.id = purchases.ebook_id AND ebooks.creator_id = ?", creator.ID).
		Find(&clients).Error

	if err != nil {
		log.Printf("Erro na exportação de clientes: %s", err)
		return nil, errors.New("erro na exportação de clientes")
	}
	return &clients, nil
}

func (cr *ClientGormRepository) FindByEmail(email string) (*salesmodel.Client, error) {
	var client salesmodel.Client
	err := database.DB.Where("email = ?", email).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("Erro ao buscar cliente por email: %s", err)
		return nil, errors.New("erro ao buscar cliente")
	}
	return &client, nil
}

func (cr *ClientGormRepository) FindByCPF(cpf string) (*salesmodel.Client, error) {
	var client salesmodel.Client
	err := database.DB.Where("cpf = ?", cpf).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("Erro ao buscar cliente por CPF: %s", err)
		return nil, errors.New("erro ao buscar cliente")
	}
	return &client, nil
}

func getOffset(pagination *salesmodel.Pagination) int {
	if pagination == nil {
		return 0
	}
	return (pagination.Page - 1) * pagination.Limit
}

func getLimit(pagination *salesmodel.Pagination) int {
	if pagination == nil {
		return 10
	}
	return pagination.Limit
}