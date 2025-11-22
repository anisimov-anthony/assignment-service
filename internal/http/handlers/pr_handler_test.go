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
	"go.uber.org/zap"
)

func TestPRHandlerCreatePR(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful creation", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

		author := &domain.User{
			UserID:   "user-1",
			Username: "author",
			TeamName: "team-1",
			IsActive: true,
		}

		teamMembers := []*domain.User{
			{UserID: "user-2", Username: "reviewer1", TeamName: "team-1", IsActive: true},
			{UserID: "user-3", Username: "reviewer2", TeamName: "team-1", IsActive: true},
		}

		mockPRRepo.On("Exists", mock.Anything, "pr-1").Return(false, nil)
		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(author, nil)
		mockUserRepo.On("GetActiveByTeam", mock.Anything, "team-1").Return(teamMembers, nil)
		mockPRRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.PullRequest")).Return(nil)

		reqBody := dto.CreatePRRequest{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        "user-1",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response dto.PRResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "pr-1", response.PR.PullRequestID)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("PR already exists", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

		mockPRRepo.On("Exists", mock.Anything, "pr-1").Return(true, nil)

		reqBody := dto.CreatePRRequest{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        "user-1",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

		mockPRRepo.On("Exists", mock.Anything, "pr-1").Return(false, nil)
		mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(nil, domain.ErrUserNotFound)

		reqBody := dto.CreatePRRequest{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        "user-1",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handler := NewPRHandler(nil, logger)

		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		handler := NewPRHandler(nil, logger)

		reqBody := dto.CreatePRRequest{
			PullRequestID: "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong HTTP method", func(t *testing.T) {
		handler := NewPRHandler(nil, logger)

		req := httptest.NewRequest(http.MethodGet, "/pullRequest/create", nil)
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestPRHandlerMergePR(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful merge", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-2"},
			CreatedAt:         &now,
		}

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(pr, nil).Once()
		mockPRRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.PullRequest")).Return(nil)

		reqBody := dto.MergePRRequest{
			PullRequestID: "pr-1",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.MergePR(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("already merged PR", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

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

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(pr, nil)

		reqBody := dto.MergePRRequest{
			PullRequestID: "pr-1",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.MergePR(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(nil, domain.ErrPRNotFound)

		reqBody := dto.MergePRRequest{
			PullRequestID: "pr-1",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.MergePR(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("missing pull_request_id", func(t *testing.T) {
		handler := NewPRHandler(nil, logger)

		reqBody := dto.MergePRRequest{
			PullRequestID: "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.MergePR(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPRHandlerReassignReviewer(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful reassignment", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

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

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(pr, nil)
		mockUserRepo.On("GetByID", mock.Anything, "user-2").Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveByTeam", mock.Anything, "team-1").Return([]*domain.User{newReviewer}, nil)
		mockPRRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.PullRequest")).Return(nil)

		reqBody := dto.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "user-2",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.ReassignResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "user-3", response.ReplacedBy)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("PR not found", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(nil, domain.ErrPRNotFound)

		reqBody := dto.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "user-2",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("PR already merged", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

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

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(pr, nil)

		reqBody := dto.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "user-2",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("reviewer not assigned", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

		now := time.Now()
		pr := &domain.PullRequest{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "user-1",
			Status:            domain.PRStatusOpen,
			AssignedReviewers: []string{"user-3"}, // user-2 is not assigned
			CreatedAt:         &now,
		}

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(pr, nil)

		reqBody := dto.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "user-2",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockPRRepo.AssertExpectations(t)
	})

	t.Run("no candidate", func(t *testing.T) {
		mockPRRepo := new(mocks.MockPRRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		prService := service.NewPRService(mockPRRepo, mockUserRepo, logger)
		handler := NewPRHandler(prService, logger)

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

		mockPRRepo.On("GetByID", mock.Anything, "pr-1").Return(pr, nil)
		mockUserRepo.On("GetByID", mock.Anything, "user-2").Return(oldReviewer, nil)
		mockUserRepo.On("GetActiveByTeam", mock.Anything, "team-1").Return([]*domain.User{}, nil) // no candidates

		reqBody := dto.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "user-2",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockPRRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("missing required fields", func(t *testing.T) {
		handler := NewPRHandler(nil, logger)

		reqBody := dto.ReassignReviewerRequest{
			PullRequestID: "",
			OldReviewerID: "user-2",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
