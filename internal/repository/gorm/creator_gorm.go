package gorm

import (
	"errors"
	"log"
	"log/slog"

	"gorm.io/gorm"

	"github.com/anglesson/simple-web-server/internal/models"
	"github.com/anglesson/simple-web-server/pkg/database"
)

type CreatorRepository struct {
	db *gorm.DB
}

func NewCreatorRepository(db *gorm.DB) *CreatorRepository {
	return &CreatorRepository{db}
}

func (cr *CreatorRepository) FindByID(id uint) (*models.Creator, error) {
	var creator models.Creator
	err := database.DB.First(&creator, id).Error
	if err != nil {
		log.Printf("creator isn't recovery by ID %d. error: %s", id, err.Error())
		return nil, errors.New("creator not found")
	}
	return &creator, nil
}

func (cr *CreatorRepository) FindCreatorByUserID(userID uint) (*models.Creator, error) {
	var creator models.Creator
	err := database.DB.
		First(&creator, "user_id = ?", userID).Error
	if err != nil {
		log.Printf("creator isn't recovery. error: %s", err.Error())
		return nil, errors.New("creator not found")
	}
	return &creator, nil
}

func (cr *CreatorRepository) FindCreatorByUserEmail(email string) (*models.Creator, error) {
	var creator models.Creator
	err := database.DB.
		Joins("JOIN users ON users.id = creators.user_id").
		First(&creator, "users.email = ?", email).Error
	if err != nil {
		log.Printf("creator isn't recovery. error: %s", err.Error())
		return nil, errors.New("creator not found")
	}
	return &creator, nil
}

func (cr *CreatorRepository) FindByCPF(cpf string) (*models.Creator, error) {
	var creator models.Creator
	err := database.DB.
		First(&creator, "cpf = ?", cpf).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Creator not found, but that's not an error
		}
		log.Printf("error finding creator by CPF: %s", err.Error())
		return nil, errors.New("error finding creator")
	}
	return &creator, nil
}

func (cr *CreatorRepository) Update(creator *models.Creator) error {
	err := cr.db.Save(&creator).Error
	if err != nil {
		slog.Error("failed to update creator", "error", err)
		return err
	}

	return nil
}

func (cr *CreatorRepository) Create(creator *models.Creator) error {
	err := database.DB.Create(&creator).Error
	if err != nil {
		log.Printf("fail on create 'creator': %s", err.Error())
		return errors.New("creator not found")
	}
	return nil
}
