package service

import (
	"context"
	"fmt"
	"testing"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestTeamServiceCreateTeam(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
				{UserID: "user-2", Username: "user2", IsActive: true},
			},
		}

		mockTeamRepo.On("Exists", ctx, "team-1").Return(false, nil)
		mockTeamRepo.On("Create", ctx, team).Return(nil)
		mockUserRepo.On("CreateOrUpdate", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Times(2)

		err := service.CreateTeam(ctx, team)

		assert.NoError(t, err)
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("team already exists", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members:  []domain.TeamMember{},
		}

		mockTeamRepo.On("Exists", ctx, "team-1").Return(true, nil)

		err := service.CreateTeam(ctx, team)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTeamExists, err)
		mockTeamRepo.AssertExpectations(t)
	})
}

func TestTeamServiceGetTeam(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful get team", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		users := []*domain.User{
			{UserID: "user-1", Username: "user1", TeamName: "team-1", IsActive: false}, // actual db data
		}

		mockTeamRepo.On("GetByName", ctx, "team-1").Return(team, nil)
		mockUserRepo.On("GetByTeam", ctx, "team-1").Return(users, nil)

		result, err := service.GetTeam(ctx, "team-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "team-1", result.TeamName)
		require.Len(t, result.Members, 1)
		assert.Equal(t, "user-1", result.Members[0].UserID)
		assert.Equal(t, "user1", result.Members[0].Username)
		assert.Equal(t, false, result.Members[0].IsActive) // must be actual status
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewTeamService(mockTeamRepo, mockUserRepo, logger)

		mockTeamRepo.On("GetByName", ctx, "team-1").Return(nil, domain.ErrTeamNotFound)

		result, err := service.GetTeam(ctx, "team-1")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrTeamNotFound, err)
		assert.Nil(t, result)
		mockTeamRepo.AssertExpectations(t)
	})
}

func TestTeamServiceCreateTeamErrorCoverage(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("error on Exists", func(t *testing.T) {
		teamRepo := new(mocks.MockTeamRepository)
		userRepo := new(mocks.MockUserRepository)

		teamRepo.On("Exists", ctx, "team-1").Return(false, fmt.Errorf("db down"))

		svc := NewTeamService(teamRepo, userRepo, logger)

		err := svc.CreateTeam(ctx, &domain.Team{TeamName: "team-1"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check team existence")
		teamRepo.AssertExpectations(t)
		teamRepo.AssertNotCalled(t, "Create")
		userRepo.AssertNotCalled(t, "CreateOrUpdate")
	})

	t.Run("error on Create team", func(t *testing.T) {
		teamRepo := new(mocks.MockTeamRepository)
		userRepo := new(mocks.MockUserRepository)

		teamRepo.On("Exists", ctx, "team-1").Return(false, nil)
		teamRepo.On("Create", ctx, mock.Anything).Return(fmt.Errorf("insert failed"))

		svc := NewTeamService(teamRepo, userRepo, logger)

		err := svc.CreateTeam(ctx, &domain.Team{TeamName: "team-1"})

		assert.Error(t, err)
		assert.Equal(t, "insert failed", err.Error())
		teamRepo.AssertExpectations(t)
		userRepo.AssertNotCalled(t, "CreateOrUpdate")
	})

	t.Run("error on CreateOrUpdate user + logging", func(t *testing.T) {
		teamRepo := new(mocks.MockTeamRepository)
		userRepo := new(mocks.MockUserRepository)

		mockLogger := zaptest.NewLogger(t)

		team := &domain.Team{
			TeamName: "team-1",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "u1", IsActive: true},
				{UserID: "user-2", Username: "u2", IsActive: false},
			},
		}

		teamRepo.On("Exists", ctx, "team-1").Return(false, nil)
		teamRepo.On("Create", ctx, team).Return(nil)

		userRepo.On("CreateOrUpdate", ctx, mock.MatchedBy(func(u *domain.User) bool {
			return u.UserID == "user-1"
		})).Return(nil).Once()

		userRepo.On("CreateOrUpdate", ctx, mock.MatchedBy(func(u *domain.User) bool {
			return u.UserID == "user-2"
		})).Return(fmt.Errorf("unique violation")).Once()

		svc := NewTeamService(teamRepo, userRepo, mockLogger)

		err := svc.CreateTeam(ctx, team)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create or update user user-2")
		assert.Contains(t, err.Error(), "unique violation")
	})

	t.Run("success with empty members", func(t *testing.T) {
		teamRepo := new(mocks.MockTeamRepository)
		userRepo := new(mocks.MockUserRepository)

		team := &domain.Team{TeamName: "team-empty"}

		teamRepo.On("Exists", ctx, "team-empty").Return(false, nil)
		teamRepo.On("Create", ctx, team).Return(nil)

		svc := NewTeamService(teamRepo, userRepo, logger)

		err := svc.CreateTeam(ctx, team)

		assert.NoError(t, err)
		teamRepo.AssertExpectations(t)
		userRepo.AssertNotCalled(t, "CreateOrUpdate")
	})
}
