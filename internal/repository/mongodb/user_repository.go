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

const usersCollection = "users"

type UserRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewUserRepository(client *Client, logger *zap.Logger) *UserRepository {
	collection := client.Database().Collection(usersCollection)

	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "team_name", Value: 1}},
	})

	return &UserRepository{
		collection: collection,
		logger:     logger,
	}
}

func (r *UserRepository) CreateOrUpdate(ctx context.Context, user *domain.User) error {
	filter := bson.M{"user_id": user.UserID}
	update := bson.M{"$set": user}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.logger.Error("failed to create or update user", zap.Error(err), zap.String("user_id", user.UserID))
		return fmt.Errorf("failed to create or update user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	filter := bson.M{"user_id": userID}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		r.logger.Error("failed to get user by ID", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetActiveByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	filter := bson.M{
		"team_name": teamName,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("failed to find active users by team", zap.Error(err), zap.String("team_name", teamName))
		return nil, fmt.Errorf("failed to find active users by team: %w", err)
	}
	//nolint:errcheck
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error("failed to decode users", zap.Error(err))
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) UpdateIsActive(ctx context.Context, userID string, isActive bool) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"is_active": isActive}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("failed to update user is_active", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("failed to update user is_active: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) GetByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	filter := bson.M{"team_name": teamName}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("failed to find users by team", zap.Error(err), zap.String("team_name", teamName))
		return nil, fmt.Errorf("failed to find users by team: %w", err)
	}
	//nolint:errcheck
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error("failed to decode users", zap.Error(err))
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}
