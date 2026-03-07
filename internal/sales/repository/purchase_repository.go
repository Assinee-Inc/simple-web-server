package repository

import (
	"errors"
	"log"
	"log/slog"

	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"github.com/anglesson/simple-web-server/pkg/database"
	"gorm.io/gorm"
)

type PurchaseRepository struct {
}

func NewPurchaseRepository() *PurchaseRepository {
	return &PurchaseRepository{}
}

func (pr *PurchaseRepository) CreateManyPurchases(purchases []*salesmodel.Purchase) error {
	err := database.DB.CreateInBatches(purchases, 1000).Error
	if err != nil {
		log.Printf("[PURCHASE-REPOSITORY] ERROR: %s", err)
		return errors.New("falha no processamento do envio")
	}

	ids := make([]uint, len(purchases))
	for i, p := range purchases {
		ids[i] = p.ID
	}

	var loadedPurchases []*salesmodel.Purchase
	log.Printf("[PURCHASE-REPOSITORY] Buscando purchases com IDs: %v", ids)

	err = database.DB.
		Preload("Client").
		Preload("Ebook").
		Where("id IN ?", ids).
		Find(&loadedPurchases).Error
	if err != nil {
		log.Printf("[PURCHASE-REPOSITORY] LOAD ERROR: %s", err)
		return errors.New("falha ao carregar dados relacionados")
	}

	log.Printf("[PURCHASE-REPOSITORY] Encontradas %d purchases carregadas", len(loadedPurchases))
	if len(loadedPurchases) > 0 {
		log.Printf("[PURCHASE-REPOSITORY] Primeira purchase - EbookID: %d, ClientID: %d",
			loadedPurchases[0].EbookID, loadedPurchases[0].ClientID)
		if loadedPurchases[0].Ebook.ID > 0 {
			log.Printf("[PURCHASE-REPOSITORY] Ebook carregado: %s", loadedPurchases[0].Ebook.Title)
		}
		if loadedPurchases[0].Client.ID > 0 {
			log.Printf("[PURCHASE-REPOSITORY] Client carregado: %s", loadedPurchases[0].Client.Name)
		}
	}

	for i, loadedPurchase := range loadedPurchases {
		if i < len(purchases) {
			*purchases[i] = *loadedPurchase
		}
	}

	log.Printf("[PURCHASE-REPOSITORY] Carregadas %d purchases com relacionamentos", len(loadedPurchases))

	return nil
}

func (pr *PurchaseRepository) FindByID(id uint) (*salesmodel.Purchase, error) {
	var purchase salesmodel.Purchase
	log.Printf("Buscando a compra: %v", id)
	err := database.DB.Preload("Client").
		Preload("Ebook").
		Preload("Ebook.Files").
		First(&purchase, id).Error
	if err != nil {
		log.Printf("Erro na busca da compra: %s", err)
		return nil, errors.New("erro na busca da compra")
	}

	log.Printf("✅ Compra encontrada: ID=%d, DownloadsUsed=%d, DownloadLimit=%d",
		purchase.ID, purchase.DownloadsUsed, purchase.DownloadLimit)

	return &purchase, nil
}

func (pr *PurchaseRepository) Update(purchase *salesmodel.Purchase) error {
	if purchase.ID == 0 {
		log.Printf("error to update purchase: %v", purchase)
		return errors.New("erro ao atualizar downloads")
	}

	err := database.DB.Save(purchase).Error
	if err != nil {
		log.Printf("Erro na busca da compra: %s", err)
		return errors.New("erro na busca da compra")
	}

	return nil
}

func (pr *PurchaseRepository) FindExistingPurchase(ebookID uint, clientID uint) (*salesmodel.Purchase, error) {
	var purchase salesmodel.Purchase
	err := database.DB.
		Preload("Client").
		Preload("Ebook").
		Where("ebook_id = ? AND client_id = ?", ebookID, clientID).
		First(&purchase).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}

func (pr *PurchaseRepository) FindByCreatorIDWithFilters(creatorID uint, page, limit int, ebookID uint, clientName, clientEmail string) ([]*salesmodel.Purchase, int64, error) {
	var purchases []*salesmodel.Purchase
	var count int64

	offset := (page - 1) * limit

	query := database.DB.Preload("Client").Preload("Ebook").
		Joins("JOIN ebooks ON purchases.ebook_id = ebooks.id").
		Where("ebooks.creator_id = ?", creatorID)

	countQuery := database.DB.Model(&salesmodel.Purchase{}).
		Joins("JOIN ebooks ON purchases.ebook_id = ebooks.id").
		Where("ebooks.creator_id = ?", creatorID)

	if ebookID > 0 {
		query = query.Where("purchases.ebook_id = ?", ebookID)
		countQuery = countQuery.Where("purchases.ebook_id = ?", ebookID)
	}

	if clientName != "" {
		query = query.Joins("JOIN clients ON purchases.client_id = clients.id").
			Where("clients.name LIKE ?", "%"+clientName+"%")
		countQuery = countQuery.Joins("JOIN clients ON purchases.client_id = clients.id").
			Where("clients.name LIKE ?", "%"+clientName+"%")
	}

	if clientEmail != "" {
		if clientName == "" {
			query = query.Joins("JOIN clients ON purchases.client_id = clients.id")
			countQuery = countQuery.Joins("JOIN clients ON purchases.client_id = clients.id")
		}
		query = query.Where("clients.email LIKE ?", "%"+clientEmail+"%")
		countQuery = countQuery.Where("clients.email LIKE ?", "%"+clientEmail+"%")
	}

	err := countQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("purchases.created_at desc").
		Offset(offset).Limit(limit).
		Find(&purchases).Error
	if err != nil {
		return nil, 0, err
	}

	return purchases, count, nil
}

func (pr *PurchaseRepository) FindEbookByPurchaseHash(hashID string) (*salesmodel.Purchase, error) {
	var purchase salesmodel.Purchase

	slog.Info("Buscando a compra", "hashID", hashID)

	err := database.DB.Preload("Client").
		Preload("Ebook").
		Preload("Ebook.Files").
		Where("purchases.hash_id = ?", hashID).
		First(&purchase).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		slog.Error("Erro na busca da compra", "error", err)
		return nil, errors.New("erro na busca da compra")
	}

	slog.Info("Compra encontrada", "id", purchase.ID, "downloadsUsed", purchase.DownloadsUsed, "downloadLimit", purchase.DownloadLimit)

	return &purchase, nil
}
