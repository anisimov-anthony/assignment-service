package mongodb

import (
	"context"
	"testing"
	"time"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func setupTestDB(t *testing.T) (*Client, func()) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	uri := "mongodb://localhost:27017"
	dbName := "test_assignment_service"
	timeout := 5 * time.Second

	client, err := NewClient(ctx, uri, dbName, timeout, logger)
	if err != nil {
		t.Skipf("Skipping integration test - MongoDB not available: %v", err)
		return nil, nil
	}

	cleanup := func() {
		_ = client.Database().Drop(ctx)
		_ = client.Close(ctx)
	}

	return client, cleanup
}

func TestPRRepositoryCreate(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewPRRepository(client, logger)

	t.Run("successful creation", func(t *testing.T) {
		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		err := repo.Create(ctx, pr)

		assert.NoError(t, err)
	})

	t.Run("duplicate PR ID", func(t *testing.T) {
		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		err := repo.Create(ctx, pr)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPRExists, err)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewPRRepository(closedClient, logger)
		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-error",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		err := badRepo.Create(ctx, pr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create PR")
	})
}

func TestPRRepositoryGetByID(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewPRRepository(client, logger)

	now := time.Now()
	pr := &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "user-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"user-2"},
		CreatedAt:         &now,
	}

	err := repo.Create(ctx, pr)
	require.NoError(t, err)

	t.Run("successful get", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "pr-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pr-1", result.PullRequestID)
		assert.Equal(t, "Test PR", result.PullRequestName)
	})

	t.Run("PR not found", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "pr-nonexistent")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPRNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewPRRepository(closedClient, logger)

		result, err := badRepo.GetByID(ctx, "pr-1")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get PR by ID")
	})
}

func TestPRRepositoryUpdate(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewPRRepository(client, logger)

	now := time.Now()
	pr := &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "user-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"user-2"},
		CreatedAt:         &now,
	}

	err := repo.Create(ctx, pr)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		mergedAt := time.Now()
		pr.Status = domain.PRStatusMerged
		pr.MergedAt = &mergedAt

		err := repo.Update(ctx, pr)

		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, "pr-1")
		assert.NoError(t, err)
		assert.Equal(t, domain.PRStatusMerged, updated.Status)
		assert.NotNil(t, updated.MergedAt)
	})

	t.Run("update non-existent PR", func(t *testing.T) {
		nonExistentPR := &domain.PullRequest{
			PullRequestID:     "pr-nonexistent",
			PullRequestName:   "Non-existent PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		err := repo.Update(ctx, nonExistentPR)

		assert.NoError(t, err)
	})

	t.Run("database error during update - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewPRRepository(closedClient, logger)

		err := badRepo.Update(ctx, pr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update PR")
	})
}

func TestPRRepositoryExists(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewPRRepository(client, logger)

	now := time.Now()
	pr := &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "user-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"user-2"},
		CreatedAt:         &now,
	}

	err := repo.Create(ctx, pr)
	require.NoError(t, err)

	t.Run("PR exists", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "pr-1")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("PR does not exist", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "pr-nonexistent")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewPRRepository(closedClient, logger)

		exists, err := badRepo.Exists(ctx, "pr-1")

		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "failed to check PR existence")
	})
}

func TestPRRepositoryGetByReviewer(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewPRRepository(client, logger)

	now := time.Now()
	pr1 := &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "PR 1",
		AuthorID:          "user-1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"reviewer-1"},
		CreatedAt:         &now,
	}

	pr2 := &domain.PullRequest{
		PullRequestID:     "pr-2",
		PullRequestName:   "PR 2",
		AuthorID:          "user-2",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"reviewer-1", "reviewer-2"},
		CreatedAt:         &now,
	}

	err := repo.Create(ctx, pr1)
	require.NoError(t, err)
	err = repo.Create(ctx, pr2)
	require.NoError(t, err)

	t.Run("get PRs by reviewer", func(t *testing.T) {
		prs, err := repo.GetByReviewer(ctx, "reviewer-1")

		assert.NoError(t, err)
		assert.Len(t, prs, 2)
	})

	t.Run("no PRs for reviewer", func(t *testing.T) {
		prs, err := repo.GetByReviewer(ctx, "reviewer-nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, prs)
	})

	t.Run("database error during find - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewPRRepository(closedClient, logger)

		prs, err := badRepo.GetByReviewer(ctx, "reviewer-1")

		assert.Error(t, err)
		assert.Nil(t, prs)
		assert.Contains(t, err.Error(), "failed to find PRs by reviewer")
	})
}

func TestPRRepositoryGetOpenByTeam(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("get open PRs", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewPRRepository(client, logger)

		now := time.Now()
		openPRs := []*domain.PullRequest{
			{
				PullRequestID:     "pr-open-1",
				PullRequestName:   "Open PR 1",
				AuthorID:          "user-1",
				Status:            domain.PRStatusOpen,
				AssignedReviewers: []string{"reviewer-1"},
				CreatedAt:         &now,
			},
			{
				PullRequestID:     "pr-open-2",
				PullRequestName:   "Open PR 2",
				AuthorID:          "user-2",
				Status:            domain.PRStatusOpen,
				AssignedReviewers: []string{"reviewer-2"},
				CreatedAt:         &now,
			},
		}

		closedPR := &domain.PullRequest{
			PullRequestID:     "pr-closed-1",
			PullRequestName:   "Closed PR",
			AuthorID:          "user-3",
			Status:            domain.PRStatusMerged,
			AssignedReviewers: []string{"reviewer-1"},
			CreatedAt:         &now,
			MergedAt:          &now,
		}

		for _, pr := range openPRs {
			err := repo.Create(ctx, pr)
			require.NoError(t, err)
		}
		err := repo.Create(ctx, closedPR)
		require.NoError(t, err)

		prs, err := repo.GetOpenByTeam(ctx, "any-team")

		assert.NoError(t, err)
		assert.Len(t, prs, 2)

		for _, pr := range prs {
			assert.Equal(t, domain.PRStatusOpen, pr.Status)
			assert.Contains(t, []string{"pr-open-1", "pr-open-2"}, pr.PullRequestID)
		}
	})

	t.Run("no open PRs", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewPRRepository(client, logger)

		prs, err := repo.GetOpenByTeam(ctx, "any-team")

		assert.NoError(t, err)
		assert.Empty(t, prs)
	})

	t.Run("database error during find - covers error logging branch", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}

		repo := NewPRRepository(client, logger)
		client.Close(ctx)

		prs, err := repo.GetOpenByTeam(ctx, "any-team")

		assert.Error(t, err)
		assert.Nil(t, prs)
		assert.Contains(t, err.Error(), "failed to find open PRs")

		cleanup()
	})
}

func TestPRRepositoryIndexes(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewPRRepository(client, logger)

	t.Run("check unique index on pull_request_id", func(t *testing.T) {
		now := time.Now()
		pr1 := &domain.PullRequest{
			PullRequestID:     "duplicate-id",
			PullRequestName:   "PR 1",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		err := repo.Create(ctx, pr1)
		assert.NoError(t, err)

		pr2 := &domain.PullRequest{
			PullRequestID:     "duplicate-id",
			PullRequestName:   "PR 2",
			AuthorID:          "user-3",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-4"},
			CreatedAt:         &now,
		}

		err = repo.Create(ctx, pr2)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrPRExists, err)
	})
}

func TestPRRepositoryDecodeErrors(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("error during decode in GetByReviewer", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		collection := client.Database().Collection("pull_requests")
		_, err := collection.InsertOne(ctx, bson.M{
			"pull_request_id":    "bad-pr-1",
			"status":             12345,
			"pull_request_name":  "Bad PR",
			"author_id":          "user-1",
			"assigned_reviewers": []string{"reviewer-1"},
			"created_at":         time.Now(),
		})
		require.NoError(t, err)

		repo := NewPRRepository(client, logger)

		prs, err := repo.GetByReviewer(ctx, "reviewer-1")

		assert.Error(t, err)
		assert.Nil(t, prs)
		assert.Contains(t, err.Error(), "failed to decode PRs")
	})
}
