package service

import (
	"context"
	"fmt"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository"

	"go.uber.org/zap"
)

type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewTeamService(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *domain.Team) error {
	exists, err := s.teamRepo.Exists(ctx, team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return domain.ErrTeamExists
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return err
	}

	for _, member := range team.Members {
		user := &domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		}

		if err := s.userRepo.CreateOrUpdate(ctx, user); err != nil {
			s.logger.Error("failed to create or update user", zap.Error(err), zap.String("user_id", member.UserID))
			return fmt.Errorf("failed to create or update user %s: %w", member.UserID, err)
		}
	}

	return nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.GetByTeam(ctx, teamName)
	if err != nil {
		s.logger.Error("failed to get team users", zap.Error(err), zap.String("team_name", teamName))
		return nil, fmt.Errorf("failed to get team users: %w", err)
	}

	userMap := make(map[string]*domain.User)
	for _, user := range users {
		userMap[user.UserID] = user
	}

	updatedMembers := make([]domain.TeamMember, 0, len(team.Members))
	for _, member := range team.Members {
		if user, exists := userMap[member.UserID]; exists {
			updatedMembers = append(updatedMembers, domain.TeamMember{
				UserID:   user.UserID,
				Username: user.Username,
				IsActive: user.IsActive,
			})
		} else {
			updatedMembers = append(updatedMembers, member)
		}
	}

	team.Members = updatedMembers
	return team, nil
}
