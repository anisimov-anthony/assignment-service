package service

import (
	"context"

	"assignment-service/internal/domain"
	"assignment-service/internal/repository"

	"go.uber.org/zap"
)

type UserService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo repository.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateIsActive(ctx, userID, isActive); err != nil {
		return nil, err
	}

	user.IsActive = isActive
	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// DeactivateTeamMembers массово деактивирует пользователей команды.
func (s *UserService) DeactivateTeamMembers(ctx context.Context, teamName string) error {
	users, err := s.userRepo.GetByTeam(ctx, teamName)
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.IsActive {
			if err := s.userRepo.UpdateIsActive(ctx, user.UserID, false); err != nil {
				s.logger.Error("failed to deactivate user", zap.Error(err), zap.String("user_id", user.UserID))
				return err
			}
		}
	}

	return nil
}
