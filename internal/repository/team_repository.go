package repository

import (
	"context"

	"assignment-service/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error

	GetByName(ctx context.Context, teamName string) (*domain.Team, error)

	Exists(ctx context.Context, teamName string) (bool, error)
}
