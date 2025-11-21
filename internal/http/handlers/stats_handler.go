package handlers

import (
	"encoding/json"
	"net/http"

	"assignment-service/internal/domain"
	"assignment-service/internal/http/dto"
	"assignment-service/internal/service"

	"go.uber.org/zap"
)

type StatsHandler struct {
	statsService *service.StatsService
	logger       *zap.Logger
}

func NewStatsHandler(statsService *service.StatsService, logger *zap.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		logger:       logger,
	}
}

func (h *StatsHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.sendError(w, domain.ErrorCodeNotFound, "user_id is required", http.StatusBadRequest)
		return
	}

	stats, err := h.statsService.GetUserStats(r.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			h.sendError(w, domain.ToErrorCode(err), err.Error(), http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get user stats", zap.Error(err))
		h.sendError(w, domain.ErrorCodeNotFound, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (h *StatsHandler) sendError(w http.ResponseWriter, code domain.ErrorCode, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    string(code),
			Message: message,
		},
	})
}
