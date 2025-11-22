package mocks

import (
	"context"
	"errors"
	"testing"
	"time"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockPRRepositoryCreate(t *testing.T) {
	mockRepo := new(MockPRRepository)
	ctx := context.Background()
	now := time.Now()

	pr := &domain.PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Fix login bug",
		AuthorID:          "user-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"user-2"},
		CreatedAt:         &now,
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Create", ctx, pr).Return(nil).Once()

		err := mockRepo.Create(ctx, pr)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error", func(t *testing.T) {
		expectedErr := errors.New("database is on fire")
		mockRepo.On("Create", ctx, pr).Return(expectedErr).Once()

		err := mockRepo.Create(ctx, pr)

		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockPRRepositoryUpdate(t *testing.T) {
	mockRepo := new(MockPRRepository)
	ctx := context.Background()
	pr := &domain.PullRequest{PullRequestID: "pr-999"}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Update", ctx, pr).Return(nil).Once()
		assert.NoError(t, mockRepo.Update(ctx, pr))
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.On("Update", ctx, pr).Return(errors.New("optimistic lock failed")).Once()
		assert.Error(t, mockRepo.Update(ctx, pr))
		mockRepo.AssertExpectations(t)
	})
}

func TestMockPRRepositoryExists(t *testing.T) {
	mockRepo := new(MockPRRepository)
	ctx := context.Background()
	prID := "pr-777"

	t.Run("exists = true", func(t *testing.T) {
		mockRepo.On("Exists", ctx, prID).Return(true, nil).Once()

		exists, err := mockRepo.Exists(ctx, prID)

		require.NoError(t, err)
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("exists = false", func(t *testing.T) {
		mockRepo.On("Exists", ctx, prID).Return(false, nil).Once()

		exists, err := mockRepo.Exists(ctx, prID)

		require.NoError(t, err)
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error from repo", func(t *testing.T) {
		expectedErr := errors.New("scan error")
		mockRepo.On("Exists", ctx, prID).Return(false, expectedErr).Once()

		exists, err := mockRepo.Exists(ctx, prID)

		assert.False(t, exists)
		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockPRRepositoryGetByID(t *testing.T) {
	mockRepo := new(MockPRRepository)
	ctx := context.Background()
	prID := "pr-123"

	expectedPR := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   "Add new feature",
		AuthorID:          "user-42",
		Status:            domain.PRStatusMerged,
		AssignedReviewers: []string{"rev-1", "rev-2"},
		CreatedAt:         &time.Time{},
		MergedAt:          &time.Time{},
	}

	t.Run("found - returns PR", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, prID).Return(expectedPR, nil).Once()

		pr, err := mockRepo.GetByID(ctx, prID)

		require.NoError(t, err)
		assert.Equal(t, expectedPR, pr)
		assert.True(t, pr.IsMerged())
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found - returns nil and error", func(t *testing.T) {
		expectedErr := errors.New("pull request not found")
		mockRepo.On("GetByID", ctx, prID).Return((*domain.PullRequest)(nil), expectedErr).Once()

		pr, err := mockRepo.GetByID(ctx, prID)

		assert.Nil(t, pr)
		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("other error - still returns nil PR", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, prID).Return((*domain.PullRequest)(nil), errors.New("connection timeout")).Once()

		pr, err := mockRepo.GetByID(ctx, prID)

		assert.Nil(t, pr)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil PR with error - covers nil branch", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, "non-existent").Return(nil, errors.New("not found")).Once()

		pr, err := mockRepo.GetByID(ctx, "non-existent")

		assert.Nil(t, pr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestMockPRRepositoryGetByReviewer(t *testing.T) {
	mockRepo := new(MockPRRepository)
	ctx := context.Background()
	userID := "user-2"

	prs := []*domain.PullRequest{
		{
			PullRequestID:     "pr-1",
			PullRequestName:   "Refactor auth",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{userID},
		},
	}

	t.Run("returns PRs", func(t *testing.T) {
		mockRepo.On("GetByReviewer", ctx, userID).Return(prs, nil).Once()

		result, err := mockRepo.GetByReviewer(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, prs, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no PRs for reviewer", func(t *testing.T) {
		mockRepo.On("GetByReviewer", ctx, userID).Return([]*domain.PullRequest{}, nil).Once()

		result, err := mockRepo.GetByReviewer(ctx, userID)

		require.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.On("GetByReviewer", ctx, userID).Return(([]*domain.PullRequest)(nil), errors.New("db down")).Once()

		result, err := mockRepo.GetByReviewer(ctx, userID)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil slice with error - covers nil branch", func(t *testing.T) {
		mockRepo.On("GetByReviewer", ctx, "invalid-user").Return(nil, errors.New("user not found")).Once()

		result, err := mockRepo.GetByReviewer(ctx, "invalid-user")

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestMockPRRepositoryGetOpenByTeam(t *testing.T) {
	mockRepo := new(MockPRRepository)
	ctx := context.Background()
	team := "backend"

	openPRs := []*domain.PullRequest{
		{
			PullRequestID:   "pr-10",
			PullRequestName: "Add caching",
			Status:          domain.PRStatusOpen,
			AuthorID:        "user-10",
		},
	}

	t.Run("returns open PRs", func(t *testing.T) {
		mockRepo.On("GetOpenByTeam", ctx, team).Return(openPRs, nil).Once()

		result, err := mockRepo.GetOpenByTeam(ctx, team)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.False(t, result[0].IsMerged())
		mockRepo.AssertExpectations(t)
	})

	t.Run("no open PRs in team", func(t *testing.T) {
		mockRepo.On("GetOpenByTeam", ctx, team).Return([]*domain.PullRequest{}, nil).Once()

		result, err := mockRepo.GetOpenByTeam(ctx, team)

		require.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("GetOpenByTeam", ctx, team).Return(([]*domain.PullRequest)(nil), errors.New("query failed")).Once()

		result, err := mockRepo.GetOpenByTeam(ctx, team)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil slice with error - covers nil branch", func(t *testing.T) {
		mockRepo.On("GetOpenByTeam", ctx, "non-existent-team").Return(nil, errors.New("team not found")).Once()

		result, err := mockRepo.GetOpenByTeam(ctx, "non-existent-team")

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "team not found")
		mockRepo.AssertExpectations(t)
	})
}
