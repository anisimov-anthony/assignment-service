package repository

import (
	"context"

	"assignment-service/internal/domain"
)

type PRRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error

	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)

	Update(ctx context.Context, pr *domain.PullRequest) error

	Exists(ctx context.Context, prID string) (bool, error)

	GetByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error)

	GetOpenByTeam(ctx context.Context, teamName string) ([]*domain.PullRequest, error)
}
