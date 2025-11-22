package mocks

import (
	"context"

	"assignment-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepository) Exists(ctx context.Context, prID string) (bool, error) {
	args := m.Called(ctx, prID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPRRepository) GetByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PullRequest), args.Error(1)
}

func (m *MockPRRepository) GetOpenByTeam(ctx context.Context, teamName string) ([]*domain.PullRequest, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PullRequest), args.Error(1)
}
