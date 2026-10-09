package memory_test

import (
	"context"
	"testing"

	"github.com/askemblerrr/pr-review-bot/internal/repository"
	"github.com/askemblerrr/pr-review-bot/internal/repository/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositorySaveAndGet(t *testing.T) {
	repo := memory.NewUserRepository()

	err := repo.Save(context.Background(), 123, "octocat")
	require.NoError(t, err)

	login, err := repo.GetGitHubLogin(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, "octocat", login)

	id, err := repo.GetTelegramID(context.Background(), "octocat")
	require.NoError(t, err)
	assert.Equal(t, int64(123), id)
}

func TestUserRepositorySaveOverwrites(t *testing.T) {
	repo := memory.NewUserRepository()

	err := repo.Save(context.Background(), 123, "alice")
	require.NoError(t, err)

	err = repo.Save(context.Background(), 123, "bob")
	require.NoError(t, err)

	login, err := repo.GetGitHubLogin(context.Background(), 123)
	require.NoError(t, err)
	assert.Equal(t, "bob", login)

	_, err = repo.GetTelegramID(context.Background(), "alice")
	require.ErrorIs(t, err, repository.ErrUserNotFound)

	id, err := repo.GetTelegramID(context.Background(), "bob")
	require.NoError(t, err)
	assert.Equal(t, int64(123), id)
}

func TestUserRepositoryNotFound(t *testing.T) {
	repo := memory.NewUserRepository()

	_, err := repo.GetGitHubLogin(context.Background(), 123)
	require.ErrorIs(t, err, repository.ErrUserNotFound)

	_, err = repo.GetTelegramID(context.Background(), "alice")
	require.ErrorIs(t, err, repository.ErrUserNotFound)
}
