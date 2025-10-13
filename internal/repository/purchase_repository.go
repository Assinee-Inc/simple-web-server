package repository

import (
	"errors"
	"log"
	"log/slog"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/pkg/database"
	"gorm.io/gorm"
)

type PurchaseRepository struct {
}

func NewPurchaseRepository() *PurchaseRepository {
	return &PurchaseRepository{}
}

func (pr *PurchaseRepository) CreateManyPurchases(purchases []*models.Purchase) error {
	err := database.DB.CreateInBatches(purchases, 1000).Error
	if err != nil {
		log.Printf("[PURCHASE-REPOSITORY] ERROR: %s", err)
		return errors.New("falha no processamento do envio")
	}

	// Recarrega as purchases com os dados relacionados
	ids := make([]uint, len(purchases))
	for i, p := range purchases {
		ids[i] = p.ID
	}

	// Carrega as purchases com relacionamentos usando uma nova query
	var loadedPurchases []*models.Purchase
	log.Printf("[PURCHASE-REPOSITORY] Buscando purchases com IDs: %v", ids)

	err = database.DB.
		Preload("Client").
		Preload("Ebook").
		Preload("Ebook.Creator").
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

	// Atualiza o slice original com os dados carregados
	for i, loadedPurchase := range loadedPurchases {
		if i < len(purchases) {
			*purchases[i] = *loadedPurchase
		}
	}

	log.Printf("[PURCHASE-REPOSITORY] Carregadas %d purchases com relacionamentos", len(loadedPurchases))

	return nil
}

func (pr *PurchaseRepository) FindByID(id uint) (*models.Purchase, error) {
	var purchase models.Purchase
	log.Printf("Buscando a compra: %v", id)
	err := database.DB.Preload("Client").
		Preload("Ebook.Creator").
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

func (pr *PurchaseRepository) Update(purchase *models.Purchase) error {
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

func (pr *PurchaseRepository) FindExistingPurchase(ebookID uint, clientID uint) (*models.Purchase, error) {
	var purchase models.Purchase
	err := database.DB.Where("ebook_id = ? AND client_id = ?", ebookID, clientID).First(&purchase).Error
	if err != nil {
		return nil, err // Pode ser "record not found" se não existir
	}
	return &purchase, nil
}

// FindByCreatorIDWithFilters busca purchases por ID do criador com filtros
func (pr *PurchaseRepository) FindByCreatorIDWithFilters(creatorID uint, page, limit int, ebookID uint, clientName, clientEmail string) ([]*models.Purchase, int64, error) {
	var purchases []*models.Purchase
	var count int64

	offset := (page - 1) * limit

	// Construir query base com preloads
	query := database.DB.Preload("Client").Preload("Ebook").Preload("Ebook.Creator").
		Joins("JOIN ebooks ON purchases.ebook_id = ebooks.id").
		Where("ebooks.creator_id = ?", creatorID)

	countQuery := database.DB.Model(&models.Purchase{}).
		Joins("JOIN ebooks ON purchases.ebook_id = ebooks.id").
		Where("ebooks.creator_id = ?", creatorID)

	// Aplicar filtro de ebook se fornecido
	if ebookID > 0 {
		query = query.Where("purchases.ebook_id = ?", ebookID)
		countQuery = countQuery.Where("purchases.ebook_id = ?", ebookID)
	}

	// Aplicar filtro de nome do cliente se fornecido
	if clientName != "" {
		query = query.Joins("JOIN clients ON purchases.client_id = clients.id").
			Where("clients.name LIKE ?", "%"+clientName+"%")
		countQuery = countQuery.Joins("JOIN clients ON purchases.client_id = clients.id").
			Where("clients.name LIKE ?", "%"+clientName+"%")
	}

	// Aplicar filtro de email do cliente se fornecido
	if clientEmail != "" {
		if clientName == "" {
			// Se ainda não fez o join com clients, fazer agora
			query = query.Joins("JOIN clients ON purchases.client_id = clients.id")
			countQuery = countQuery.Joins("JOIN clients ON purchases.client_id = clients.id")
		}
		query = query.Where("clients.email LIKE ?", "%"+clientEmail+"%")
		countQuery = countQuery.Where("clients.email LIKE ?", "%"+clientEmail+"%")
	}

	// Contar total
	err := countQuery.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Buscar purchases paginadas
	err = query.Order("purchases.created_at desc").
		Offset(offset).Limit(limit).
		Find(&purchases).Error
	if err != nil {
		return nil, 0, err
	}

	return purchases, count, nil
}

func (pr *PurchaseRepository) FindEbookByPurchaseHash(hashID string) (*models.Purchase, error) {
	var purchase models.Purchase

	slog.Info("Buscando a compra: %v", hashID)

	err := database.DB.Preload("Client").
		Preload("Ebook.Creator").
		Preload("Ebook.Files").
		Where("purchases.hash_id = ?", hashID).
		First(&purchase).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		slog.Error("Erro na busca da compra: %s", err)
		return nil, errors.New("erro na busca da compra")
	}

	slog.Info("✅ Compra encontrada: ID=%d, DownloadsUsed=%d, DownloadLimit=%d",
		purchase.ID, purchase.DownloadsUsed, purchase.DownloadLimit)

	return &purchase, nil
}
