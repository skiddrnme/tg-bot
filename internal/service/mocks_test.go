package service_test

import (
	"context"
	"io"
	"log/slog"

	"github.com/stretchr/testify/mock"
)

// mockUserRepo реализует repository.UserRepository.
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Save(ctx context.Context, telegramID int64, githubLogin string) error {
	args := m.Called(ctx, telegramID, githubLogin)
	return args.Error(0)
}

func (m *mockUserRepo) GetGitHubLogin(ctx context.Context, telegramID int64) (string, error) {
	args := m.Called(ctx, telegramID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepo) GetTelegramID(ctx context.Context, githubLogin string) (int64, error) {
	args := m.Called(ctx, githubLogin)
	return args.Get(0).(int64), args.Error(1)
}

// mockNotifier реализует service.Notifier.
type mockNotifier struct {
	mock.Mock
}

func (m *mockNotifier) Send(chatID int64, text string) error {
	args := m.Called(chatID, text)
	return args.Error(0)
}

// discardLogger глушит логи, чтобы не засорять вывод go test.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
