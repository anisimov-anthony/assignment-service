package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestStatsServiceGetUserStats(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful get stats", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewStatsService(mockPRRepo, mockUserRepo, logger)

		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: true,
		}

		now := time.Now()
		prs := []*domain.PullRequest{
			{
				PullRequestID:     "pr-1",
				PullRequestName:   "PR 1",
				AuthorID:          "author-1",
				Status:            domain.PRStatusOpen,
				AssignedReviewers: []string{"user-1"},
				CreatedAt:         &now,
			},
			{
				PullRequestID:     "pr-2",
				PullRequestName:   "PR 2",
				AuthorID:          "author-2",
				Status:            domain.PRStatusMerged,
				AssignedReviewers: []string{"user-1"},
				CreatedAt:         &now,
			},
			{
				PullRequestID:     "pr-3",
				PullRequestName:   "PR 3",
				AuthorID:          "author-3",
				Status:            domain.PRStatusOpen,
				AssignedReviewers: []string{"user-1"},
				CreatedAt:         &now,
			},
		}

		mockUserRepo.On("GetByID", ctx, "user-1").Return(user, nil)
		mockPRRepo.On("GetByReviewer", ctx, "user-1").Return(prs, nil)

		stats, err := service.GetUserStats(ctx, "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, "user-1", stats.UserID)
		assert.Equal(t, 3, stats.AssignedCount)
		assert.Equal(t, 2, stats.OpenPRCount)
		assert.Equal(t, 1, stats.MergedPRCount)
		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewStatsService(mockPRRepo, mockUserRepo, logger)

		mockUserRepo.On("GetByID", ctx, "user-1").Return(nil, domain.ErrUserNotFound)

		stats, err := service.GetUserStats(ctx, "user-1")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, stats)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("no PRs assigned", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewStatsService(mockPRRepo, mockUserRepo, logger)

		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: true,
		}

		mockUserRepo.On("GetByID", ctx, "user-1").Return(user, nil)
		mockPRRepo.On("GetByReviewer", ctx, "user-1").Return([]*domain.PullRequest{}, nil)

		stats, err := service.GetUserStats(ctx, "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, "user-1", stats.UserID)
		assert.Equal(t, 0, stats.AssignedCount)
		assert.Equal(t, 0, stats.OpenPRCount)
		assert.Equal(t, 0, stats.MergedPRCount)
		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})
}

func TestStatsServiceGetAllUserStats(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("returns empty list", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewStatsService(mockPRRepo, mockUserRepo, logger)

		stats, err := service.GetAllUserStats(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Empty(t, stats)
	})

	t.Run("error when fetching PRs by reviewer", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		logger := zap.NewNop()
		service := NewStatsService(mockPRRepo, mockUserRepo, logger)

		ctx := context.Background()

		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: true,
		}
		mockUserRepo.On("GetByID", ctx, "user-1").Return(user, nil)

		dbErr := fmt.Errorf("database connection timeout")
		mockPRRepo.On("GetByReviewer", ctx, "user-1").Return([]*domain.PullRequest(nil), dbErr)

		stats, err := service.GetUserStats(ctx, "user-1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get PRs by reviewer")
		assert.Contains(t, err.Error(), "database connection timeout")
		assert.Nil(t, stats)

		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})
}
