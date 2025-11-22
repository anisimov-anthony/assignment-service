package mocks

import (
	"context"
	"errors"
	"testing"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockUserRepositoryCreateOrUpdate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()

	user := &domain.User{
		UserID:   "user-1",
		Username: "john_doe",
		TeamName: "backend",
		IsActive: true,
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("CreateOrUpdate", ctx, user).Return(nil).Once()

		err := mockRepo.CreateOrUpdate(ctx, user)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.On("CreateOrUpdate", ctx, user).Return(expectedErr).Once()

		err := mockRepo.CreateOrUpdate(ctx, user)

		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockUserRepositoryGetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	userID := "user-1"

	expectedUser := &domain.User{
		UserID:   userID,
		Username: "john_doe",
		TeamName: "backend",
		IsActive: true,
	}

	t.Run("found - returns user", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil).Once()

		user, err := mockRepo.GetByID(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		assert.Equal(t, userID, user.UserID)
		assert.Equal(t, "backend", user.TeamName)
		assert.True(t, user.IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found - returns nil and error", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockRepo.On("GetByID", ctx, userID).Return((*domain.User)(nil), expectedErr).Once()

		user, err := mockRepo.GetByID(ctx, userID)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("other error - returns nil user", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, userID).Return((*domain.User)(nil), errors.New("connection timeout")).Once()

		user, err := mockRepo.GetByID(ctx, userID)

		assert.Nil(t, user)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil user with error - covers nil branch", func(t *testing.T) {
		nonExistentUser := "non-existent"
		mockRepo.On("GetByID", ctx, nonExistentUser).Return(nil, errors.New("user not found")).Once()

		user, err := mockRepo.GetByID(ctx, nonExistentUser)

		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestMockUserRepositoryGetActiveByTeam(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	teamName := "backend"

	activeUsers := []*domain.User{
		{
			UserID:   "user-1",
			Username: "john_doe",
			TeamName: teamName,
			IsActive: true,
		},
		{
			UserID:   "user-2",
			Username: "jane_smith",
			TeamName: teamName,
			IsActive: true,
		},
	}

	t.Run("returns active users", func(t *testing.T) {
		mockRepo.On("GetActiveByTeam", ctx, teamName).Return(activeUsers, nil).Once()

		users, err := mockRepo.GetActiveByTeam(ctx, teamName)

		require.NoError(t, err)
		assert.Equal(t, activeUsers, users)
		assert.Len(t, users, 2)
		assert.True(t, users[0].IsActive)
		assert.True(t, users[1].IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no active users in team", func(t *testing.T) {
		mockRepo.On("GetActiveByTeam", ctx, teamName).Return([]*domain.User{}, nil).Once()

		users, err := mockRepo.GetActiveByTeam(ctx, teamName)

		require.NoError(t, err)
		assert.Empty(t, users)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.On("GetActiveByTeam", ctx, teamName).Return(([]*domain.User)(nil), errors.New("db down")).Once()

		users, err := mockRepo.GetActiveByTeam(ctx, teamName)

		assert.Nil(t, users)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil slice with error - covers nil branch", func(t *testing.T) {
		nonExistentTeam := "non-existent"
		mockRepo.On("GetActiveByTeam", ctx, nonExistentTeam).Return(nil, errors.New("team not found")).Once()

		users, err := mockRepo.GetActiveByTeam(ctx, nonExistentTeam)

		assert.Nil(t, users)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "team not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestMockUserRepositoryUpdateIsActive(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	userID := "user-1"

	t.Run("activate user - success", func(t *testing.T) {
		mockRepo.On("UpdateIsActive", ctx, userID, true).Return(nil).Once()

		err := mockRepo.UpdateIsActive(ctx, userID, true)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deactivate user - success", func(t *testing.T) {
		mockRepo.On("UpdateIsActive", ctx, userID, false).Return(nil).Once()

		err := mockRepo.UpdateIsActive(ctx, userID, false)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("activate user - error", func(t *testing.T) {
		expectedErr := errors.New("update failed")
		mockRepo.On("UpdateIsActive", ctx, userID, true).Return(expectedErr).Once()

		err := mockRepo.UpdateIsActive(ctx, userID, true)

		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deactivate user - error", func(t *testing.T) {
		expectedErr := errors.New("update failed")
		mockRepo.On("UpdateIsActive", ctx, userID, false).Return(expectedErr).Once()

		err := mockRepo.UpdateIsActive(ctx, userID, false)

		assert.ErrorIs(t, err, expectedErr)
		mockRepo.AssertExpectations(t)
	})
}

func TestMockUserRepositoryGetByTeam(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()
	teamName := "backend"

	teamUsers := []*domain.User{
		{
			UserID:   "user-1",
			Username: "john_doe",
			TeamName: teamName,
			IsActive: true,
		},
		{
			UserID:   "user-2",
			Username: "jane_smith",
			TeamName: teamName,
			IsActive: false,
		},
		{
			UserID:   "user-3",
			Username: "bob_wilson",
			TeamName: teamName,
			IsActive: true,
		},
	}

	t.Run("returns all team users", func(t *testing.T) {
		mockRepo.On("GetByTeam", ctx, teamName).Return(teamUsers, nil).Once()

		users, err := mockRepo.GetByTeam(ctx, teamName)

		require.NoError(t, err)
		assert.Equal(t, teamUsers, users)
		assert.Len(t, users, 3)

		assert.True(t, users[0].IsActive)
		assert.False(t, users[1].IsActive)
		assert.True(t, users[2].IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no users in team", func(t *testing.T) {
		mockRepo.On("GetByTeam", ctx, teamName).Return([]*domain.User{}, nil).Once()

		users, err := mockRepo.GetByTeam(ctx, teamName)

		require.NoError(t, err)
		assert.Empty(t, users)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.On("GetByTeam", ctx, teamName).Return(([]*domain.User)(nil), errors.New("query failed")).Once()

		users, err := mockRepo.GetByTeam(ctx, teamName)

		assert.Nil(t, users)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil slice with error - covers nil branch", func(t *testing.T) {
		nonExistentTeam := "non-existent"
		mockRepo.On("GetByTeam", ctx, nonExistentTeam).Return(nil, errors.New("team not found")).Once()

		users, err := mockRepo.GetByTeam(ctx, nonExistentTeam)

		assert.Nil(t, users)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "team not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestMockUserRepositoryEdgeCases(t *testing.T) {
	mockRepo := new(MockUserRepository)
	ctx := context.Background()

	t.Run("empty user ID", func(t *testing.T) {
		emptyUserID := ""
		mockRepo.On("GetByID", ctx, emptyUserID).Return((*domain.User)(nil), errors.New("empty user ID")).Once()

		user, err := mockRepo.GetByID(ctx, emptyUserID)

		assert.Nil(t, user)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty team name for GetByTeam", func(t *testing.T) {
		emptyTeam := ""
		mockRepo.On("GetByTeam", ctx, emptyTeam).Return([]*domain.User{}, nil).Once()

		users, err := mockRepo.GetByTeam(ctx, emptyTeam)

		require.NoError(t, err)
		assert.Empty(t, users)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty team name for GetActiveByTeam", func(t *testing.T) {
		emptyTeam := ""
		mockRepo.On("GetActiveByTeam", ctx, emptyTeam).Return([]*domain.User{}, nil).Once()

		users, err := mockRepo.GetActiveByTeam(ctx, emptyTeam)

		require.NoError(t, err)
		assert.Empty(t, users)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user without team", func(t *testing.T) {
		userWithoutTeam := &domain.User{
			UserID:   "user-99",
			Username: "no_team_user",
			TeamName: "",
			IsActive: true,
		}
		mockRepo.On("GetByID", ctx, "user-99").Return(userWithoutTeam, nil).Once()

		user, err := mockRepo.GetByID(ctx, "user-99")

		require.NoError(t, err)
		assert.Equal(t, "user-99", user.UserID)
		assert.Empty(t, user.TeamName)
		assert.True(t, user.IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update isActive for non-existent user", func(t *testing.T) {
		nonExistentUser := "non-existent"
		mockRepo.On("UpdateIsActive", ctx, nonExistentUser, true).Return(errors.New("user not found")).Once()

		err := mockRepo.UpdateIsActive(ctx, nonExistentUser, true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}
