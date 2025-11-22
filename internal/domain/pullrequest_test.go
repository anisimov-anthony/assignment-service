package domain

import (
	"testing"
)

func TestPullRequest_IsMerged(t *testing.T) {
	tests := []struct {
		name     string
		pr       PullRequest
		expected bool
	}{
		{
			name: "merged PR",
			pr: PullRequest{
				Status: PRStatusMerged,
			},
			expected: true,
		},
		{
			name: "open PR",
			pr: PullRequest{
				Status: PRStatusOpen,
			},
			expected: false,
		},
		{
			name: "empty status",
			pr: PullRequest{
				Status: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pr.IsMerged()
			if result != tt.expected {
				t.Errorf("IsMerged() = %v, want %v", result, tt.expected)
			}
		})
	}
}
