package repository

import (
	"context"

	"assignment-service/internal/domain"
)

type UserRepository interface {
	CreateOrUpdate(ctx context.Context, user *domain.User) error

	GetByID(ctx context.Context, userID string) (*domain.User, error)

	GetActiveByTeam(ctx context.Context, teamName string) ([]*domain.User, error)

	UpdateIsActive(ctx context.Context, userID string, isActive bool) error

	GetByTeam(ctx context.Context, teamName string) ([]*domain.User, error)
}
