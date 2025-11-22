package mocks

import (
	"context"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, team *domain.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}
