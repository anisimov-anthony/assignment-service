package mongodb

import (
	"context"
	"testing"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestUserRepositoryCreateOrUpdate(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewUserRepository(client, logger)

	t.Run("successful create", func(t *testing.T) {
		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: true,
		}

		err := repo.CreateOrUpdate(ctx, user)

		assert.NoError(t, err)
	})

	t.Run("successful update", func(t *testing.T) {
		user := &domain.User{
			UserID:   "user-1",
			Username: "updateduser",
			TeamName: "team-1",
			IsActive: false,
		}

		err := repo.CreateOrUpdate(ctx, user)

		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, "user-1")
		assert.NoError(t, err)
		assert.Equal(t, "updateduser", updated.Username)
		assert.False(t, updated.IsActive)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewUserRepository(closedClient, logger)
		user := &domain.User{
			UserID:   "user-error",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: true,
		}

		err := badRepo.CreateOrUpdate(ctx, user)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create or update user")
	})
}

func TestUserRepositoryGetByID(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewUserRepository(client, logger)

	user := &domain.User{
		UserID:   "user-1",
		Username: "testuser",
		TeamName: "team-1",
		IsActive: true,
	}

	err := repo.CreateOrUpdate(ctx, user)
	require.NoError(t, err)

	t.Run("successful get", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user-1", result.UserID)
		assert.Equal(t, "testuser", result.Username)
	})

	t.Run("user not found", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "user-nonexistent")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewUserRepository(closedClient, logger)

		result, err := badRepo.GetByID(ctx, "user-1")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get user by ID")
	})
}

func TestUserRepositoryGetActiveByTeam(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("get active users by team", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		user1 := &domain.User{
			UserID:   "user-1",
			Username: "user1",
			TeamName: "team-1",
			IsActive: true,
		}

		user2 := &domain.User{
			UserID:   "user-2",
			Username: "user2",
			TeamName: "team-1",
			IsActive: false,
		}

		user3 := &domain.User{
			UserID:   "user-3",
			Username: "user3",
			TeamName: "team-1",
			IsActive: true,
		}

		err := repo.CreateOrUpdate(ctx, user1)
		require.NoError(t, err)
		err = repo.CreateOrUpdate(ctx, user2)
		require.NoError(t, err)
		err = repo.CreateOrUpdate(ctx, user3)
		require.NoError(t, err)

		users, err := repo.GetActiveByTeam(ctx, "team-1")

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		for _, user := range users {
			assert.True(t, user.IsActive)
			assert.Equal(t, "team-1", user.TeamName)
		}
	})

	t.Run("no active users", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		users, err := repo.GetActiveByTeam(ctx, "team-nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("database error during find - covers error logging branch", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		client.Close(ctx)

		badRepo := NewUserRepository(client, logger)

		users, err := badRepo.GetActiveByTeam(ctx, "team-1")

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "failed to find active users by team")

		cleanup()
	})
}

func TestUserRepositoryUpdateIsActive(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewUserRepository(client, logger)

	user := &domain.User{
		UserID:   "user-1",
		Username: "testuser",
		TeamName: "team-1",
		IsActive: true,
	}

	err := repo.CreateOrUpdate(ctx, user)
	require.NoError(t, err)

	t.Run("successful update - activate", func(t *testing.T) {
		err := repo.UpdateIsActive(ctx, "user-1", false)

		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, "user-1")
		assert.NoError(t, err)
		assert.False(t, updated.IsActive)
	})

	t.Run("successful update - deactivate", func(t *testing.T) {
		err := repo.UpdateIsActive(ctx, "user-1", true)

		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, "user-1")
		assert.NoError(t, err)
		assert.True(t, updated.IsActive)
	})

	t.Run("user not found", func(t *testing.T) {
		err := repo.UpdateIsActive(ctx, "user-nonexistent", true)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewUserRepository(closedClient, logger)

		err := badRepo.UpdateIsActive(ctx, "user-1", true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user is_active")
	})
}

func TestUserRepositoryGetByTeam(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("get all users by team", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		user1 := &domain.User{
			UserID:   "user-1",
			Username: "user1",
			TeamName: "team-1",
			IsActive: true,
		}

		user2 := &domain.User{
			UserID:   "user-2",
			Username: "user2",
			TeamName: "team-1",
			IsActive: false,
		}

		err := repo.CreateOrUpdate(ctx, user1)
		require.NoError(t, err)
		err = repo.CreateOrUpdate(ctx, user2)
		require.NoError(t, err)

		users, err := repo.GetByTeam(ctx, "team-1")

		assert.NoError(t, err)
		assert.Len(t, users, 2)
		for _, user := range users {
			assert.Equal(t, "team-1", user.TeamName)
		}
	})

	t.Run("no users for team", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		users, err := repo.GetByTeam(ctx, "team-nonexistent")

		assert.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("database error during find - covers error logging branch", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		client.Close(ctx)

		badRepo := NewUserRepository(client, logger)

		users, err := badRepo.GetByTeam(ctx, "team-1")

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "failed to find users by team")

		cleanup()
	})
}

func TestUserRepositoryIndexes(t *testing.T) {
	client, cleanup := setupTestDB(t)
	if client == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	repo := NewUserRepository(client, logger)

	t.Run("check unique index on user_id", func(t *testing.T) {
		user1 := &domain.User{
			UserID:   "duplicate-id",
			Username: "user1",
			TeamName: "team-1",
			IsActive: true,
		}

		err := repo.CreateOrUpdate(ctx, user1)
		assert.NoError(t, err)

		user2 := &domain.User{
			UserID:   "duplicate-id",
			Username: "user2",
			TeamName: "team-2",
			IsActive: false,
		}

		err = repo.CreateOrUpdate(ctx, user2)
		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, "duplicate-id")
		assert.NoError(t, err)
		assert.Equal(t, "user2", updated.Username)
		assert.Equal(t, "team-2", updated.TeamName)
		assert.False(t, updated.IsActive)
	})
}

func TestUserRepositoryEdgeCases(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("empty user ID", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		user := &domain.User{
			UserID:   "",
			Username: "emptyuser",
			TeamName: "team-1",
			IsActive: true,
		}

		err := repo.CreateOrUpdate(ctx, user)
		assert.NoError(t, err)

		result, err := repo.GetByID(ctx, "")
		assert.NoError(t, err)
		assert.Equal(t, "", result.UserID)
		assert.Equal(t, "emptyuser", result.Username)
	})

	t.Run("user without team", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		user := &domain.User{
			UserID:   "user-no-team",
			Username: "noteamuser",
			TeamName: "",
			IsActive: true,
		}

		err := repo.CreateOrUpdate(ctx, user)
		assert.NoError(t, err)

		users, err := repo.GetByTeam(ctx, "")
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "user-no-team", users[0].UserID)
	})
}

func TestUserRepositoryDecodeErrors(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("error during decode with cancelled context", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		user1 := &domain.User{
			UserID:   "user-1",
			Username: "user1",
			TeamName: "team-1",
			IsActive: true,
		}
		err := repo.CreateOrUpdate(ctx, user1)
		require.NoError(t, err)

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		users, err := repo.GetActiveByTeam(cancelledCtx, "team-1")

		assert.Error(t, err)
		assert.Nil(t, users)
	})

	t.Run("error during decode in GetByTeam with cancelled context", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewUserRepository(client, logger)

		user1 := &domain.User{
			UserID:   "user-1",
			Username: "user1",
			TeamName: "team-1",
			IsActive: true,
		}
		err := repo.CreateOrUpdate(ctx, user1)
		require.NoError(t, err)

		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		users, err := repo.GetByTeam(cancelledCtx, "team-1")

		assert.Error(t, err)
		assert.Nil(t, users)
	})
}
