package model

import (
	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	"gorm.io/gorm"
)

type ClientCreator struct {
	gorm.Model
	ClientID  uint                `json:"client_id"`
	Client    Client              `gorm:"foreignKey:ClientID"`
	CreatorID uint                `json:"creator_id"`
	Creator   accountmodel.Creator `gorm:"foreignKey:CreatorID"`
}
