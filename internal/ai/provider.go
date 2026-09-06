package ai

import (
	"context"
)

// Provider defines the interface for generating commit messages via AI models.
type Provider interface {
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)
	TestConnection(ctx context.Context) error
}

// NewProvider returns an AI Provider configured according to the configuration.
func NewProvider(cfg Config) Provider {
	return NewOpenAICompatible(cfg)
}
