package mongodb

import (
	"context"
	"fmt"

	"assignment-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const teamsCollection = "teams"

type TeamRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewTeamRepository(client *Client, logger *zap.Logger) *TeamRepository {
	collection := client.Database().Collection(teamsCollection)

	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "team_name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &TeamRepository{
		collection: collection,
		logger:     logger,
	}
}

func (r *TeamRepository) Create(ctx context.Context, team *domain.Team) error {
	_, err := r.collection.InsertOne(ctx, team)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrTeamExists
		}

		r.logger.Error("failed to create team", zap.Error(err), zap.String("team_name", team.TeamName))
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

func (r *TeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	var team domain.Team
	filter := bson.M{"team_name": teamName}

	err := r.collection.FindOne(ctx, filter).Decode(&team)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrTeamNotFound
	}
	if err != nil {
		r.logger.Error("failed to get team by name", zap.Error(err), zap.String("team_name", teamName))
		return nil, fmt.Errorf("failed to get team by name: %w", err)
	}

	return &team, nil
}

func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	filter := bson.M{"team_name": teamName}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("failed to check team existence", zap.Error(err), zap.String("team_name", teamName))
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return count > 0, nil
}
