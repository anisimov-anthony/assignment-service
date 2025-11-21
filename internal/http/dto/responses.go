package dto

import "assignment-service/internal/domain"

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TeamResponse struct {
	Team domain.Team `json:"team"`
}

type UserResponse struct {
	User domain.User `json:"user"`
}

type PRResponse struct {
	PR domain.PullRequest `json:"pr"`
}

type ReassignResponse struct {
	PR         domain.PullRequest `json:"pr"`
	ReplacedBy string             `json:"replaced_by"`
}

type GetReviewResponse struct {
	UserID       string               `json:"user_id"`
	PullRequests []domain.PullRequest `json:"pull_requests"`
}
