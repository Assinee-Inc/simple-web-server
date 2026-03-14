package repository

import (
	"errors"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	"gorm.io/gorm"
)

type EbookQuery struct {
	Term       string
	Pagination *salesmodel.Pagination
}

type EbookRepository interface {
	Create(ebook *librarymodel.Ebook) error
	FindByID(id uint) (*librarymodel.Ebook, error)
	FindByCreator(creatorID uint) ([]*librarymodel.Ebook, error)
	FindBySlug(slug string) (*librarymodel.Ebook, error)
	Update(ebook *librarymodel.Ebook) error
	Delete(id uint) error
	FindAll() ([]*librarymodel.Ebook, error)
	FindActive() ([]*librarymodel.Ebook, error)
	ListEbooksForUser(userID uint, query EbookQuery) (*[]librarymodel.Ebook, error)
	RemoveFileAssociation(ebookID, fileID uint) error
	AppendFiles(ebookID uint, files []*librarymodel.File) error
}

type GormEbookRepository struct {
	db *gorm.DB
}

func NewGormEbookRepository(db *gorm.DB) *GormEbookRepository {
	return &GormEbookRepository{db: db}
}

func (r *GormEbookRepository) Create(ebook *librarymodel.Ebook) error {
	return r.db.Create(ebook).Error
}

func (r *GormEbookRepository) FindByID(id uint) (*librarymodel.Ebook, error) {
	var ebook librarymodel.Ebook
	err := r.db.Preload("Files").First(&ebook, id).Error
	if err != nil {
		return nil, err
	}

	if ebook.Files == nil {
		ebook.Files = []*librarymodel.File{}
	}

	return &ebook, nil
}

func (r *GormEbookRepository) FindByCreator(creatorID uint) ([]*librarymodel.Ebook, error) {
	var ebooks []*librarymodel.Ebook
	err := r.db.Where("creator_id = ?", creatorID).Preload("Files").Order("created_at DESC").Find(&ebooks).Error
	return ebooks, err
}

func (r *GormEbookRepository) FindBySlug(slug string) (*librarymodel.Ebook, error) {
	var ebook librarymodel.Ebook

	err := r.db.Where("slug = ?", slug).
		Preload("Files").
		First(&ebook).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if ebook.Files == nil {
		ebook.Files = []*librarymodel.File{}
	}

	return &ebook, nil
}

func (r *GormEbookRepository) Update(ebook *librarymodel.Ebook) error {
	return r.db.Omit("Files").Save(ebook).Error
}

func (r *GormEbookRepository) AppendFiles(ebookID uint, files []*librarymodel.File) error {
	ebook := &librarymodel.Ebook{}
	ebook.ID = ebookID
	return r.db.Model(ebook).Association("Files").Append(files)
}

func (r *GormEbookRepository) RemoveFileAssociation(ebookID, fileID uint) error {
	ebook := &librarymodel.Ebook{}
	ebook.ID = ebookID
	file := &librarymodel.File{}
	file.ID = fileID
	return r.db.Model(ebook).Association("Files").Delete(file)
}

func (r *GormEbookRepository) Delete(id uint) error {
	return r.db.Delete(&librarymodel.Ebook{}, id).Error
}

func (r *GormEbookRepository) FindAll() ([]*librarymodel.Ebook, error) {
	var ebooks []*librarymodel.Ebook
	err := r.db.Preload("Files").Order("created_at DESC").Find(&ebooks).Error
	return ebooks, err
}

func (r *GormEbookRepository) FindActive() ([]*librarymodel.Ebook, error) {
	var ebooks []*librarymodel.Ebook
	err := r.db.Where("status = ?", true).Preload("Files").Order("created_at DESC").Find(&ebooks).Error
	return ebooks, err
}

func (r *GormEbookRepository) ListEbooksForUser(userID uint, query EbookQuery) (*[]librarymodel.Ebook, error) {
	var ebooks []librarymodel.Ebook

	db := r.db.Preload("Files")

	db = db.Joins("JOIN creators ON ebooks.creator_id = creators.id").
		Where("creators.user_id = ?", userID)

	if query.Term != "" {
		db = db.Where("ebooks.title_normalized LIKE ? OR ebooks.description_normalized LIKE ?", "%"+query.Term+"%", "%"+query.Term+"%")
	}

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
