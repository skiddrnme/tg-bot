package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var errNotImplemented = errors.New("not implemented")

type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создаёт пул соединений и проверяет, что база отвечает.
// Вызывающий код обязан вызвать Close при остановке приложения.
func NewUserRepository(ctx context.Context, dsn string) (*UserRepository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &UserRepository{pool: pool}, nil
}

func (r *UserRepository) Close() {
	r.pool.Close()
}

func (r *UserRepository) Save(ctx context.Context, telegramID int64, githubLogin string) error {
	return errNotImplemented
}

func (r *UserRepository) GetGitHubLogin(ctx context.Context, telegramID int64) (string, error) {
	// TODO(human): найти github_login по telegram_id.
	// Нет строки — repository.ErrUserNotFound, любая другая ошибка — обернуть через fmt.Errorf с %w.
	return "", errNotImplemented
}

func (r *UserRepository) GetTelegramID(ctx context.Context, githubLogin string) (int64, error) {
	return 0, errNotImplemented
}
