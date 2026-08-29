// Package kms owns the server's integration boundary with OpenBao.
package kms

import (
	"context"
	"fmt"

	bao "github.com/openbao/openbao/api/v2"
)

// Config describes how to connect to OpenBao.
type Config struct {
	Address string
	Token   string
}

// Health is the subset of OpenBao health state needed during startup.
type Health struct {
	Initialized bool
	Sealed      bool
	Version     string
}

// Client wraps the OpenBao API client. Key lifecycle operations will be added here.
type Client struct {
	client *bao.Client
}

// New creates an OpenBao client without performing any key operations.
func New(config Config) (*Client, error) {
	if config.Address == "" {
		return nil, fmt.Errorf("OpenBao address is required")
	}

	clientConfig := bao.NewConfig()
	clientConfig.Address = config.Address
	clientConfig.DisableEnvironment = true

	client, err := bao.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("create OpenBao client: %w", err)
	}
	client.SetToken(config.Token)

	return &Client{client: client}, nil
}

// Health verifies that OpenBao is reachable. Sealed and uninitialized servers are
// reported as state rather than errors so production can be initialized out of band.
func (c *Client) Health(ctx context.Context) (Health, error) {
	response, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return Health{}, fmt.Errorf("check OpenBao health: %w", err)
	}

	return Health{
		Initialized: response.Initialized,
		Sealed:      response.Sealed,
		Version:     response.Version,
	}, nil
}
