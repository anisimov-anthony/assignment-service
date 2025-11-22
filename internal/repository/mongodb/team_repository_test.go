package mongodb

import (
	"context"
	"testing"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap/zaptest"
)

func TestTeamRepositoryCreate(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("successful creation", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
				{UserID: "user-2", Username: "user2", IsActive: true},
			},
		}

		err := repo.Create(ctx, team)

		assert.NoError(t, err)
	})

	t.Run("duplicate team name", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team1 := &domain.Team{
			TeamName: "team-1",
			Members:  []domain.TeamMember{},
		}

		err := repo.Create(ctx, team1)
		require.NoError(t, err)

		team2 := &domain.Team{
			TeamName: "team-1",
			Members:  []domain.TeamMember{},
		}

		err = repo.Create(ctx, team2)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTeamExists, err)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewTeamRepository(closedClient, logger)
		team := &domain.Team{
			TeamName: "team-error",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		err := badRepo.Create(ctx, team)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create team")
	})
}

func TestTeamRepositoryGetByName(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("successful get", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		err := repo.Create(ctx, team)
		require.NoError(t, err)

		result, err := repo.GetByName(ctx, "team-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "team-1", result.TeamName)
		assert.Len(t, result.Members, 1)
	})

	t.Run("team not found", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		result, err := repo.GetByName(ctx, "team-nonexistent")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTeamNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewTeamRepository(closedClient, logger)

		result, err := badRepo.GetByName(ctx, "team-1")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get team by name")
	})

	t.Run("error during decode - covers error logging branch", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		collection := client.Database().Collection("teams")
		_, err := collection.InsertOne(ctx, bson.M{
			"team_name": "bad-team",
			"members":   12345,
		})
		require.NoError(t, err)

		repo := NewTeamRepository(client, logger)

		result, err := repo.GetByName(ctx, "bad-team")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get team by name")
	})
}

func TestTeamRepositoryExists(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("team exists", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members:  []domain.TeamMember{},
		}

		err := repo.Create(ctx, team)
		require.NoError(t, err)

		exists, err := repo.Exists(ctx, "team-1")

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("team does not exist", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		exists, err := repo.Exists(ctx, "team-nonexistent")

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("database error - covers error logging branch", func(t *testing.T) {
		closedClient, _ := setupTestDB(t)
		if closedClient == nil {
			t.Skip("MongoDB not available")
		}
		closedClient.Close(ctx)

		badRepo := NewTeamRepository(closedClient, logger)

		exists, err := badRepo.Exists(ctx, "team-1")

		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "failed to check team existence")
	})
}

func TestTeamRepositoryIndexes(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("check unique index on team_name", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team1 := &domain.Team{
			TeamName: "duplicate-team",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		err := repo.Create(ctx, team1)
		assert.NoError(t, err)

		team2 := &domain.Team{
			TeamName: "duplicate-team",
			Members: []domain.TeamMember{
				{UserID: "user-2", Username: "user2", IsActive: true},
			},
		}

		err = repo.Create(ctx, team2)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrTeamExists, err)
	})
}

func TestTeamRepositoryEdgeCases(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	t.Run("empty team name", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "", // Пустое имя команды
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		err := repo.Create(ctx, team)
		assert.NoError(t, err)

		result, err := repo.GetByName(ctx, "")
		assert.NoError(t, err)
		assert.Equal(t, "", result.TeamName)
		assert.Len(t, result.Members, 1)

		exists, err := repo.Exists(ctx, "")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("team with no members", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "empty-team",
			Members:  []domain.TeamMember{},
		}

		err := repo.Create(ctx, team)
		assert.NoError(t, err)

		result, err := repo.GetByName(ctx, "empty-team")
		assert.NoError(t, err)
		assert.Equal(t, "empty-team", result.TeamName)
		assert.Empty(t, result.Members)
	})

	t.Run("team with inactive members", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "inactive-team",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: false},
				{UserID: "user-2", Username: "user2", IsActive: true},
				{UserID: "user-3", Username: "user3", IsActive: false},
			},
		}

		err := repo.Create(ctx, team)
		assert.NoError(t, err)

		result, err := repo.GetByName(ctx, "inactive-team")
		assert.NoError(t, err)
		assert.Equal(t, "inactive-team", result.TeamName)
		assert.Len(t, result.Members, 3)
		assert.False(t, result.Members[0].IsActive)
		assert.True(t, result.Members[1].IsActive)
		assert.False(t, result.Members[2].IsActive)
	})
}

func TestTeamRepositoryContextErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("error with cancelled context in GetByName", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		err := repo.Create(context.Background(), team)
		require.NoError(t, err)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := repo.GetByName(cancelledCtx, "team-1")

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("error with cancelled context in Exists", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members:  []domain.TeamMember{},
		}

		err := repo.Create(context.Background(), team)
		require.NoError(t, err)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		exists, err := repo.Exists(cancelledCtx, "team-1")

		assert.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("error with cancelled context in Create", func(t *testing.T) {
		client, cleanup := setupTestDB(t)
		if client == nil {
			t.Skip("MongoDB not available")
		}
		defer cleanup()

		repo := NewTeamRepository(client, logger)

		team := &domain.Team{
			TeamName: "team-cancelled",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		err := repo.Create(cancelledCtx, team)

		assert.Error(t, err)
	})
}
