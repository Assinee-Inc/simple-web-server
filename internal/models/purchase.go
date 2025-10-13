package models

import (
	"time"

	"gorm.io/gorm"
)

type Purchase struct {
	gorm.Model
	EbookID       uint      `json:"ebook_id"`
	Ebook         Ebook     `gorm:"foreignKey:EbookID"`
	ClientID      uint      `json:"client_id"`
	Client        Client    `gorm:"foreignKey:ClientID"`
	ExpiresAt     time.Time `json:"expires_at"`
	DownloadsUsed int       `json:"downloads_used"`
	DownloadLimit int       `json:"download_limit"`
	Downloads     []DownloadLog
	HashID        string `json:"purchase_id" gorm:"uniqueIndex:purchase_id_unique"`
}

func NewPurchase(ebookID, clientID uint, hashID string) *Purchase {
	return &Purchase{
		EbookID:       ebookID,
		ClientID:      clientID,
		DownloadLimit: -1,
		HashID:        hashID,
	}
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
	p.Downloads = append(p.Downloads, DownloadLog{
		Purchase: p,
	})
}
