package model

import "gorm.io/gorm"

type DownloadLog struct {
	gorm.Model
	PurchaseID uint      `json:"purchase_id"`
	Purchase   *Purchase `gorm:"foreignKey:PurchaseID"`
}
