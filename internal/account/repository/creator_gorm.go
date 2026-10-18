package repository

import (
	"errors"
	"log"
	"log/slog"

	"gorm.io/gorm"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
)

type GormCreatorRepository struct {
	db *gorm.DB
}

func NewGormCreatorRepository(db *gorm.DB) *GormCreatorRepository {
	return &GormCreatorRepository{db}
}

func (cr *GormCreatorRepository) FindByID(id uint) (*accountmodel.Creator, error) {
	var creator accountmodel.Creator
	err := cr.db.First(&creator, id).Error
	if err != nil {
		log.Printf("creator isn't recovery by ID %d. error: %s", id, err.Error())
		return nil, errors.New("creator not found")
	}
	return &creator, nil
}

func (cr *GormCreatorRepository) FindByPublicID(publicID string) (*accountmodel.Creator, error) {
	var creator accountmodel.Creator
	err := cr.db.Where("public_id = ?", publicID).First(&creator).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("creator not found")
		}
		log.Printf("error finding creator by PublicID %s: %s", publicID, err.Error())
		return nil, errors.New("error finding creator")
	}
	return &creator, nil
}

func (cr *GormCreatorRepository) FindCreatorByUserID(userID uint) (*accountmodel.Creator, error) {
	var creator accountmodel.Creator
	err := cr.db.
		First(&creator, "user_id = ?", userID).Error
	if err != nil {
		log.Printf("creator isn't recovery. error: %s", err.Error())
		return nil, errors.New("creator not found")
	}
	return &creator, nil
}

func (cr *GormCreatorRepository) FindCreatorByUserEmail(email string) (*accountmodel.Creator, error) {
	var creator accountmodel.Creator
	err := cr.db.
		Joins("JOIN users ON users.id = creators.user_id").
		First(&creator, "users.email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback: If not found via JOIN, search creator by email directly
			var directCreator accountmodel.Creator
			errDirect := cr.db.Where("email = ?", email).First(&directCreator).Error
			if errDirect == nil {
				// Creator exists directly under this email. Let's lookup the user ID.
				var user struct {
					ID uint
				}
				errUser := cr.db.Table("users").Select("id").Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
				if errUser == nil {
					// Auto-heal the mismatched association
					log.Printf("Auto-healing mismatched creator relationship: user %d updated on creator %d (email %s)", user.ID, directCreator.ID, email)
					directCreator.UserID = user.ID
					if errSave := cr.db.Save(&directCreator).Error; errSave == nil {
						return &directCreator, nil
					} else {
						log.Printf("failed to save auto-healed creator: %v", errSave)
					}
				}
			}
		}
		log.Printf("creator isn't recovery. error: %s", err.Error())
		return nil, errors.New("creator not found")
	}
	return &creator, nil
}

func (cr *GormCreatorRepository) FindByCPF(cpf string) (*accountmodel.Creator, error) {
	var creator accountmodel.Creator
	err := cr.db.
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

func (cr *GormCreatorRepository) Update(creator *accountmodel.Creator) error {
	err := cr.db.Save(creator).Error
	if err != nil {
		slog.Error("failed to update creator", "error", err)
		return err
	}
	return nil
}

func (cr *GormCreatorRepository) Create(creator *accountmodel.Creator) error {
	err := cr.db.Create(creator).Error
	if err != nil {
		log.Printf("fail on create 'creator': %s", err.Error())
		return errors.New("creator not found")
	}
	return nil
}
