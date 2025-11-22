package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"assignment-service/internal/domain"
	"assignment-service/internal/http/dto"
	"assignment-service/internal/repository/mocks"
	"assignment-service/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestTeamHandlerCreateTeam(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful creation", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		teamService := service.NewTeamService(mockTeamRepo, mockUserRepo, logger)
		handler := NewTeamHandler(teamService, logger)

		mockTeamRepo.On("Exists", mock.Anything, "team-1").Return(false, nil)
		mockTeamRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Team")).Return(nil)
		mockUserRepo.On("CreateOrUpdate", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Times(2)

		reqBody := dto.CreateTeamRequest{
			TeamName: "team-1",
			Members: []dto.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
				{UserID: "user-2", Username: "user2", IsActive: true},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("team already exists", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		teamService := service.NewTeamService(mockTeamRepo, mockUserRepo, logger)
		handler := NewTeamHandler(teamService, logger)

		mockTeamRepo.On("Exists", mock.Anything, "team-1").Return(true, nil)

		reqBody := dto.CreateTeamRequest{
			TeamName: "team-1",
			Members:  []dto.TeamMember{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("missing team_name", func(t *testing.T) {
		handler := NewTeamHandler(nil, logger)

		reqBody := dto.CreateTeamRequest{
			TeamName: "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		handler := NewTeamHandler(nil, logger)

		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong HTTP method", func(t *testing.T) {
		handler := NewTeamHandler(nil, logger)

		req := httptest.NewRequest(http.MethodGet, "/team/add", nil)
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestTeamHandlerGetTeam(t *testing.T) {
	logger := zap.NewNop()

	t.Run("successful get team", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		teamService := service.NewTeamService(mockTeamRepo, mockUserRepo, logger)
		handler := NewTeamHandler(teamService, logger)

		team := &domain.Team{
			TeamName: "team-1",
			Members: []domain.TeamMember{
				{UserID: "user-1", Username: "user1", IsActive: true},
			},
		}

		mockTeamRepo.On("GetByName", mock.Anything, "team-1").Return(team, nil)

		req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=team-1", nil)
		w := httptest.NewRecorder()

		handler.GetTeam(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		mockTeamRepo := new(mocks.MockTeamRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		teamService := service.NewTeamService(mockTeamRepo, mockUserRepo, logger)
		handler := NewTeamHandler(teamService, logger)

		mockTeamRepo.On("GetByName", mock.Anything, "team-1").Return(nil, domain.ErrTeamNotFound)

		req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=team-1", nil)
		w := httptest.NewRecorder()

		handler.GetTeam(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("missing team_name", func(t *testing.T) {
		handler := NewTeamHandler(nil, logger)

		req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
		w := httptest.NewRecorder()

		handler.GetTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong HTTP method", func(t *testing.T) {
		handler := NewTeamHandler(nil, logger)

		req := httptest.NewRequest(http.MethodPost, "/team/get?team_name=team-1", nil)
		w := httptest.NewRecorder()

		handler.GetTeam(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
