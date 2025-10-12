package repository

import (
	"errors"

	"github.com/anglesson/simple-web-server/internal/models"
	"gorm.io/gorm"
)

type EbookQuery struct {
	Term       string
	Pagination *models.Pagination
}

type EbookRepository interface {
	Create(ebook *models.Ebook) error
	FindByID(id uint) (*models.Ebook, error)
	FindByCreator(creatorID uint) ([]*models.Ebook, error)
	FindBySlug(slug string) (*models.Ebook, error)
	Update(ebook *models.Ebook) error
	Delete(id uint) error
	FindAll() ([]*models.Ebook, error)
	FindActive() ([]*models.Ebook, error)
	ListEbooksForUser(userID uint, query EbookQuery) (*[]models.Ebook, error)
}

type GormEbookRepository struct {
	db *gorm.DB
}

func NewGormEbookRepository(db *gorm.DB) *GormEbookRepository {
	return &GormEbookRepository{db: db}
}

func (r *GormEbookRepository) Create(ebook *models.Ebook) error {
	return r.db.Create(ebook).Error
}

func (r *GormEbookRepository) FindByID(id uint) (*models.Ebook, error) {
	var ebook models.Ebook
	err := r.db.Preload("Creator").Preload("Files").First(&ebook, id).Error
	if err != nil {
		return nil, err
	}

	// Garantir que Files seja inicializado como slice vazio se for nil
	if ebook.Files == nil {
		ebook.Files = []*models.File{}
	}

	return &ebook, nil
}

func (r *GormEbookRepository) FindByCreator(creatorID uint) ([]*models.Ebook, error) {
	var ebooks []*models.Ebook
	err := r.db.Where("creator_id = ?", creatorID).Preload("Files").Order("created_at DESC").Find(&ebooks).Error
	return ebooks, err
}

func (r *GormEbookRepository) FindBySlug(slug string) (*models.Ebook, error) {
	var ebook models.Ebook

	// Carregar o ebook com todos os relacionamentos necessários
	err := r.db.Where("slug = ?", slug).
		Preload("Creator").
		Preload("Files").
		First(&ebook).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Garantir que Files seja inicializado como slice vazio se for nil
	if ebook.Files == nil {
		ebook.Files = []*models.File{}
	}

	return &ebook, nil
}

func (r *GormEbookRepository) Update(ebook *models.Ebook) error {
	return r.db.Save(ebook).Error
}

func (r *GormEbookRepository) Delete(id uint) error {
	return r.db.Delete(&models.Ebook{}, id).Error
}

func (r *GormEbookRepository) FindAll() ([]*models.Ebook, error) {
	var ebooks []*models.Ebook
	err := r.db.Preload("Creator").Preload("Files").Order("created_at DESC").Find(&ebooks).Error
	return ebooks, err
}

func (r *GormEbookRepository) FindActive() ([]*models.Ebook, error) {
	var ebooks []*models.Ebook
	err := r.db.Where("status = ?", true).Preload("Creator").Preload("Files").Order("created_at DESC").Find(&ebooks).Error
	return ebooks, err
}

func (r *GormEbookRepository) ListEbooksForUser(userID uint, query EbookQuery) (*[]models.Ebook, error) {
	var ebooks []models.Ebook

	db := r.db.Preload("Creator").Preload("Files")

	// Buscar ebooks do criador associado ao usuário
	db = db.Joins("JOIN creators ON ebooks.creator_id = creators.id").
		Where("creators.user_id = ?", userID)

	// Aplicar filtro de título se fornecido
	if query.Term != "" {
		db = db.Where("ebooks.title_normalized LIKE ? OR ebooks.description_normalized LIKE ?", "%"+query.Term+"%", "%"+query.Term+"%")
	}

	// Aplicar paginação se fornecida
	if query.Pagination != nil {
		offset := (query.Pagination.Page - 1) * query.Pagination.Limit
		db = db.Offset(offset).Limit(query.Pagination.Limit)
	}

	err := db.Order("ebooks.created_at DESC").Find(&ebooks).Error
	if err != nil {
		return nil, err
	}

	return &ebooks, nil
}
