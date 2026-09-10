package repository

import (
	"context"
	"sync"
)

type UserRepository interface {
	SaveUser(ctx context.Context, telegramID int64, githubLogin string) error
	GetGitHubLogin(ctx context.Context, telegramID int64) (string, error)
	GetTelegramID(ctx context.Context, githubLogin string) (int64, error)
}

type InMemoryRepo struct {
    mu    sync.RWMutex
    users map[int64]string // telegramID -> githubLogin
}