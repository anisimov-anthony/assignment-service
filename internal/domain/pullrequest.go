package domain

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string     `bson:"pull_request_id" json:"pull_request_id"`
	PullRequestName   string     `bson:"pull_request_name" json:"pull_request_name"`
	AuthorID          string     `bson:"author_id" json:"author_id"`
	Status            PRStatus   `bson:"status" json:"status"`
	AssignedReviewers []string   `bson:"assigned_reviewers" json:"assigned_reviewers"`
	CreatedAt         *time.Time `bson:"created_at,omitempty" json:"createdAt,omitempty"`
	MergedAt          *time.Time `bson:"merged_at,omitempty" json:"mergedAt,omitempty"`
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == PRStatusMerged
}

type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          PRStatus `json:"status"`
}
