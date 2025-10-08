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
	// Verificar se o cliente já existe pelo CPF
	var existingClient models.Client
	err := database.DB.Where("cpf = ?", client.CPF).First(&existingClient).Error

	if err != nil {
		// Cliente não existe, criar novo com as associações
		err = database.DB.Create(client).Error
		if err != nil {
			log.Printf("Erro ao criar cliente: %s", err)
			return errors.New("erro ao salvar dados")
		}
		log.Printf("Novo cliente criado com ID=%d e %d creators associados", client.ID, len(client.Creators))
	} else {
		// Cliente existe - atualizar dados básicos
		originalClientID := client.ID
		client.ID = existingClient.ID
		client.CreatedAt = existingClient.CreatedAt

		// Determinar se é uma atualização simples ou se precisa adicionar associações
		isSimpleUpdate := originalClientID != 0 && originalClientID == existingClient.ID

		if isSimpleUpdate {
			// É uma atualização simples (cliente já existia no objeto) - não mexer nas associações
			originalCreators := client.Creators
			client.Creators = nil

			err = database.DB.Save(client).Error
			if err != nil {
				log.Printf("Erro ao atualizar cliente: %s", err)
				return errors.New("erro ao salvar dados")
			}

			// Restaurar creators no objeto (mas não salvar no banco)
			client.Creators = originalCreators
			log.Printf("Cliente existente atualizado com ID=%d (sem alterar associações)", client.ID)
		} else {
			// É um caso onde precisamos adicionar associações (mesmo CPF, mas objeto client novo)
			err = database.DB.Save(client).Error
			if err != nil {
				log.Printf("Erro ao atualizar cliente: %s", err)
				return errors.New("erro ao salvar dados")
			}

			// Adicionar novas associações se necessário usando SQL direto
			if len(client.Creators) > 0 {
				for _, creator := range client.Creators {
					// Verificar se associação já existe
					var count int64
					err = database.DB.Table("client_creators").
						Where("client_id = ? AND creator_id = ?", client.ID, creator.ID).
						Count(&count).Error
					if err != nil {
						log.Printf("Erro ao verificar associação existente: %s", err)
						continue
					}

					// Criar associação apenas se não existir
					if count == 0 {
						err = database.DB.Exec("INSERT INTO client_creators (client_id, creator_id) VALUES (?, ?)",
							client.ID, creator.ID).Error
						if err != nil {
							log.Printf("Erro ao criar associação cliente-creator: %s", err)
							return errors.New("erro ao salvar associações")
						}
						log.Printf("Nova associação criada: cliente %d -> creator %d", client.ID, creator.ID)
					}
				}
			}
			log.Printf("Cliente existente atualizado com ID=%d e associações", client.ID)
		}
	}

	return nil
}

func (cr *ClientGormRepository) FindClientsByCreator(creator *models.Creator, query models.ClientFilter) (*[]models.Client, error) {
	var clients []models.Client

	// Buscar clientes de duas formas:
	// 1. Clientes associados diretamente ao creator (tabela client_creators)
	// 2. Clientes que fizeram compras de ebooks desse creator
	subquery := database.DB.Model(&models.Client{}).
		Select("clients.id").
		Distinct().
		Joins("LEFT JOIN client_creators ON client_creators.client_id = clients.id").
		Joins("LEFT JOIN purchases ON purchases.client_id = clients.id").
		Joins("LEFT JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Where("client_creators.creator_id = ? OR ebooks.creator_id = ?", creator.ID, creator.ID)

	err := database.DB.
		Offset(getOffset(query.Pagination)).
		Limit(getLimit(query.Pagination)).
		Model(&models.Client{}).
		Where("clients.id IN (?)", subquery).
		Preload("Creators").
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

func (cr *ClientGormRepository) FindByIDAndCreators(client *models.Client, clientID uint, creator string) error {
	// Se não foi fornecido o email do creator, buscar diretamente pelo ID
	if creator == "" {
		err := database.DB.
			Preload("Creators").
			First(client, clientID).
			Error
		if err != nil {
			log.Printf("Erro na busca do client por ID: %s", err)
			return errors.New("não foi possível recuperar as informações do cliente")
		}
		return nil
	}

	// Buscar cliente que tenha associação com o creator (direta ou por compra)
	subquery := database.DB.Model(&models.Client{}).
		Select("clients.id").
		Distinct().
		Joins("LEFT JOIN client_creators ON client_creators.client_id = clients.id").
		Joins("LEFT JOIN creators as cc ON cc.id = client_creators.creator_id").
		Joins("LEFT JOIN purchases ON purchases.client_id = clients.id").
		Joins("LEFT JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("LEFT JOIN creators as ec ON ec.id = ebooks.creator_id").
		Where("(cc.email = ? OR ec.email = ?) AND clients.id = ?", creator, creator, clientID)

	err := database.DB.
		Preload("Creators").
		Where("id IN (?)", subquery).
		First(client).
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
