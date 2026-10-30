package telegram_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/askemblerrr/pr-review-bot/internal/delivery/telegram"
	"github.com/askemblerrr/pr-review-bot/internal/repository/memory"
	"github.com/askemblerrr/pr-review-bot/internal/service"
	"github.com/askemblerrr/pr-review-bot/internal/types"
)

const (
	userID int64 = 42
	chatID int64 = 100
)

// fakeSender запоминает все ответы бота вместо отправки в Telegram.
type fakeSender struct {
	replies []string
}

func (s *fakeSender) Send(_ int64, text string) error {
	s.replies = append(s.replies, text)
	return nil
}

// newHandler собирает настоящий UserService поверх in-memory репозитория.
// Логин "taken" заранее привязан к другому пользователю — для сценария с занятым логином.
func newHandler(t *testing.T) (*telegram.Handler, *fakeSender) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := memory.NewUserRepository()
	require.NoError(t, repo.Save(context.Background(), 999, "taken"))

	sender := &fakeSender{}
	return telegram.NewHandler(service.NewUserService(repo, logger), sender, logger), sender
}

func TestHandle(t *testing.T) {
	tests := []struct {
		name      string
		command   string   // команда без слеша: "start", "set_github", ""
		args      []string // аргументы команды
		wantReply string   // подстрока, которая должна быть в единственном ответе бота
	}{
		{
			name:      "старт",
			command:   "start",
			args:      []string{},
			wantReply: "Привет",
		},
		{
			name:      "неизвестная",
			command:   "foo",
			args:      []string{},
			wantReply: "Неизвестная команда",
		},
		{
			name:      "просто привет",
			command:   "",
			args:      []string{},
			wantReply: "Неизвестная команда",
		},
		{
			name:      "set_github без логина",
			command:   "set_github",
			args:      []string{},
			wantReply: "Использование: /set_github <github_login>",
		},
		{
			name:      "set_github a",
			command:   "set_github",
			args:      []string{"a"},
			wantReply: "GitHub-логин сохранён",
		},
		{
			name:      "set_github octocat",
			command:   "set_github",
			args:      []string{"octocat"},
			wantReply: "GitHub-логин сохранён",
		},
		{
			name:      "set_github -bad",
			command:   "set_github",
			args:      []string{"-bad"},
			wantReply: "Неверный формат GitHub-логина",
		},
		{
			name:      "set_github taken",
			command:   "set_github",
			args:      []string{"taken"},
			wantReply: "Этот GitHub-логин уже привязан к другому пользователю",
		},
		{
			name:      "set_github a b",
			command:   "set_github",
			args:      []string{"a", "b"},
			wantReply: "Использование: /set_github <github_login>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, sender := newHandler(t)

			h.Handle(context.Background(), &types.Message{
				UserID: userID, ChatID: chatID, Command: tt.command, Args: tt.args,
			})

			require.Len(t, sender.replies, 1, "бот должен ответить ровно одним сообщением")
			assert.Contains(t, sender.replies[0], tt.wantReply)
		})
	}
}

func TestHandle_InternalError(t *testing.T) {
	h, sender := newHandler(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменённый контекст — memory-репозиторий вернёт ошибку, как упавшая база

	h.Handle(ctx, &types.Message{UserID: userID, ChatID: chatID, Command: "set_github", Args: []string{"octocat"}})

	require.Len(t, sender.replies, 1)
	assert.Contains(t, sender.replies[0], "Внутренняя ошибка")
}
