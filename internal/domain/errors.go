package domain

import "errors"

// Domain error codes
var (
	ErrTeamExists   = errors.New("team_name already exists")
	ErrPRExists     = errors.New("PR id already exists")
	ErrPRMerged     = errors.New("cannot reassign on merged PR")
	ErrNotAssigned  = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate  = errors.New("no active replacement candidate in team")
	ErrNotFound     = errors.New("resource not found")
	ErrUserNotFound = errors.New("user not found")
	ErrTeamNotFound = errors.New("team not found")
	ErrPRNotFound   = errors.New("PR not found")
)

type ErrorCode string

// API error codes
// ErrorResponse:error:code:enum
const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

// domain error code -> API error code
func ToErrorCode(err error) ErrorCode {
	switch err {
	case ErrTeamExists:
		return ErrorCodeTeamExists
	case ErrPRExists:
		return ErrorCodePRExists
	case ErrPRMerged:
		return ErrorCodePRMerged
	case ErrNotAssigned:
		return ErrorCodeNotAssigned
	case ErrNoCandidate:
		return ErrorCodeNoCandidate
	case ErrNotFound, ErrUserNotFound, ErrTeamNotFound, ErrPRNotFound:
		return ErrorCodeNotFound
	default:
		return ErrorCodeNotFound
	}
}
