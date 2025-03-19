package service

import (
	"context"
	"github.com/HJyup/mlt-configuration/internal/store"
)

type Store interface {
	CreateConfiguration(ctx context.Context, userID int) (*store.Configuration, error)
	GetConfiguration(ctx context.Context, id int) (*store.Configuration, error)
	UpdateConfiguration(ctx context.Context, config store.Configuration) (*store.Configuration, error)
	DeleteConfiguration(ctx context.Context, userID int) error
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}
