package service

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository"

	"go.uber.org/zap"
)

type PRService struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewPRService(
	prRepo repository.PRRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return nil, domain.ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	teamMembers, err := s.userRepo.GetActiveByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	var candidates []*domain.User
	for _, member := range teamMembers {
		if member.UserID != authorID {
			candidates = append(candidates, member)
		}
	}

	reviewers := s.selectReviewers(candidates, 2)

	now := time.Now()
	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         &now,
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	// if already MERGED then return current state (for idempotency)
	if pr.IsMerged() {
		return pr, nil
	}

	now := time.Now()
	pr.Status = domain.PRStatusMerged
	pr.MergedAt = &now

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.IsMerged() {
		return nil, "", domain.ErrPRMerged
	}

	if !s.isReviewerAssigned(pr.AssignedReviewers, oldReviewerID) {
		return nil, "", domain.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		return nil, "", domain.ErrUserNotFound
	}

	teamMembers, err := s.userRepo.GetActiveByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get team members: %w", err)
	}

	var candidates []*domain.User
	for _, member := range teamMembers {
		if member.UserID != pr.AuthorID &&
			member.UserID != oldReviewerID &&
			!s.isReviewerAssigned(pr.AssignedReviewers, member.UserID) {
			candidates = append(candidates, member)
		}
	}

	if len(candidates) == 0 {
		return nil, "", domain.ErrNoCandidate
	}

	newReviewer := s.selectRandomCandidate(candidates)

	for i, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewer.UserID
			break
		}
	}

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, "", err
	}

	return pr, newReviewer.UserID, nil
}

func (s *PRService) GetPRsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (s *PRService) ReassignOpenPRsForTeam(ctx context.Context, teamName string) error {
	users, err := s.userRepo.GetByTeam(ctx, teamName)
	if err != nil {
		return fmt.Errorf("failed to get team members: %w", err)
	}

	teamUserIDs := make(map[string]bool)
	for _, user := range users {
		teamUserIDs[user.UserID] = true
	}

	var prsToReassign []*domain.PullRequest

	for _, user := range users {
		prs, err := s.prRepo.GetByReviewer(ctx, user.UserID)
		if err != nil {
			continue
		}

		for _, pr := range prs {
			if pr.Status == domain.PRStatusOpen {
				if teamUserIDs[pr.AuthorID] {
					prsToReassign = append(prsToReassign, pr)
				}
			}
		}
	}

	uniquePRs := make(map[string]*domain.PullRequest)
	for _, pr := range prsToReassign {
		uniquePRs[pr.PullRequestID] = pr
	}

	for _, pr := range uniquePRs {
		author, err := s.userRepo.GetByID(ctx, pr.AuthorID)
		if err != nil {
			continue
		}

		teamMembers, err := s.userRepo.GetActiveByTeam(ctx, author.TeamName)
		if err != nil {
			continue
		}

		var candidates []*domain.User
		for _, member := range teamMembers {
			if member.UserID != pr.AuthorID &&
				!s.isReviewerAssigned(pr.AssignedReviewers, member.UserID) {
				candidates = append(candidates, member)
			}
		}

		newReviewers := s.selectReviewers(candidates, 2)
		pr.AssignedReviewers = newReviewers

		if err := s.prRepo.Update(ctx, pr); err != nil {
			s.logger.Error("failed to update PR during reassignment",
				zap.Error(err),
				zap.String("pr_id", pr.PullRequestID))
		}
	}

	return nil
}

func (s *PRService) selectReviewers(candidates []*domain.User, maxReviewers int) []string {
	if len(candidates) == 0 {
		return nil
	}

	count := min(len(candidates), maxReviewers)

	shuffled := slices.Clone(candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	selected := shuffled[:count]
	reviewers := make([]string, 0, count)
	for _, u := range selected {
		reviewers = append(reviewers, u.UserID)
	}

	return reviewers
}

func (s *PRService) selectRandomCandidate(candidates []*domain.User) *domain.User {
	if len(candidates) == 0 {
		return nil
	}
	idx := rand.Intn(len(candidates))
	return candidates[idx]
}

func (s *PRService) isReviewerAssigned(reviewers []string, userID string) bool {
	return slices.Contains(reviewers, userID)
}
