package service

import (
	"context"
	"testing"
	"time"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestPRServiceCreatePR(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		author := &domain.User{
			UserID:   "user-1",
			Username: "author",
			TeamName: "team-1",
			IsActive: true,
		}

		teamMembers := []*domain.User{
			{UserID: "user-2", Username: "reviewer1", TeamName: "team-1", IsActive: true},
			{UserID: "user-3", Username: "reviewer2", TeamName: "team-1", IsActive: true},
			{UserID: "user-4", Username: "reviewer3", TeamName: "team-1", IsActive: true},
		}

		mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
		mockUserRepo.On("GetByID", ctx, "user-1").Return(author, nil)
		mockUserRepo.On("GetActiveByTeam", ctx, "team-1").Return(teamMembers, nil)
		mockPRRepo.On("Create", ctx, mock.AnythingOfType("*domain.PullRequest")).Return(nil)

		pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Equal(t, "pr-1", pr.PullRequestID)
		assert.Equal(t, "Test PR", pr.PullRequestName)
		assert.Equal(t, "user-1", pr.AuthorID)
		assert.Equal(t, domain.PRStatusOpen, pr.Status)
		assert.Len(t, pr.AssignedReviewers, 2)
		assert.NotNil(t, pr.CreatedAt)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("PR already exists", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		mockPRRepo.On("Exists", ctx, "pr-1").Return(true, nil)

		pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "user-1")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPRExists, err)
		assert.Nil(t, pr)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
		mockUserRepo.On("GetByID", ctx, "user-1").Return(nil, domain.ErrUserNotFound)

		pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "user-1")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, pr)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("no candidates for review", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		author := &domain.User{
			UserID:   "user-1",
			Username: "author",
			TeamName: "team-1",
			IsActive: true,
		}

		mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
		mockUserRepo.On("GetByID", ctx, "user-1").Return(author, nil)
		mockUserRepo.On("GetActiveByTeam", ctx, "team-1").Return([]*domain.User{}, nil)
		mockPRRepo.On("Create", ctx, mock.AnythingOfType("*domain.PullRequest")).Return(nil)

		pr, err := service.CreatePR(ctx, "pr-1", "Test PR", "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.Empty(t, pr.AssignedReviewers)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestPRServiceMergePR(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful merge", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
		mockPRRepo.On("Update", ctx, mock.AnythingOfType("*domain.PullRequest")).Return(nil).Run(func(args mock.Arguments) {
			updatedPR := args.Get(1).(*domain.PullRequest)
			assert.Equal(t, domain.PRStatusMerged, updatedPR.Status)
			assert.NotNil(t, updatedPR.MergedAt)
		})

		result, err := service.MergePR(ctx, "pr-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.PRStatusMerged, result.Status)
		assert.NotNil(t, result.MergedAt)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("already merged PR", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		now := time.Now()
		mergedAt := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusMerged,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
			MergedAt:          &mergedAt,
		}

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

		result, err := service.MergePR(ctx, "pr-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.PRStatusMerged, result.Status)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(nil, domain.ErrPRNotFound)

		result, err := service.MergePR(ctx, "pr-1")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPRNotFound, err)
		assert.Nil(t, result)
		mockPRRepo.AssertExpectations(t)
	})
}

func TestPRServiceReassignReviewer(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful reassignment", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		oldReviewer := &domain.User{
			UserID:   "user-2",
			Username: "reviewer",
			TeamName: "team-1",
			IsActive: true,
		}

		newReviewer := &domain.User{
			UserID:   "user-3",
			Username: "newreviewer",
			TeamName: "team-1",
			IsActive: true,
		}

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
		mockUserRepo.On("GetByID", ctx, "user-2").Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveByTeam", ctx, "team-1").Return([]*domain.User{newReviewer}, nil)
		mockPRRepo.On("Update", ctx, mock.AnythingOfType("*domain.PullRequest")).Return(nil)

		result, newUserID, err := service.ReassignReviewer(ctx, "pr-1", "user-2")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user-3", newUserID)
		assert.Contains(t, result.AssignedReviewers, "user-3")
		assert.NotContains(t, result.AssignedReviewers, "user-2")
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(nil, domain.ErrPRNotFound)

		result, newUserID, err := service.ReassignReviewer(ctx, "pr-1", "user-2")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPRNotFound, err)
		assert.Nil(t, result)
		assert.Empty(t, newUserID)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("PR already merged", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		now := time.Now()
		mergedAt := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusMerged,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
			MergedAt:          &mergedAt,
		}

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

		result, newUserID, err := service.ReassignReviewer(ctx, "pr-1", "user-2")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrPRMerged, err)
		assert.Nil(t, result)
		assert.Empty(t, newUserID)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("reviewer not assigned", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-3"}, // user-2 is not assigned
			CreatedAt:         &now,
		}

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

		result, newUserID, err := service.ReassignReviewer(ctx, "pr-1", "user-2")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotAssigned, err)
		assert.Nil(t, result)
		assert.Empty(t, newUserID)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("no candidate", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		oldReviewer := &domain.User{
			UserID:   "user-2",
			Username: "reviewer",
			TeamName: "team-1",
			IsActive: true,
		}

		mockPRRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
		mockUserRepo.On("GetByID", ctx, "user-2").Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveByTeam", ctx, "team-1").Return([]*domain.User{}, nil)

		result, newUserID, err := service.ReassignReviewer(ctx, "pr-1", "user-2")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNoCandidate, err)
		assert.Nil(t, result)
		assert.Empty(t, newUserID)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestPRServiceGetPRsByReviewer(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful get PRs", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		user := &domain.User{
			UserID:   "user-1",
			Username: "reviewer",
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
		}

		mockUserRepo.On("GetByID", ctx, "user-1").Return(user, nil)
		mockPRRepo.On("GetByReviewer", ctx, "user-1").Return(prs, nil)

		result, err := service.GetPRsByReviewer(ctx, "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		mockUserRepo.On("GetByID", ctx, "user-1").Return(nil, domain.ErrUserNotFound)

		result, err := service.GetPRsByReviewer(ctx, "user-1")

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestPRServiceReassignOpenPRsForTeam(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	t.Run("successful reassignment", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		service := NewPRService(mockPRRepo, mockUserRepo, logger)

		users := []*domain.User{
			{UserID: "user-1", Username: "user1", TeamName: "team-1", IsActive: true},
			{UserID: "user-2", Username: "user2", TeamName: "team-1", IsActive: false},
		}

		now := time.Now()
		prs := []*domain.PullRequest{
			{
				PullRequestID:     "pr-1",
				PullRequestName:   "PR 1",
				AuthorID:          "user-1",
				Status:            domain.PRStatusOpen,
				AssignedReviewers: []string{"user-2"},
				CreatedAt:         &now,
			},
		}

		newReviewers := []*domain.User{
			{UserID: "user-3", Username: "user3", TeamName: "team-1", IsActive: true},
		}

		mockUserRepo.On("GetByTeam", ctx, "team-1").Return(users, nil)
		mockPRRepo.On("GetByReviewer", ctx, "user-1").Return(prs, nil)
		mockPRRepo.On("GetByReviewer", ctx, "user-2").Return([]*domain.PullRequest{}, nil)
		mockUserRepo.On("GetByID", ctx, "user-1").Return(users[0], nil)
		mockUserRepo.On("GetActiveByTeam", ctx, "team-1").Return(newReviewers, nil)
		mockPRRepo.On("Update", ctx, mock.AnythingOfType("*domain.PullRequest")).Return(nil)

		err := service.ReassignOpenPRsForTeam(ctx, "team-1")

		assert.NoError(t, err)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestPRServiceSelectReviewers(t *testing.T) {
	logger := zap.NewNop()
	service := NewPRService(nil, nil, logger)

	t.Run("select from multiple candidates", func(t *testing.T) {
		candidates := []*domain.User{
			{UserID: "user-1"},
			{UserID: "user-2"},
			{UserID: "user-3"},
			{UserID: "user-4"},
		}

		reviewers := service.selectReviewers(candidates, 2)

		assert.Len(t, reviewers, 2)
		for _, reviewer := range reviewers {
			assert.Contains(t, []string{"user-1", "user-2", "user-3", "user-4"}, reviewer)
		}
	})

	t.Run("select all when candidates less than max", func(t *testing.T) {
		candidates := []*domain.User{
			{UserID: "user-1"},
		}

		reviewers := service.selectReviewers(candidates, 2)

		assert.Len(t, reviewers, 1)
		assert.Equal(t, "user-1", reviewers[0])
	})

	t.Run("empty candidates", func(t *testing.T) {
		reviewers := service.selectReviewers([]*domain.User{}, 2)

		assert.Empty(t, reviewers)
	})
}

func TestPRServiceIsReviewerAssigned(t *testing.T) {
	logger := zap.NewNop()
	service := NewPRService(nil, nil, logger)

	t.Run("reviewer is assigned", func(t *testing.T) {
		reviewers := []string{"user-1", "user-2", "user-3"}

		result := service.isReviewerAssigned(reviewers, "user-2")

		assert.True(t, result)
	})

	t.Run("reviewer is not assigned", func(t *testing.T) {
		reviewers := []string{"user-1", "user-2", "user-3"}

		result := service.isReviewerAssigned(reviewers, "user-4")

		assert.False(t, result)
	})

	t.Run("empty reviewers list", func(t *testing.T) {
		result := service.isReviewerAssigned([]string{}, "user-1")

		assert.False(t, result)
	})
}

func TestPRServiceCreatePRCreateError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("error when creating PR in repo", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		svc := NewPRService(mockPRRepo, mockUserRepo, logger)

		author := &domain.User{UserID: "author-1", TeamName: "team-1"}
		teamMembers := []*domain.User{
			{UserID: "reviewer-1", TeamName: "team-1"},
			{UserID: "reviewer-2", TeamName: "team-1"},
		}

		mockPRRepo.On("Exists", ctx, "pr-1").Return(false, nil)
		mockUserRepo.On("GetByID", ctx, "author-1").Return(author, nil)
		mockUserRepo.On("GetActiveByTeam", ctx, "team-1").Return(teamMembers, nil)

		createErr := assert.AnError
		mockPRRepo.On("Create", ctx, mock.Anything).Return(createErr)

		pr, err := svc.CreatePR(ctx, "pr-1", "Fix bug", "author-1")

		assert.Error(t, err)
		assert.Equal(t, createErr, err)
		assert.Nil(t, pr)

		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
