package mongodb

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestClientDatabase(t *testing.T) {
	client := &Client{
		db:     nil,
		logger: zap.NewNop(),
	}

	db := client.Database()
	if db != nil {
		t.Errorf("expected nil database, got %v", db)
	}

	// Test with actual database will tested in integration tests
}

func TestClientClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	// This would require a real MongoDB instance
	uri := "mongodb://localhost:27017"
	dbName := "test_db"
	timeout := 5 * time.Second

	client, err := NewClient(ctx, uri, dbName, timeout, logger)
	if err != nil {
		t.Skipf("Skipping integration test - MongoDB not available: %v", err)
		return
	}

	err = client.Close(ctx)
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestNewClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	logger := zap.NewNop()

	// This would require a real MongoDB instance in integration tests
	uri := "mongodb://localhost:27017"
	dbName := "test_db"
	timeout := 5 * time.Second

	client, err := NewClient(ctx, uri, dbName, timeout, logger)
	if err != nil {
		t.Skipf("Skipping integration test - MongoDB not available: %v", err)
		return
	}

	defer client.Close(ctx)

	if client == nil {
		t.Fatal("expected client, got nil")
	}

	db := client.Database()
	if db == nil {
		t.Fatal("expected database, got nil")
	}

	if db.Name() != dbName {
		t.Errorf("expected database name %s, got %s", dbName, db.Name())
	}
}
