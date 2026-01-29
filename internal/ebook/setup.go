package ebook

import (
	"github.com/anglesson/simple-web-server/internal/platform/router"
	"github.com/anglesson/simple-web-server/internal/platform/uuid"
	"gorm.io/gorm"
)

type Module struct {
	handler *APIHandler
	manager *Manager
}

func Setup(db *gorm.DB, idGen uuid.IDGenerator, r router.Router) {
	repo := NewInMemoryRepository()
	mgr := NewManager(idGen, repo)
	h := NewHandler(mgr)

	h.RegisterRoutes(r)
}
