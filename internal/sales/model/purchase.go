package model

import (
	"time"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/pkg/utils"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type Purchase struct {
	gorm.Model
	PublicID      string             `json:"public_id" gorm:"type:varchar(40);uniqueIndex"`
	EbookID       uint               `json:"ebook_id"`
	Ebook         librarymodel.Ebook `gorm:"foreignKey:EbookID"`
	ClientID      uint               `json:"client_id"`
	Client        Client             `gorm:"foreignKey:ClientID"`
	ExpiresAt     time.Time          `json:"expires_at"`
	DownloadsUsed int                `json:"downloads_used"`
	DownloadLimit int                `json:"download_limit"`
	HashID        string             `json:"purchase_id" gorm:"uniqueIndex:purchase_id_unique"`
	PaymentStatus PaymentStatus      `json:"payment_status" gorm:"type:varchar(20);default:'pending'"`
}

func (p *Purchase) BeforeCreate(tx *gorm.DB) error {
	if p.PublicID == "" {
		p.PublicID = utils.GeneratePublicID("pur_")
	}
	return nil
}

func NewPurchase(ebookID, clientID uint, hashID string) *Purchase {
	return &Purchase{
		EbookID:       ebookID,
		ClientID:      clientID,
		DownloadLimit: -1,
		HashID:        hashID,
		PaymentStatus: PaymentStatusPending,
	}
}

func (p *Purchase) IsPaymentConfirmed() bool {
	return p.PaymentStatus == PaymentStatusConfirmed
}

func (p *Purchase) AvailableDownloads() bool {
	if p.DownloadLimit == -1 {
		return true
	}

	if p.DownloadsUsed >= p.DownloadLimit {
		return false
	}

	return true
}

func (p *Purchase) IsExpired() bool {
	if p.ExpiresAt.IsZero() {
		return false
	}
	return p.ExpiresAt.Before(time.Now())
}

func (p *Purchase) UseDownload() {
	p.DownloadsUsed++
}
