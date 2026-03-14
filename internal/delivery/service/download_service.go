package service

import (
	"errors"
	"fmt"

	deliverymodel "github.com/anglesson/simple-web-server/internal/delivery/model"
	deliveryrepo "github.com/anglesson/simple-web-server/internal/delivery/repository"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	salesrepo "github.com/anglesson/simple-web-server/internal/sales/repository"
	salesvc "github.com/anglesson/simple-web-server/internal/sales/service"
)

// DownloadService gerencia a entrega de conteúdo digital adquirido.
type DownloadService interface {
	FindPurchaseByHash(hashID string) (*salesmodel.Purchase, error)
	GetEbookFile(hashID string, filePublicID string) (string, error)
	GetEbookFiles(purchaseID int) ([]*librarymodel.File, error)
}

type downloadServiceImpl struct {
	purchaseRepo *salesrepo.PurchaseRepository
	downloadRepo deliveryrepo.DownloadRepository
}

func NewDownloadService(purchaseRepo *salesrepo.PurchaseRepository, downloadRepo deliveryrepo.DownloadRepository) DownloadService {
	return &downloadServiceImpl{
		purchaseRepo: purchaseRepo,
		downloadRepo: downloadRepo,
	}
}

func (s *downloadServiceImpl) FindPurchaseByHash(hashID string) (*salesmodel.Purchase, error) {
	return s.purchaseRepo.FindEbookByPurchaseHash(hashID)
}

func (s *downloadServiceImpl) GetEbookFile(hashID string, filePublicID string) (string, error) {
	purchase, err := s.purchaseRepo.FindEbookByPurchaseHash(hashID)
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
		if file.PublicID == filePublicID {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return "", errors.New("arquivo não encontrado neste ebook")
	}

	watermarkText := fmt.Sprintf("%s - %s - %s", purchase.Client.Name, purchase.Client.CPF, purchase.Client.Email)
	outputFilePath, err := salesvc.ApplyWatermark(targetFile.S3Key, watermarkText)
	if err != nil {
		return "", err
	}

	purchase.UseDownload()
	s.purchaseRepo.Update(purchase)
	s.downloadRepo.Create(&deliverymodel.DownloadLog{PurchaseID: purchase.ID})

	return outputFilePath, nil
}

func (s *downloadServiceImpl) GetEbookFiles(purchaseID int) ([]*librarymodel.File, error) {
	purchase, err := s.purchaseRepo.FindByID(uint(purchaseID))
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
