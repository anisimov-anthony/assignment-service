package handlers

import (
	"encoding/json"
	"net/http"

	"assignment-service/internal/domain"
	"assignment-service/internal/http/dto"
	"assignment-service/internal/service"

	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
	prService   *service.PRService
	logger      *zap.Logger
}

func NewUserHandler(userService *service.UserService, prService *service.PRService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
		logger:      logger,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, domain.ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "user_id is required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if err == domain.ErrUserNotFound {
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error("failed to set user is_active", zap.Error(err))
		h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.UserResponse{User: *user})
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "user_id is required", http.StatusBadRequest)
		return
	}

	prs, err := h.prService.GetPRsByReviewer(r.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get PRs by reviewer", zap.Error(err))
		h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
		return
	}

	shortPRs := make([]domain.PullRequest, 0, len(prs))
	for _, pr := range prs {
		shortPRs = append(shortPRs, domain.PullRequest{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.GetReviewResponse{
		UserID:       userID,
		PullRequests: shortPRs,
	})
}

func (h *UserHandler) sendError(w http.ResponseWriter, code domain.ErrorCode, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    string(code),
			Message: message,
		},
	})
}
