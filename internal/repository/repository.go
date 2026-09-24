package repository

import (
	"context"
	"errors"
)


var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
    Save(ctx context.Context, telegramID int64, githubLogin string) error
    GetGitHubLogin(ctx context.Context, telegramID int64) (string, error)
    GetTelegramID(ctx context.Context, githubLogin string) (int64, error)
}