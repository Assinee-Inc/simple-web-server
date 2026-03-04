package repository

import (
	"log"

	"github.com/anglesson/simple-web-server/pkg/database"
)

type DashboardRepository struct {
	UserID uint
}

func NewDashboardRepository(userID uint) *DashboardRepository {
	return &DashboardRepository{
		UserID: userID,
	}
}

// TODO: Create a service and use EbookRepository
func (dr *DashboardRepository) GetTotalEbooks() int64 {
	var count int64
	err := database.DB.
		Table("ebooks").
		Joins("JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Count(&count).Error
	if err != nil {
		log.Panicf("Erro na busca de totais: %s", err)
		return 0
	}

	return count
}

// TODO: Create a service and use EbookRepository
func (dr *DashboardRepository) GetTotalSendEbooks() int64 {
	var total int64
	err := database.DB.
		Table("purchases").
		Joins("JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Count(&total).Error
	if err != nil {
		log.Panicf("Erro na busca de totais: %s", err)
		return 0
	}

	return total
}

func (dr *DashboardRepository) GetTotalClients() int64 {
	var total int64
	err := database.DB.
		Table("clients").
		Joins("JOIN client_creators ON client_creators.client_id = clients.id").
		Joins("JOIN creators ON creators.id = client_creators.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Count(&total).Error
	if err != nil {
		log.Panicf("Erro na busca de totais: %s", err)
		return 0
	}

	return total
}

func (dr *DashboardRepository) GetEngagementMetric() float64 {
	var metric float64
	err := database.DB.
		Table("purchases").
		Select("ROUND(AVG(CAST(downloads_used AS float) / NULLIF(download_limit, 0)) * 100, 2) AS average_engagement").
		Joins("INNER JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Scan(&metric).Error

	if err != nil {
		log.Panicf("Erro na busca de engajamento: %s", err)
		return 0
	}

	return metric
}

// PurchaseSummary is used to return a simplified view of purchases for the dashboard.
type PurchaseSummary struct {
	ID         uint
	EbookTitle string
	CreatedAt  string
}

func (dr *DashboardRepository) GetLastPurchases() []PurchaseSummary {
	var purchases []PurchaseSummary

	err := database.DB.
		Table("purchases").
		Select("purchases.id, ebooks.title as ebook_title, purchases.created_at").
		Joins("INNER JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Order("purchases.created_at DESC").
		Limit(10).
		Scan(&purchases).Error

	if err != nil {
		log.Panicf("Erro ao buscar últimos envios/compras: %s", err)
	}

	return purchases
}

type EbookStats struct {
	ID              uint   `json:"id"`
	Title           string `json:"title"`
	TotalPurchases  int64  `json:"total_purchases"`
	TotalDownloads  int64  `json:"total_downloads"`
	TotalClients    int64  `json:"total_clients"`
	UniqueDownloads int64  `json:"unique_downloads"`
}

func (es *EbookStats) PercentDownloads() float64 {
	return float64((es.UniqueDownloads * 100) / es.TotalClients)
}

func (dr *DashboardRepository) GetEbookStats() ([]EbookStats, error) {
	var stats []EbookStats

	err := database.DB.
		Table("ebooks").
		Select(`
			ebooks.id,
			ebooks.title,
			COUNT(purchases.id) AS total_purchases,
			COALESCE(SUM(purchases.downloads_used), 0) AS total_downloads,
			COUNT(DISTINCT CASE WHEN purchases.downloads_used > 0 THEN purchases.client_id END) AS unique_downloads,
			COUNT(DISTINCT purchases.client_id) AS total_clients
		`).
		Joins("INNER JOIN purchases ON purchases.ebook_id = ebooks.id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Group("ebooks.id").
		Scan(&stats).Error

	if err != nil {
		log.Printf("Erro ao buscar estatísticas dos ebooks: %v", err)
		return nil, err
	}

	return stats, nil
}

type DailyStats struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TopEbook struct {
	Title          string `json:"title"`
	TotalPurchases int64  `json:"total_purchases"`
}

type TopClient struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	TotalPurchases int64  `json:"total_purchases"`
}

type TopDownloadedEbook struct {
	Title          string `json:"title"`
	TotalDownloads int64  `json:"total_downloads"`
}

func (dr *DashboardRepository) GetDailyPurchases() ([]DailyStats, error) {
	var stats []DailyStats

	// Use date function directly in Where clause for SQLite
	err := database.DB.
		Debug().
		Table("purchases").
		Select("strftime('%Y-%m-%d', purchases.created_at) as date, COUNT(*) as count").
		Joins("INNER JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Where("purchases.created_at >= datetime('now', '-7 days', 'localtime')"). // Use localtime for correct timezone
		Group("strftime('%Y-%m-%d', purchases.created_at)").
		Order("date ASC").
		Scan(&stats).Error

	if err != nil {
		log.Printf("Erro ao buscar estatísticas diárias de envios: %v", err)
		return nil, err
	}

	return stats, nil
}

func (dr *DashboardRepository) GetDailyDownloads() ([]DailyStats, error) {
	var stats []DailyStats

	// Use date function directly in Where clause for SQLite
	err := database.DB.
		Table("purchases").
		Select("DATE(purchases.updated_at) as date, SUM(purchases.downloads_used) as count").
		Joins("INNER JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Where("purchases.updated_at >= date('now', '-7 days')"). // Direct date comparison
		Where("purchases.downloads_used > 0").
		Group("DATE(purchases.updated_at)").
		Order("date ASC").
		Scan(&stats).Error

	if err != nil {
		log.Printf("Erro ao buscar estatísticas diárias de downloads: %v", err)
		return nil, err
	}

	return stats, nil
}

func (dr *DashboardRepository) GetTopEbooks() ([]TopEbook, error) {
	var ebooks []TopEbook

	err := database.DB.
		Table("ebooks").
		Select("ebooks.title, COUNT(purchases.id) as total_purchases").
		Joins("INNER JOIN purchases ON purchases.ebook_id = ebooks.id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Group("ebooks.id").
		Order("total_purchases DESC").
		Limit(3).
		Scan(&ebooks).Error

	if err != nil {
		log.Printf("Erro ao buscar top ebooks: %v", err)
		return nil, err
	}

	return ebooks, nil
}

func (dr *DashboardRepository) GetTopClients() ([]TopClient, error) {
	var clients []TopClient

	err := database.DB.
		Table("clients").
		Select("clients.name, clients.email, COUNT(purchases.id) as total_purchases").
		Joins("INNER JOIN purchases ON purchases.client_id = clients.id").
		Joins("INNER JOIN ebooks ON ebooks.id = purchases.ebook_id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Group("clients.id, clients.email").
		Order("total_purchases DESC").
		Limit(10).
		Scan(&clients).Error

	if err != nil {
		log.Printf("Erro ao buscar top clientes: %v", err)
		return nil, err
	}

	return clients, nil
}

func (dr *DashboardRepository) GetTopDownloadedEbooks() ([]TopDownloadedEbook, error) {
	var ebooks []TopDownloadedEbook

	err := database.DB.
		Table("ebooks").
		Select("ebooks.title, SUM(purchases.downloads_used) as total_downloads").
		Joins("INNER JOIN purchases ON purchases.ebook_id = ebooks.id").
		Joins("INNER JOIN creators ON creators.id = ebooks.creator_id").
		Where("creators.user_id = ?", dr.UserID).
		Group("ebooks.id").
		Order("total_downloads DESC").
		Limit(3).
		Scan(&ebooks).Error

	if err != nil {
		log.Printf("Erro ao buscar top ebooks mais baixados: %v", err)
		return nil, err
	}

	return ebooks, nil
}
