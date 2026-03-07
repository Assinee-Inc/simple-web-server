package repository

import (
	deliverymodel "github.com/anglesson/simple-web-server/internal/delivery/model"
	"github.com/anglesson/simple-web-server/pkg/database"
)

type DownloadRepository interface {
	Create(log *deliverymodel.DownloadLog) error
	FindByPurchaseID(purchaseID uint) ([]*deliverymodel.DownloadLog, error)
}

type GormDownloadRepository struct{}

func NewGormDownloadRepository() DownloadRepository {
	return &GormDownloadRepository{}
}

func (r *GormDownloadRepository) Create(log *deliverymodel.DownloadLog) error {
	return database.DB.Create(log).Error
}

func (r *GormDownloadRepository) FindByPurchaseID(purchaseID uint) ([]*deliverymodel.DownloadLog, error) {
	var logs []*deliverymodel.DownloadLog
	err := database.DB.Where("purchase_id = ?", purchaseID).Find(&logs).Error
	return logs, err
}
