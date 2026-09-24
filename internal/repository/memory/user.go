package memory

import (
    "context"
    "sync"

    "github.com/askemblerrr/pr-review-bot/internal/repository"
)

type UserRepository struct {
    mu         sync.RWMutex
    byTelegram map[int64]string
    byGitHub   map[string]int64
}

func NewUserRepository() *UserRepository {
    return &UserRepository{
        byTelegram: make(map[int64]string),
        byGitHub:   make(map[string]int64),
    }
}

func (r *UserRepository) Save(ctx context.Context, telegramID int64, githubLogin string) error {
    if err := ctx.Err(); err != nil {
        return err
    }

    r.mu.Lock()           // write lock — потому что меняем оба map
    defer r.mu.Unlock()   // ← Unlock, а не RUnlock

    if old, ok := r.byTelegram[telegramID]; ok && old != githubLogin {
        delete(r.byGitHub, old)
    }
    r.byTelegram[telegramID] = githubLogin
    r.byGitHub[githubLogin] = telegramID
    return nil
}

func (r *UserRepository) GetGitHubLogin(ctx context.Context, telegramID int64) (string, error) {
    if err := ctx.Err(); err != nil {
        return "", err
    }

    r.mu.RLock()          // read lock — только читаем
    defer r.mu.RUnlock()  // ← RUnlock — симметрично RLock

    login, ok := r.byTelegram[telegramID]
    if !ok {
        return "", repository.ErrUserNotFound
    }
    return login, nil
}

func (r *UserRepository) GetTelegramID(ctx context.Context, githubLogin string) (int64, error) {
    if err := ctx.Err(); err != nil {
        return 0, err
    }

    r.mu.RLock()
    defer r.mu.RUnlock()

    id, ok := r.byGitHub[githubLogin]
    if !ok {
        return 0, repository.ErrUserNotFound
    }
    return id, nil
}