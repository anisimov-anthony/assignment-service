package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"assignment-service/internal/domain"
	"assignment-service/internal/http/dto"
	"assignment-service/internal/repository/mocks"
	"assignment-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUserHandlerSetIsActive(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful set is_active", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockUserRepo, logger)
		mockPRRepo := new(mocks.MockPRRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewUserHandler(userService, prService, logger)

		user := &domain.User{
			UserID:   "user-1",
			Username: "testuser",
			TeamName: "team-1",
			IsActive: false,
		}

		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
		mockUserRepo.On("UpdateIsActive", mock.Anything, "user-1", true).Return(nil)

		reqBody := map[string]any{
			"user_id":   "user-1",
			"is_active": true,
		}
		body := createJSONBody(t, reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", body)
		w := httptest.NewRecorder()

		handler.SetIsActive(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockUserRepo, logger)
		mockPRRepo := new(mocks.MockPRRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewUserHandler(userService, prService, logger)

		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(nil, domain.ErrUserNotFound)

		reqBody := map[string]any{
			"user_id":   "user-1",
			"is_active": true,
		}
		body := createJSONBody(t, reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", body)
		w := httptest.NewRecorder()

		handler.SetIsActive(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		handler := NewUserHandler(nil, nil, logger)

		reqBody := map[string]any{
			"is_active": true,
		}
		body := createJSONBody(t, reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", body)
		w := httptest.NewRecorder()

		handler.SetIsActive(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong HTTP method", func(t *testing.T) {
		handler := NewUserHandler(nil, nil, logger)

		req := httptest.NewRequest(http.MethodGet, "/users/setIsActive", nil)
		w := httptest.NewRecorder()

		handler.SetIsActive(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestUserHandlerGetReview(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful get review", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockPRRepo := new(mocks.MockPRRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		userService := service.NewUserService(mockUserRepo, logger)
		handler := NewUserHandler(userService, prService, logger)

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

		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
		mockPRRepo.On("GetByReviewer", mock.Anything, "user-1").Return(prs, nil)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=user-1", nil)
		w := httptest.NewRecorder()

		handler.GetReview(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response dto.GetReviewResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "user-1", response.UserID)
		require.Len(t, response.PullRequests, 1)
		assert.Equal(t, "pr-1", response.PullRequests[0].PullRequestID)
		assert.Equal(t, "PR 1", response.PullRequests[0].PullRequestName)
		assert.Equal(t, "author-1", response.PullRequests[0].AuthorID)
		assert.Equal(t, domain.PRStatusOpen, response.PullRequests[0].Status)

		mockUserRepo.AssertExpectations(t)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo := new(mocks.MockUserRepository)
		mockPRRepo := new(mocks.MockPRRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		userService := service.NewUserService(mockUserRepo, logger)
		handler := NewUserHandler(userService, prService, logger)

		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(nil, domain.ErrUserNotFound)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=user-1", nil)
		w := httptest.NewRecorder()

		handler.GetReview(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		handler := NewUserHandler(nil, nil, logger)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
		w := httptest.NewRecorder()

		handler.GetReview(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong HTTP method", func(t *testing.T) {
		handler := NewUserHandler(nil, nil, logger)

		req := httptest.NewRequest(http.MethodPost, "/users/getReview?user_id=user-1", nil)
		w := httptest.NewRecorder()

		handler.GetReview(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func createJSONBody(t *testing.T, data any) *bytes.Buffer {
	t.Helper()
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(data)
	if err != nil {
		t.Fatalf("failed to encode JSON: %v", err)
	}
	return body
}
