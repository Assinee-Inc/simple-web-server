package gorm

type gormRepository struct{}

func NewGormRepository() *gormRepository {
	return &gormRepository{}
}
