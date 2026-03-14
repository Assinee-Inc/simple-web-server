package repository

import (
	"strconv"
	"strings"

	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepository interface {
	CreateTransaction(transaction *salesmodel.Transaction) error
	UpdateTransaction(transaction *salesmodel.Transaction) error
	FindByID(id uint) (*salesmodel.Transaction, error)
	FindByPublicID(publicID string) (*salesmodel.Transaction, error)
	FindByCreatorID(creatorID uint, page, limit int) ([]*salesmodel.Transaction, int64, error)
	FindByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*salesmodel.Transaction, int64, error)
	FindByPurchaseID(purchaseID uint) (*salesmodel.Transaction, error)
	UpdateTransactionStatus(id uint, status salesmodel.TransactionStatus) error
}

type transactionRepositoryImpl struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepositoryImpl{
		db: db,
	}
}

func (r *transactionRepositoryImpl) CreateTransaction(transaction *salesmodel.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepositoryImpl) UpdateTransaction(transaction *salesmodel.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepositoryImpl) FindByID(id uint) (*salesmodel.Transaction, error) {
	var transaction salesmodel.Transaction
	err := r.db.Preload("Creator").Preload("Purchase").Preload("Purchase.Ebook").First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepositoryImpl) FindByPublicID(publicID string) (*salesmodel.Transaction, error) {
	var transaction salesmodel.Transaction
	err := r.db.Preload("Creator").Preload("Purchase").Preload("Purchase.Ebook").
		Where("public_id = ?", publicID).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepositoryImpl) FindByCreatorID(creatorID uint, page, limit int) ([]*salesmodel.Transaction, int64, error) {
	var transactions []*salesmodel.Transaction
	var count int64

	offset := (page - 1) * limit

	err := r.db.Model(&salesmodel.Transaction{}).Where("creator_id = ?", creatorID).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

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

func (r *transactionRepositoryImpl) FindByCreatorIDWithFilters(creatorID uint, page, limit int, search, status string) ([]*salesmodel.Transaction, int64, error) {
	var transactions []*salesmodel.Transaction
	var count int64

	offset := (page - 1) * limit

	query := r.db.Preload("Purchase").Preload("Purchase.Ebook").Where("creator_id = ?", creatorID)
	countQuery := r.db.Model(&salesmodel.Transaction{}).Where("creator_id = ?", creatorID)

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
			statusValue = status
		}

		if statusValue != "" {
			query = query.Where("status = ?", statusValue)
			countQuery = countQuery.Where("status = ?", statusValue)
		}
	}

	if search != "" {
		var transactionIDs []uint

		var directMatch uint
		if id, err := strconv.ParseUint(search, 10, 32); err == nil {
			directMatch = uint(id)
		}

		var ebookTransactions []salesmodel.Transaction
		r.db.Preload("Purchase").Preload("Purchase.Ebook").
			Where("creator_id = ?", creatorID).
			Find(&ebookTransactions)

		searchLower := strings.ToLower(search)
		for _, t := range ebookTransactions {
			if t.ID == directMatch {
				transactionIDs = append(transactionIDs, t.ID)
				continue
			}

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
			return []*salesmodel.Transaction{}, 0, nil
		}
	}

	err := countQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at desc").
		Offset(offset).Limit(limit).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, count, nil
}

func (r *transactionRepositoryImpl) FindByPurchaseID(purchaseID uint) (*salesmodel.Transaction, error) {
	var transaction salesmodel.Transaction
	err := r.db.Preload("Creator").Preload("Purchase").Preload("Purchase.Ebook").
		Where("purchase_id = ?", purchaseID).
		Order("created_at ASC").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepositoryImpl) UpdateTransactionStatus(id uint, status salesmodel.TransactionStatus) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var transaction salesmodel.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&transaction, id).Error; err != nil {
			return err
		}

		if transaction.Status == salesmodel.TransactionStatusCompleted {
			return nil
		}

		transaction.Status = status
		return tx.Save(&transaction).Error
	})
}
