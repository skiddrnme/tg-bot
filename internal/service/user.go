package service

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "regexp"
    "strings"

    "github.com/askemblerrr/pr-review-bot/internal/repository"
)

var (
    ErrEmptyLogin   = errors.New("github login is empty")
    ErrInvalidLogin = errors.New("invalid github login format")
)

var githubLoginRe = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,38})$`)

type UserService struct {
    repo   repository.UserRepository
    logger *slog.Logger
}

func NewUserService(repo repository.UserRepository, logger *slog.Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}

func (s *UserService) SetGitHubLogin(ctx context.Context, telegramID int64, login string) error {
    login = strings.TrimSpace(login)
    if login == "" {
        return ErrEmptyLogin
    }
    if !githubLoginRe.MatchString(login) {
        return fmt.Errorf("%w: %q", ErrInvalidLogin, login)
    }
    if err := s.repo.Save(ctx, telegramID, login); err != nil {
        return fmt.Errorf("save user: %w", err)
    }
    s.logger.Info("user saved", "telegram_id", telegramID, "github_login", login)
    return nil
}