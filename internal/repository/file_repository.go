package repository

import (
	"log"

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
	Create(file *models.File) error
	FindByID(id uint) (*models.File, error)
	FindByCreator(creatorID uint) ([]*models.File, error)
	FindByCreatorPaginated(query FileQuery) ([]*models.File, int64, error)
	Update(file *models.File) error
	Delete(id uint) error
	FindByType(creatorID uint, fileType string) ([]*models.File, error)
	FindActiveByCreator(creatorID uint) ([]*models.File, error)
}

type GormFileRepository struct {
	db *gorm.DB
}

func NewGormFileRepository(db *gorm.DB) *GormFileRepository {
	return &GormFileRepository{db: db}
}

func (r *GormFileRepository) Create(file *models.File) error {
	log.Printf("Criando arquivo: Nome=%s, CreatorID=%d, Tipo=%s", file.Name, file.CreatorID, file.FileType)
	err := r.db.Create(file).Error
	log.Printf("Arquivo criado com sucesso, ID=%d, erro: %v", file.Model.ID, err)
	return err
}

func (r *GormFileRepository) FindByID(id uint) (*models.File, error) {
	var file models.File
	err := r.db.Preload("Creator").Preload("Ebooks").First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *GormFileRepository) FindByCreator(creatorID uint) ([]*models.File, error) {
	var files []*models.File
	log.Printf("Executando consulta FindByCreator para creatorID: %d", creatorID)
	err := r.db.Where("creator_id = ?", creatorID).Order("created_at DESC").Find(&files).Error
	log.Printf("Consulta FindByCreator retornou %d arquivos, erro: %v", len(files), err)
	return files, err
}

func (r *GormFileRepository) FindByCreatorPaginated(query FileQuery) ([]*models.File, int64, error) {
	var files []*models.File
	var total int64

	// Construir query base
	db := r.db.Where("creator_id = ?", query.CreatorID)

	// Aplicar filtros
	if query.FileType != "" {
		db = db.Where("file_type = ?", query.FileType)
	}

	if query.SearchTerm != "" {
		searchTerm := "%" + query.SearchTerm + "%"
		db = db.Where("original_name LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	// Contar total de registros
	if err := db.Model(&models.File{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar paginação
	if query.Pagination != nil {
		offset := (query.Pagination.Page - 1) * query.Pagination.Limit
		db = db.Offset(offset).Limit(query.Pagination.Limit)
	}

	// Executar query
	err := db.Order("created_at DESC").Find(&files).Error
	return files, total, err
}

func (r *GormFileRepository) Update(file *models.File) error {
	return r.db.Save(file).Error
}

func (r *GormFileRepository) Delete(id uint) error {
	return r.db.Delete(&models.File{}, id).Error
}

func (r *GormFileRepository) FindByType(creatorID uint, fileType string) ([]*models.File, error) {
	var files []*models.File
	err := r.db.Where("creator_id = ? AND file_type = ?", creatorID, fileType).Order("created_at DESC").Find(&files).Error
	return files, err
}

func (r *GormFileRepository) FindActiveByCreator(creatorID uint) ([]*models.File, error) {
	var files []*models.File
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, true).Order("created_at DESC").Find(&files).Error
	return files, err
}
