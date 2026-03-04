package gorm

import (
	accountrepo "github.com/anglesson/simple-web-server/internal/account/repository"
	"gorm.io/gorm"
)

// CreatorRepository is a type alias for accountrepo.GormCreatorRepository for backwards compatibility.
type CreatorRepository = accountrepo.GormCreatorRepository

// NewCreatorRepository creates a new GormCreatorRepository.
// Deprecated: Use accountrepo.NewGormCreatorRepository instead.
func NewCreatorRepository(db *gorm.DB) *accountrepo.GormCreatorRepository {
	return accountrepo.NewGormCreatorRepository(db)
}
