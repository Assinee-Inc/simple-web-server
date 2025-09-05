package gorm

import (
	"errors"
	"log"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/pkg/database"
	"gorm.io/gorm"
)

type ClientGormRepository struct {
}

func NewClientGormRepository() *ClientGormRepository {
	return &ClientGormRepository{}
}

func (cr *ClientGormRepository) Save(client *models.Client) error {
	err := database.DB.FirstOrCreate(client, models.Client{
		CPF: client.CPF,
	}).Error
	if err != nil {
		log.Printf("Erro ao salvar dados: %s", err)
		return errors.New("erro ao salvar dados")
	}

	return nil
}

func (cr *ClientGormRepository) FindClientsByCreator(creator *models.Creator, query models.ClientFilter) (*[]models.Client, error) {
	var clients []models.Client

	err := database.DB.
		Offset(getOffset(query.Pagination)).
		Limit(getLimit(query.Pagination)).
		Model(&models.Client{}).
		Joins("JOIN client_creators ON client_creators.client_id = clients.id").
		Where("client_creators.creator_id = ?", creator.ID).
		Preload("Creators").
		Scopes(ContainsNameCpfEmailOrPhoneWith(query.Term)).
		Find(&clients).
		Error
	if err != nil {
		log.Printf("Erro na busca de clientes: %s", err)
		return nil, errors.New("erro na busca de clientes")
	}

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

func (cr *ClientGormRepository) FindByIDAndCreators(client *models.Client, clientID uint, creator string) error {
	err := database.DB.
		Preload("Creators").
		First(&client).
		Joins("JOIN client_creators ON client_creators.client_id = clients.id AND client_creators.creator_id = creators.id").
		Joins("JOIN creators ON creators.id = client_creators.creator_id").
		Where("clients.id = ? AND creators.email = ?", clientID, creator).
		Error
	if err != nil {
		log.Printf("Erro na busca do client: %s", err)
		return errors.New("não foi possível recuperar as informações do cliente")
	}
	return nil
}

func (cr *ClientGormRepository) InsertBatch(clients []*models.Client) error {
	err := database.DB.CreateInBatches(clients, 1000).Error
	if err != nil {
		log.Printf("[CLIENT-REPOSITORY] ERROR: %s", err)
		return errors.New("falha no processamento dos clientes")
	}
	return nil
}

func (cr *ClientGormRepository) FindByClientsWhereEbookNotSend(creator *models.Creator, query models.ClientFilter) (*[]models.Client, error) {
	var clients []models.Client
	err := database.DB.Debug().
		Offset(getOffset(query.Pagination)).
		Limit(getLimit(query.Pagination)).
		Model(&models.Client{}).
		Joins("JOIN client_creators ON client_creators.client_id = clients.id and client_creators.creator_id = ?", creator.ID).
		Where("clients.id NOT IN (SELECT client_id FROM purchases WHERE ebook_id = ?)", query.EbookID).
		Preload("Creators").
		Scopes(ContainsNameCpfEmailOrPhoneWith(query.Term)).
		Find(&clients).Error

	if err != nil {
		log.Printf("Erro na busca de clientes: %s", err)
		return nil, errors.New("erro na busca de clientes")
	}

	return &clients, nil
}

func (cr *ClientGormRepository) FindByClientsWhereEbookWasSend(creator *models.Creator, query models.ClientFilter) (*[]models.Client, error) {
	var clients []models.Client
	err := database.DB.Debug().
		Offset(getOffset(query.Pagination)).
		Limit(getLimit(query.Pagination)).
		Model(&models.Client{}).
		Joins("JOIN client_creators ON client_creators.client_id = clients.id and client_creators.creator_id = ?", creator.ID).
		Joins("JOIN purchases ON purchases.client_id = clients.id").
		Where("clients.id IN (SELECT client_id FROM purchases WHERE ebook_id = ?)", query.EbookID).
		Preload("Creators").
		Preload("Purchases").
		Scopes(ContainsNameCpfEmailOrPhoneWith(query.Term)).
		Find(&clients).Error

	if err != nil {
		log.Printf("Erro na busca de clientes: %s", err)
		return nil, errors.New("erro na busca de clientes")
	}

	return &clients, nil
}

func (cr *ClientGormRepository) FindByEmail(email string) (*models.Client, error) {
	var client models.Client
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

// Helper functions for pagination
func getOffset(pagination *models.Pagination) int {
	if pagination == nil {
		return 0
	}
	return (pagination.Page - 1) * pagination.Limit
}

func getLimit(pagination *models.Pagination) int {
	if pagination == nil {
		return 10 // default limit
	}
	return pagination.Limit
}
