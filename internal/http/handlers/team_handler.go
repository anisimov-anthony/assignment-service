package handlers

import (
	"encoding/json"
	"net/http"

	"assignment-service/internal/domain"
	"assignment-service/internal/http/dto"
	"assignment-service/internal/service"

	"go.uber.org/zap"
)

type TeamHandler struct {
	teamService *service.TeamService
	logger      *zap.Logger
}

func NewTeamHandler(teamService *service.TeamService, logger *zap.Logger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		logger:      logger,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, domain.ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TeamName == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "team_name is required", http.StatusBadRequest)
		return
	}

	members := make([]domain.TeamMember, len(req.Members))
	for i, m := range req.Members {
		members[i] = domain.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	team := &domain.Team{
		TeamName: req.TeamName,
		Members:  members,
	}

	if err := h.teamService.CreateTeam(r.Context(), team); err != nil {
		if err == domain.ErrTeamExists {
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusBadRequest)
			return
		}

		h.logger.Error("failed to create team", zap.Error(err))
		h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.TeamResponse{Team: *team})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		if err == domain.ErrTeamNotFound {
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusNotFound)
			return
		}

		h.logger.Error("failed to get team", zap.Error(err))
		h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(team)
}

func (h *TeamHandler) sendError(w http.ResponseWriter, code domain.ErrorCode, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    string(code),
			Message: message,
		},
	})
}
