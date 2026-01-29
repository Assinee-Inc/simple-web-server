package ebook

import (
	"context"

	"github.com/anglesson/simple-web-server/internal/platform/uuid"
)

type Manager struct {
	uuid uuid.IDGenerator
	repo Repository
}

func NewManager(uuid uuid.IDGenerator, ebookRepo Repository) *Manager {
	return &Manager{uuid, ebookRepo}
}

func (m *Manager) CreateEbook(ctx context.Context, ebook *Ebook) (*Ebook, error) {
	ebook.ID = m.uuid.Generate()

	if err := m.repo.Save(ctx, ebook); err != nil {
		return nil, err
	}

	return ebook, nil
}
