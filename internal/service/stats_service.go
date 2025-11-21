package service

import (
	"context"
	"fmt"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository"

	"go.uber.org/zap"
)

type StatsService struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewStatsService(
	prRepo repository.PRRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *StatsService {
	return &StatsService{
		prRepo:   prRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

type UserStats struct {
	UserID        string `json:"user_id"`
	AssignedCount int    `json:"assigned_count"`
	OpenPRCount   int    `json:"open_pr_count"`
	MergedPRCount int    `json:"merged_pr_count"`
}

func (s *StatsService) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}

	stats := &UserStats{
		UserID:        userID,
		AssignedCount: len(prs),
	}

	for _, pr := range prs {
		switch pr.Status {
		case domain.PRStatusOpen:
			stats.OpenPRCount++
		case domain.PRStatusMerged:
			stats.MergedPRCount++
		}
	}

	return stats, nil
}

func (s *StatsService) GetAllUserStats(ctx context.Context) ([]UserStats, error) {
	return []UserStats{}, nil
}
