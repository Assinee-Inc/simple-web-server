package repository

import (
	"log"

	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	"github.com/anglesson/simple-web-server/internal/models"
	"gorm.io/gorm"
)

type FileQuery struct {
	CreatorID  uint
	FileType   string
	SearchTerm string
	Pagination *models.Pagination
}

type FileRepository interface {
	Create(file *librarymodel.File) error
	FindByID(id uint) (*librarymodel.File, error)
	FindByCreator(creatorID uint) ([]*librarymodel.File, error)
	FindByCreatorPaginated(query FileQuery) ([]*librarymodel.File, int64, error)
	Update(file *librarymodel.File) error
	Delete(id uint) error
	FindByType(creatorID uint, fileType string) ([]*librarymodel.File, error)
	FindActiveByCreator(creatorID uint) ([]*librarymodel.File, error)
}

type GormFileRepository struct {
	db *gorm.DB
}

func NewGormFileRepository(db *gorm.DB) *GormFileRepository {
	return &GormFileRepository{db: db}
}

func (r *GormFileRepository) Create(file *librarymodel.File) error {
	log.Printf("Criando arquivo: Nome=%s, CreatorID=%d, Tipo=%s", file.Name, file.CreatorID, file.FileType)
	err := r.db.Create(file).Error
	log.Printf("Arquivo criado com sucesso, ID=%d, erro: %v", file.Model.ID, err)
	return err
}

func (r *GormFileRepository) FindByID(id uint) (*librarymodel.File, error) {
	var file librarymodel.File
	err := r.db.Preload("Ebooks").First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *GormFileRepository) FindByCreator(creatorID uint) ([]*librarymodel.File, error) {
	var files []*librarymodel.File
	log.Printf("Executando consulta FindByCreator para creatorID: %d", creatorID)
	err := r.db.Where("creator_id = ?", creatorID).Order("created_at DESC").Find(&files).Error
	log.Printf("Consulta FindByCreator retornou %d arquivos, erro: %v", len(files), err)
	return files, err
}

func (r *GormFileRepository) FindByCreatorPaginated(query FileQuery) ([]*librarymodel.File, int64, error) {
	var files []*librarymodel.File
	var total int64

	db := r.db.Where("creator_id = ?", query.CreatorID)

	if query.FileType != "" {
		db = db.Where("file_type = ?", query.FileType)
	}

	if query.SearchTerm != "" {
		searchTerm := "%" + query.SearchTerm + "%"
		db = db.Where(
			"name LIKE ? OR name_normalized LIKE ? OR description LIKE ? OR description_normalized LIKE ?",
			searchTerm,
			searchTerm,
			searchTerm,
			searchTerm,
		)
	}

	if err := db.Model(&librarymodel.File{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query.Pagination != nil {
		offset := (query.Pagination.Page - 1) * query.Pagination.Limit
		db = db.Offset(offset).Limit(query.Pagination.Limit)
	}

	err := db.Order("created_at DESC").Find(&files).Error
	return files, total, err
}

func (r *GormFileRepository) Update(file *librarymodel.File) error {
	return r.db.Save(file).Error
}

func (r *GormFileRepository) Delete(id uint) error {
	return r.db.Delete(&librarymodel.File{}, id).Error
}

func (r *GormFileRepository) FindByType(creatorID uint, fileType string) ([]*librarymodel.File, error) {
	var files []*librarymodel.File
	err := r.db.Where("creator_id = ? AND file_type = ?", creatorID, fileType).Order("created_at DESC").Find(&files).Error
	return files, err
}

func (r *GormFileRepository) FindActiveByCreator(creatorID uint) ([]*librarymodel.File, error) {
	var files []*librarymodel.File
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, true).Order("created_at DESC").Find(&files).Error
	return files, err
}
