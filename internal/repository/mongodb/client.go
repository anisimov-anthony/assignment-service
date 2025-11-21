package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Client struct {
	db     *mongo.Database
	logger *zap.Logger
}

func NewClient(ctx context.Context, uri, dbName string, connectTimeout time.Duration, logger *zap.Logger) (*Client, error) {
	opts := options.Client().ApplyURI(uri)

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client, err := mongo.Connect(connectCtx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), connectTimeout)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(dbName)

	logger.Info("Connected to MongoDB successfully",
		zap.String("database", dbName),
		zap.Duration("connect_timeout", connectTimeout))

	return &Client{
		db:     db,
		logger: logger,
	}, nil
}

func (c *Client) Database() *mongo.Database {
	return c.db
}

func (c *Client) Close(ctx context.Context) error {
	return c.db.Client().Disconnect(ctx)
}
