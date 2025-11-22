package mocks

import (
	"context"
	"errors"
	"testing"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockTeamRepositoryCreate(t *testing.T) {
	mockRepo := new(MockTeamRepository)
	ctx := context.Background()

	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{UserID: "user-1", Username: "john_doe", IsActive: true},
			{UserID: "user-2", Username: "jane_smith", IsActive: true},
		},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Create", ctx, team).Return(nil).Once()

		err := mockRepo.Create(ctx, team)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.On("Create", ctx, team).Return(expectedErr).Once()

		err := mockRepo.Create(ctx, team)

		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockTeamRepositoryGetByName(t *testing.T) {
	mockRepo := new(MockTeamRepository)
	ctx := context.Background()
	teamName := "backend"

	expectedTeam := &domain.Team{
		TeamName: teamName,
		Members: []domain.TeamMember{
			{UserID: "user-1", Username: "john_doe", IsActive: true},
			{UserID: "user-2", Username: "jane_smith", IsActive: true},
			{UserID: "user-3", Username: "bob_wilson", IsActive: false},
		},
	}

	t.Run("found - returns team", func(t *testing.T) {
		mockRepo.On("GetByName", ctx, teamName).Return(expectedTeam, nil).Once()

		team, err := mockRepo.GetByName(ctx, teamName)

		require.NoError(t, err)
		assert.Equal(t, expectedTeam, team)
		assert.Equal(t, teamName, team.TeamName)
		assert.Len(t, team.Members, 3)
		assert.Equal(t, "john_doe", team.Members[0].Username)
		assert.True(t, team.Members[0].IsActive)
		assert.False(t, team.Members[2].IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found - returns nil and error", func(t *testing.T) {
		expectedErr := errors.New("team not found")
		mockRepo.On("GetByName", ctx, teamName).Return((*domain.Team)(nil), expectedErr).Once()

		team, err := mockRepo.GetByName(ctx, teamName)

		assert.Nil(t, team)
		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("other error - returns nil team", func(t *testing.T) {
		mockRepo.On("GetByName", ctx, teamName).Return((*domain.Team)(nil), errors.New("connection timeout")).Once()

		team, err := mockRepo.GetByName(ctx, teamName)

		assert.Nil(t, team)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil team with error - covers nil branch", func(t *testing.T) {
		nonExistentTeam := "non-existent"
		mockRepo.On("GetByName", ctx, nonExistentTeam).Return(nil, errors.New("team not found")).Once()

		team, err := mockRepo.GetByName(ctx, nonExistentTeam)

		assert.Nil(t, team)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "team not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestMockTeamRepositoryExists(t *testing.T) {
	mockRepo := new(MockTeamRepository)
	ctx := context.Background()
	teamName := "backend"

	t.Run("exists = true", func(t *testing.T) {
		mockRepo.On("Exists", ctx, teamName).Return(true, nil).Once()

		exists, err := mockRepo.Exists(ctx, teamName)

		require.NoError(t, err)
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("exists = false", func(t *testing.T) {
		mockRepo.On("Exists", ctx, teamName).Return(false, nil).Once()

		exists, err := mockRepo.Exists(ctx, teamName)

		require.NoError(t, err)
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error from repository", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.On("Exists", ctx, teamName).Return(false, expectedErr).Once()

		exists, err := mockRepo.Exists(ctx, teamName)

		assert.False(t, exists)
		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("exists with error - still returns bool value", func(t *testing.T) {
		mockRepo.On("Exists", ctx, "problem-team").Return(true, errors.New("partial failure")).Once()

		exists, err := mockRepo.Exists(ctx, "problem-team")

		assert.True(t, exists)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockTeamRepositoryEdgeCases(t *testing.T) {
	mockRepo := new(MockTeamRepository)
	ctx := context.Background()

	t.Run("empty team name", func(t *testing.T) {
		emptyTeamName := ""
		mockRepo.On("GetByName", ctx, emptyTeamName).Return((*domain.Team)(nil), errors.New("empty team name")).Once()

		team, err := mockRepo.GetByName(ctx, emptyTeamName)

		assert.Nil(t, team)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("team with no members", func(t *testing.T) {
		emptyTeam := &domain.Team{
			TeamName: "empty-team",
			Members:  []domain.TeamMember{},
		}
		mockRepo.On("GetByName", ctx, "empty-team").Return(emptyTeam, nil).Once()

		team, err := mockRepo.GetByName(ctx, "empty-team")

		require.NoError(t, err)
		assert.Equal(t, "empty-team", team.TeamName)
		assert.Empty(t, team.Members)
		mockRepo.AssertExpectations(t)
	})

	t.Run("team with inactive members only", func(t *testing.T) {
		inactiveTeam := &domain.Team{
			TeamName: "inactive-team",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "inactive_user", IsActive: false},
				{UserID: "user-2", Username: "another_inactive", IsActive: false},
			},
		}
		mockRepo.On("GetByName", ctx, "inactive-team").Return(inactiveTeam, nil).Once()

		team, err := mockRepo.GetByName(ctx, "inactive-team")

		require.NoError(t, err)
		assert.Equal(t, "inactive-team", team.TeamName)
		assert.Len(t, team.Members, 2)
		assert.False(t, team.Members[0].IsActive)
		assert.False(t, team.Members[1].IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("exists for non-existent team", func(t *testing.T) {
		nonExistentTeam := "non-existent"
		mockRepo.On("Exists", ctx, nonExistentTeam).Return(false, nil).Once()

		exists, err := mockRepo.Exists(ctx, nonExistentTeam)

		require.NoError(t, err)
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("team with mixed active/inactive members", func(t *testing.T) {
		mixedTeam := &domain.Team{
			TeamName: "mixed-team",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "active_user", IsActive: true},
				{UserID: "user-2", Username: "inactive_user", IsActive: false},
				{UserID: "user-3", Username: "another_active", IsActive: true},
			},
		}
		mockRepo.On("GetByName", ctx, "mixed-team").Return(mixedTeam, nil).Once()

		team, err := mockRepo.GetByName(ctx, "mixed-team")

		require.NoError(t, err)
		assert.Equal(t, "mixed-team", team.TeamName)
		assert.Len(t, team.Members, 3)

		assert.True(t, team.Members[0].IsActive)
		assert.False(t, team.Members[1].IsActive)
		assert.True(t, team.Members[2].IsActive)
		mockRepo.AssertExpectations(t)
	})
}
