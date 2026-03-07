package service

import (
	"errors"
	"fmt"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"github.com/google/uuid"
)

type PurchaseService interface {
	CreatePurchase(ebookId uint, clients []uint) error
	CreatePurchaseWithResult(ebookId uint, clientId uint) (*salesmodel.Purchase, error)
	GetPurchasesByCreatorIDWithFilters(creatorID uint, ebookID *uint, clientName, clientEmail string, page, limit int) ([]salesmodel.Purchase, int64, error)
	BlockDownload(purchaseID uint, creatorID uint, block bool) error
	GetPurchaseByID(id uint) (*salesmodel.Purchase, error)
	GetEbookFile(hashID string, fileID uint) (string, error)
	GetEbookFiles(purchaseID int) ([]*librarymodel.File, error)
}

type PurchaseServiceImpl struct {
	purchaseRepository *salesrepo.PurchaseRepository
	mailService        IEmailService
}

func NewPurchaseService(purchaseRepository *salesrepo.PurchaseRepository, mailService IEmailService) PurchaseService {
	return &PurchaseServiceImpl{
		purchaseRepository: purchaseRepository,
		mailService:        mailService,
	}
}

func (ps *PurchaseServiceImpl) CreatePurchase(ebookId uint, clients []uint) error {
	var purchases []*salesmodel.Purchase

	for _, clientId := range clients {
		if clientId != 0 && ebookId != 0 {
			existingPurchase, err := ps.purchaseRepository.FindExistingPurchase(ebookId, uint(clientId))
			if err == nil && existingPurchase != nil {
				continue
			}

			uuidv7, err := uuid.NewV7()
			if err != nil {
				return errors.New(err.Error())
			}
			purchases = append(purchases, salesmodel.NewPurchase(ebookId, uint(clientId), uuidv7.String()))
		}
	}

	if len(purchases) == 0 {
		return nil
	}

	err := ps.purchaseRepository.CreateManyPurchases(purchases)
	if err != nil {
		return errors.New(err.Error())
	}

	go ps.mailService.SendLinkToDownload(purchases)
	return nil
}

// CreatePurchaseWithResult cria purchase e retorna a purchase criada ou existente
func (ps *PurchaseServiceImpl) CreatePurchaseWithResult(ebookId uint, clientId uint) (*salesmodel.Purchase, error) {
	if clientId == 0 || ebookId == 0 {
		return nil, errors.New("clientId e ebookId devem ser válidos")
	}

	existingPurchase, err := ps.purchaseRepository.FindExistingPurchase(ebookId, clientId)
	if err == nil && existingPurchase != nil {
		return existingPurchase, nil
	}

	purchase := salesmodel.NewPurchase(ebookId, clientId, utils.UuidV7())
	purchases := []*salesmodel.Purchase{purchase}

	err = ps.purchaseRepository.CreateManyPurchases(purchases)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	go ps.mailService.SendLinkToDownload(purchases)
	return purchase, nil
}

func (ps *PurchaseServiceImpl) GetPurchaseByID(id uint) (*salesmodel.Purchase, error) {
	return ps.purchaseRepository.FindByID(id)
}

func (ps *PurchaseServiceImpl) GetEbookFile(hashID string, fileID uint) (string, error) {
	purchase, err := ps.purchaseRepository.FindEbookByPurchaseHash(hashID)
	if err != nil {
		return "", errors.New(err.Error())
	}

	if purchase == nil {
		return "", errors.New("Compra não localizada!")
	}

	if !purchase.AvailableDownloads() {
		return "", errors.New("não é possível realizar o download, limite de downloads atingido")
	}

	if purchase.IsExpired() {
		return "", errors.New("não é possível realizar o download, o pedido está expirado")
	}

	var targetFile *librarymodel.File
	for _, file := range purchase.Ebook.Files {
		if file.ID == fileID {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return "", errors.New("arquivo não encontrado neste ebook")
	}

	watermarkText := fmt.Sprintf("%s - %s - %s", purchase.Client.Name, purchase.Client.CPF, purchase.Client.Email)
	outputFilePath, err := ApplyWatermark(targetFile.S3Key, watermarkText)
	if err != nil {
		return "", err
	}

	purchase.UseDownload()
	ps.purchaseRepository.Update(purchase)

	return outputFilePath, nil
}

// GetEbookFiles retorna todos os arquivos do ebook para um cliente
func (ps *PurchaseServiceImpl) GetEbookFiles(purchaseID int) ([]*librarymodel.File, error) {
	purchase, err := ps.purchaseRepository.FindByID(uint(purchaseID))
	if err != nil {
		return nil, errors.New(err.Error())
	}

	if !purchase.AvailableDownloads() {
		return nil, errors.New("não é possível realizar o download, limite de downloads atingido")
	}

	if purchase.IsExpired() {
		return nil, errors.New("não é possível realizar o download, o pedido está expirado")
	}

	if len(purchase.Ebook.Files) == 0 {
		return nil, errors.New("nenhum arquivo encontrado neste ebook")
	}

	return purchase.Ebook.Files, nil
}

// GetPurchasesByCreatorIDWithFilters busca vendas por ID do criador com filtros
func (ps *PurchaseServiceImpl) GetPurchasesByCreatorIDWithFilters(creatorID uint, ebookID *uint, clientName, clientEmail string, page, limit int) ([]salesmodel.Purchase, int64, error) {
	var ebookIDVal uint
	if ebookID != nil {
		ebookIDVal = *ebookID
	}

	purchases, total, err := ps.purchaseRepository.FindByCreatorIDWithFilters(creatorID, page, limit, ebookIDVal, clientName, clientEmail)
	if err != nil {
		return nil, 0, err
	}

	result := make([]salesmodel.Purchase, len(purchases))
	for i, purchase := range purchases {
		result[i] = *purchase
	}

	return result, total, nil
}

// BlockDownload bloqueia ou desbloqueia o download de uma compra
func (ps *PurchaseServiceImpl) BlockDownload(purchaseID uint, creatorID uint, block bool) error {
	purchase, err := ps.purchaseRepository.FindByID(purchaseID)
	if err != nil {
		return err
	}

	if purchase.Ebook.CreatorID != creatorID {
		return errors.New("unauthorized: purchase does not belong to this creator")
	}

	if block {
		purchase.DownloadLimit = purchase.DownloadsUsed
	} else {
		purchase.DownloadLimit = -1
	}

	return ps.purchaseRepository.Update(purchase)
}
