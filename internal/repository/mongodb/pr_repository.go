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

const prsCollection = "pull_requests"

type PRRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewPRRepository(client *Client, logger *zap.Logger) *PRRepository {
	collection := client.Database().Collection(prsCollection)

	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "pull_request_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "assigned_reviewers", Value: 1}},
	})

	_, _ = collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "author_id", Value: 1}},
	})

	return &PRRepository{
		collection: collection,
		logger:     logger,
	}
}

func (r *PRRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	_, err := r.collection.InsertOne(ctx, pr)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrPRExists
		}

		r.logger.Error("failed to create PR", zap.Error(err), zap.String("pr_id", pr.PullRequestID))
		return fmt.Errorf("failed to create PR: %w", err)
	}

	return nil
}

func (r *PRRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	filter := bson.M{"pull_request_id": prID}

	err := r.collection.FindOne(ctx, filter).Decode(&pr)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrPRNotFound
	}
	if err != nil {
		r.logger.Error("failed to get PR by ID", zap.Error(err), zap.String("pr_id", prID))
		return nil, fmt.Errorf("failed to get PR by ID: %w", err)
	}

	return &pr, nil
}

func (r *PRRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	filter := bson.M{"pull_request_id": pr.PullRequestID}
	update := bson.M{"$set": pr}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("failed to update PR", zap.Error(err), zap.String("pr_id", pr.PullRequestID))
		return fmt.Errorf("failed to update PR: %w", err)
	}

	return nil
}

func (r *PRRepository) Exists(ctx context.Context, prID string) (bool, error) {
	filter := bson.M{"pull_request_id": prID}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("failed to check PR existence", zap.Error(err), zap.String("pr_id", prID))
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}

	return count > 0, nil
}

func (r *PRRepository) GetByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	filter := bson.M{"assigned_reviewers": userID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("failed to find PRs by reviewer", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to find PRs by reviewer: %w", err)
	}
	//nolint:errcheck
	defer cursor.Close(ctx)

	var prs []*domain.PullRequest
	if err := cursor.All(ctx, &prs); err != nil {
		r.logger.Error("failed to decode PRs", zap.Error(err))
		return nil, fmt.Errorf("failed to decode PRs: %w", err)
	}

	return prs, nil
}

func (r *PRRepository) GetOpenByTeam(ctx context.Context, teamName string) ([]*domain.PullRequest, error) {
	filter := bson.M{"status": domain.PRStatusOpen}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("failed to find open PRs", zap.Error(err))
		return nil, fmt.Errorf("failed to find open PRs: %w", err)
	}
	//nolint:errcheck
	defer cursor.Close(ctx)

	var prs []*domain.PullRequest
	if err := cursor.All(ctx, &prs); err != nil {
		r.logger.Error("failed to decode PRs", zap.Error(err))
		return nil, fmt.Errorf("failed to decode PRs: %w", err)
	}

	return prs, nil
}
