package domain

import "testing"

func TestToErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code ErrorCode
	}{
		{"team exists", ErrTeamExists, ErrorCodeTeamExists},
		{"pr exists", ErrPRExists, ErrorCodePRExists},
		{"pr merged", ErrPRMerged, ErrorCodePRMerged},
		{"not assigned", ErrNotAssigned, ErrorCodeNotAssigned},
		{"no candidate", ErrNoCandidate, ErrorCodeNoCandidate},
		{"not found generic", ErrNotFound, ErrorCodeNotFound},
		{"user not found", ErrUserNotFound, ErrorCodeNotFound},
		{"team not found", ErrTeamNotFound, ErrorCodeNotFound},
		{"pr not found", ErrPRNotFound, ErrorCodeNotFound},
		{"unknown error maps to not found", errUnknown{}, ErrorCodeNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToErrorCode(tt.err)
			if got != tt.code {
				t.Fatalf("expected %v, got %v", tt.code, got)
			}
		})
	}
}

type errUnknown struct{}

func (errUnknown) Error() string { return "unknown" }
