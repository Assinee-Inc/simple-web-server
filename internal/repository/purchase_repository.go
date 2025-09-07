package repository

import (
	"errors"
	"log"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/pkg/database"
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
