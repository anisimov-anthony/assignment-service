package handlers

import (
	"encoding/json"
	"net/http"

	"assignment-service/internal/domain"
	"assignment-service/internal/http/dto"
	"assignment-service/internal/service"

	"go.uber.org/zap"
)

type PRHandler struct {
	prService *service.PRService
	logger    *zap.Logger
}

func NewPRHandler(prService *service.PRService, logger *zap.Logger) *PRHandler {
	return &PRHandler{
		prService: prService,
		logger:    logger,
	}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, domain.ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "all fields are required", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch err {
		case domain.ErrPRExists:
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusConflict)
			return
		case domain.ErrUserNotFound:
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusNotFound)
			return
		default:
			h.logger.Error("failed to create PR", zap.Error(err))
			h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.PRResponse{PR: *pr})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, domain.ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PullRequestID == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "pull_request_id is required", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		if err == domain.ErrPRNotFound {
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusNotFound)
			return
		}

		h.logger.Error("failed to merge PR", zap.Error(err))
		h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.PRResponse{PR: *pr})
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, domain.ErrorCodeNotFound, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PullRequestID == "" || req.OldReviewerID == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "pull_request_id and old_user_id are required", http.StatusBadRequest)
		return
	}

	pr, newUserID, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		switch err {
		case domain.ErrPRNotFound, domain.ErrUserNotFound:
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusNotFound)
			return
		case domain.ErrPRMerged, domain.ErrNotAssigned, domain.ErrNoCandidate:
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusConflict)
			return
		default:
			h.logger.Error("failed to reassign reviewer", zap.Error(err))
			h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.ReassignResponse{
		PR:         *pr,
		ReplacedBy: newUserID,
	})
}

func (h *PRHandler) sendError(w http.ResponseWriter, code domain.ErrorCode, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    string(code),
			Message: message,
		},
	})
}
