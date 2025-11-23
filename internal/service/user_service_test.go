package service

import (
	"context"
	"testing"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUserServiceSetIsActive(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful set is_active", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewUserService(mockUserRepo, logger)

		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: false,
		}

		updatedUser := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: true,
		}

		mockUserRepo.On("GetByID", ctx, "user-1").Return(user, nil).Once()
		mockUserRepo.On("UpdateIsActive", ctx, "user-1", true).Return(nil)
		mockUserRepo.On("GetByID", ctx, "user-1").Return(updatedUser, nil).Once()

		result, err := service.SetIsActive(ctx, "user-1", true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsActive)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewUserService(mockUserRepo, logger)

		mockUserRepo.On("GetByID", ctx, "user-1").Return(nil, domain.ErrUserNotFound)

		result, err := service.SetIsActive(ctx, "user-1", true)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserServiceGetUserByID(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful get user", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewUserService(mockUserRepo, logger)

		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: true,
		}

		mockUserRepo.On("GetByID", ctx, "user-1").Return(user, nil)

		result, err := service.GetUserByID(ctx, "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user-1", result.UserID)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewUserService(mockUserRepo, logger)

		mockUserRepo.On("GetByID", ctx, "user-1").Return(nil, domain.ErrUserNotFound)

		result, err := service.GetUserByID(ctx, "user-1")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserServiceSetIsActiveUpdateError(t *testing.T) {
	ctx := context.Background()

	t.Run("error on UpdateIsActive", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		logger := zap.NewNop()

		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: false,
		}

		mockRepo.On("GetByID", ctx, "user-1").Return(user, nil)
		mockRepo.On("UpdateIsActive", ctx, "user-1", true).Return(assert.AnError)

		svc := NewUserService(mockRepo, logger)

		result, err := svc.SetIsActive(ctx, "user-1", true)

		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		assert.Nil(t, result)
		assert.False(t, user.IsActive)
		mockRepo.AssertExpectations(t)
	})
}
