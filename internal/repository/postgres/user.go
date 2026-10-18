package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/askemblerrr/pr-review-bot/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

	var pgErr *pgconn.PgError

	query :=
		`INSERT INTO users (telegram_id, github_login) 
	VALUES ($1, $2) 
	ON CONFLICT (telegram_id)
	DO UPDATE SET github_login = EXCLUDED.github_login`
	_, err := r.pool.Exec(ctx, query, telegramID, strings.ToLower(githubLogin))
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrLoginTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetGitHubLogin(ctx context.Context, telegramID int64) (string, error) {

	var login string
	query := `SELECT github_login FROM users WHERE telegram_id=$1`
	err := r.pool.QueryRow(ctx, query, telegramID).Scan(&login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrUserNotFound
		}
		return "", fmt.Errorf("select github login: %w", err)
	}

	return login, nil
}

func (r *UserRepository) GetTelegramID(ctx context.Context, githubLogin string) (int64, error) {
	var id int64

	query := `SELECT telegram_id FROM users WHERE github_login=$1`
	err := r.pool.QueryRow(ctx, query, strings.ToLower(githubLogin)).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, repository.ErrUserNotFound
		}
		return 0, fmt.Errorf("select telegram id: %w", err)
	}

	return id, nil

}
