package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Client encapsula o cliente MongoDB usado pela aplicação.
type Client struct {
	mongo *mongo.Client
}

// Connect cria o cliente, valida a URI e confirma que o servidor responde.
func Connect(ctx context.Context, uri string) (*Client, error) {
	if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
		return nil, fmt.Errorf("invalid mongo uri")
	}

	opts := options.Client().ApplyURI(uri).SetServerSelectionTimeout(5 * time.Second)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return &Client{mongo: client}, nil
}

// Ping confirma que o MongoDB continua acessível.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.mongo == nil {
		return fmt.Errorf("mongo client is not initialized")
	}
	return c.mongo.Ping(ctx, nil)
}

// Close encerra a conexão com o MongoDB.
func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.mongo == nil {
		return nil
	}
	return c.mongo.Disconnect(ctx)
}
