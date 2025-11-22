package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository/mocks"
	"assignment-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestStatsHandlerGetUserStats(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful get stats", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		statsService := service.NewStatsService(mockPRRepo, mockUserRepo, logger)
		handler := NewStatsHandler(statsService, logger)

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
		}

		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
		mockPRRepo.On("GetByReviewer", mock.Anything, "user-1").Return(prs, nil)

		req := httptest.NewRequest(http.MethodGet, "/stats/user?user_id=user-1", nil)
		w := httptest.NewRecorder()

		handler.GetUserStats(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		statsService := service.NewStatsService(mockPRRepo, mockUserRepo, logger)
		handler := NewStatsHandler(statsService, logger)

		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(nil, domain.ErrUserNotFound)

		req := httptest.NewRequest(http.MethodGet, "/stats/user?user_id=user-1", nil)
		w := httptest.NewRecorder()

		handler.GetUserStats(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		handler := NewStatsHandler(nil, logger)

		req := httptest.NewRequest(http.MethodGet, "/stats/user", nil)
		w := httptest.NewRecorder()

		handler.GetUserStats(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong HTTP method", func(t *testing.T) {
		handler := NewStatsHandler(nil, logger)

		req := httptest.NewRequest(http.MethodPost, "/stats/user?user_id=user-1", nil)
		w := httptest.NewRecorder()

		handler.GetUserStats(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
