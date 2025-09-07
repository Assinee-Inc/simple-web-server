package repository

import (
	"strconv"
	"strings"

	"github.com/anglesson/simple-web-server/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepository interface {
	CreateTransaction(transaction *models.Transaction) error
	UpdateTransaction(transaction *models.Transaction) error
	FindByID(id uint) (*models.Transaction, error)
	FindByCreatorID(creatorID uint, page, limit int) ([]*models.Transaction, int64, error)
	FindByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*models.Transaction, int64, error)
	FindByPurchaseID(purchaseID uint) (*models.Transaction, error)
	UpdateTransactionStatus(id uint, status models.TransactionStatus) error
}

type transactionRepositoryImpl struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepositoryImpl{
		db: db,
	}
}

// CreateTransaction cria uma nova transação
func (r *transactionRepositoryImpl) CreateTransaction(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

// UpdateTransaction atualiza uma transação existente
func (r *transactionRepositoryImpl) UpdateTransaction(transaction *models.Transaction) error {
	return r.db.Save(transaction).Error
}

// FindByID busca uma transação pelo ID
func (r *transactionRepositoryImpl) FindByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Preload("Creator").Preload("Purchase").Preload("Purchase.Ebook").First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// FindByCreatorID busca transações por ID do criador
func (r *transactionRepositoryImpl) FindByCreatorID(creatorID uint, page, limit int) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var count int64

	offset := (page - 1) * limit

	// Contar total
	err := r.db.Model(&models.Transaction{}).Where("creator_id = ?", creatorID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Buscar transações paginadas
	err = r.db.Preload("Purchase").Preload("Purchase.Ebook").
		Where("creator_id = ?", creatorID).
		Order("created_at desc").
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, count, nil
}

// FindByCreatorIDWithFilters busca transações por ID do criador com filtros de busca e status
func (r *transactionRepositoryImpl) FindByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*models.Transaction, int64, error) {
	var transactions []*models.Transaction
	var count int64

	offset := (page - 1) * limit

	// Construir query base com preloads para permitir busca nos relacionamentos
	query := r.db.Preload("Purchase").Preload("Purchase.Ebook").Where("creator_id = ?", creatorID)
	countQuery := r.db.Model(&models.Transaction{}).Where("creator_id = ?", creatorID)

	// Aplicar filtro de status se fornecido e não for "Todos"
	if status != "" && status != "Todos" {
		var statusValue string
		switch status {
		case "Concluídas":
			statusValue = "completed"
		case "Pendentes":
			statusValue = "pending"
		case "Falha":
			statusValue = "failed"
		default:
			statusValue = status // Usar o valor direto se não for um dos mapeamentos
		}

		if statusValue != "" {
			query = query.Where("status = ?", statusValue)
			countQuery = countQuery.Where("status = ?", statusValue)
		}
	}

	// Para busca, vamos fazer uma abordagem em duas etapas se houver termo de busca
	if search != "" {
		// Primeiro, buscar IDs de transações que correspondem aos critérios
		var transactionIDs []uint

		// Buscar por ID da transação diretamente
		var directMatch uint
		if id, err := strconv.ParseUint(search, 10, 32); err == nil {
			directMatch = uint(id)
		}

		// Buscar transações através de ebooks
		var ebookTransactions []models.Transaction
		r.db.Preload("Purchase").Preload("Purchase.Ebook").
			Where("creator_id = ?", creatorID).
			Find(&ebookTransactions)

		searchLower := strings.ToLower(search)
		for _, t := range ebookTransactions {
			// Verificar se corresponde ao ID
			if t.ID == directMatch {
				transactionIDs = append(transactionIDs, t.ID)
				continue
			}

			// Verificar se corresponde ao título ou descrição do ebook
			if t.Purchase.ID != 0 && t.Purchase.Ebook.ID != 0 {
				titleMatch := strings.Contains(strings.ToLower(t.Purchase.Ebook.Title), searchLower)
				descMatch := strings.Contains(strings.ToLower(t.Purchase.Ebook.Description), searchLower)

				if titleMatch || descMatch {
					transactionIDs = append(transactionIDs, t.ID)
				}
			}
		}

		if len(transactionIDs) > 0 {
			query = query.Where("id IN (?)", transactionIDs)
			countQuery = countQuery.Where("id IN (?)", transactionIDs)
		} else {
			// Se não encontrou nenhum resultado, retornar vazio
			return []*models.Transaction{}, 0, nil
		}
	}

	// Contar total
	err := countQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Buscar transações paginadas
	err = query.Order("created_at desc").
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, count, nil
} // FindByPurchaseID busca uma transação pelo ID da compra
func (r *transactionRepositoryImpl) FindByPurchaseID(purchaseID uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Preload("Creator").Preload("Purchase").Preload("Purchase.Ebook").
		Where("purchase_id = ?", purchaseID).
		Order("created_at ASC").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// UpdateTransactionStatus atualiza o status de uma transação com proteção contra race conditions
func (r *transactionRepositoryImpl) UpdateTransactionStatus(id uint, status models.TransactionStatus) error {
	// Usar transação de banco para garantir integridade
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Buscar transação atual com bloqueio
		var transaction models.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&transaction, id).Error; err != nil {
			return err
		}

		// Verificar estado atual
		if transaction.Status == models.TransactionStatusCompleted {
			return nil // Já está completa, nada a fazer
		}

		// Atualizar status
		transaction.Status = status
		return tx.Save(&transaction).Error
	})
}
