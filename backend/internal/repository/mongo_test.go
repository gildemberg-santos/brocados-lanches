package repository

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConnectRejectsInvalidURI(t *testing.T) {
	tests := []string{
		"",
		"http://localhost:27017",
		"://missing-scheme",
	}

	for _, uri := range tests {
		t.Run(uri, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if _, err := Connect(ctx, uri); err == nil {
				t.Fatal("Connect() error = nil, want error")
			}
		})
	}
}

func TestNilClient(t *testing.T) {
	var client *Client
	ctx := context.Background()

	if err := client.Ping(ctx); err == nil {
		t.Fatal("Ping() error = nil, want error")
	}
	if err := client.Close(ctx); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}

func TestConnectIntegration(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, uri)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() {
		if err := client.Close(context.Background()); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}
