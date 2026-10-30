package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/askemblerrr/pr-review-bot/internal/repository"
	"github.com/askemblerrr/pr-review-bot/internal/service"
	"github.com/askemblerrr/pr-review-bot/internal/types"
)

const chatID int64 = 7

// newNotificationService собирает сервис с моками; repo настраивается в тесте.
func newNotificationService(repo *mockUserRepo, notifier *mockNotifier) *service.NotificationService {
	return service.NewNotificationService(repo, notifier, discardLogger())
}

// captureSend ожидает ровно один вызов Send в chatID и возвращает указатель на отправленный текст.
func captureSend(notifier *mockNotifier) *string {
	var sent string
	notifier.On("Send", chatID, mock.Anything).
		Run(func(args mock.Arguments) { sent = args.String(1) }).
		Return(nil).
		Once()
	return &sent
}

func TestNotifyPRAssigned(t *testing.T) {
	ev := types.PRAssignedEvent{Assignee: "octocat", PRTitle: "Fix bug", PRURL: "https://github.com/o/r/pull/1"}

	t.Run("успех", func(t *testing.T) {
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(chatID, nil)
		sent := captureSend(notifier)

		err := newNotificationService(repo, notifier).NotifyPRAssigned(context.Background(), ev)

		require.NoError(t, err)
		assert.Contains(t, *sent, "Fix bug")
		assert.Contains(t, *sent, ev.PRURL)
		notifier.AssertExpectations(t)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(int64(0), repository.ErrUserNotFound)

		err := newNotificationService(repo, notifier).NotifyPRAssigned(context.Background(), ev)

		require.ErrorIs(t, err, repository.ErrUserNotFound)
		notifier.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(int64(0), errDB)

		err := newNotificationService(repo, notifier).NotifyPRAssigned(context.Background(), ev)

		require.ErrorIs(t, err, errDB)
		notifier.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("ошибка отправки", func(t *testing.T) {
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(chatID, nil)
		notifier.On("Send", chatID, mock.Anything).Return(errDB)

		err := newNotificationService(repo, notifier).NotifyPRAssigned(context.Background(), ev)

		require.ErrorIs(t, err, errDB)
	})
}

func TestNotifyPRReview(t *testing.T) {
	base := types.PRReviewEvent{
		AuthorLogin:   "octocat",
		ReviewerLogin: "reviewer",
		PRTitle:       "Fix bug",
		PRURL:         "https://github.com/o/r/pull/1",
		ReviewState:   "approved",
		ReviewBody:    "LGTM",
		Action:        "submitted",
	}

	tests := []struct {
		name            string
		modify          func(ev *types.PRReviewEvent) // меняет копию base под сценарий; nil — base как есть
		wantContains    []string
		wantNotContains []string
	}{
		// ===== Action =====
		{
			name:   "submitted — заголовок «Новое ревью»",
			modify: nil, // base.Action == "submitted"
			wantContains: []string{
				"Новое ревью на ваш PR",
				"Fix bug",
				"@reviewer",
				"https://github.com/o/r/pull/1",
			},
		},
		{
			name: "dismissed — заголовок «Ревью снято»",
			modify: func(ev *types.PRReviewEvent) {
				ev.Action = "dismissed"
			},
			wantContains: []string{
				"Ревью снято с вашего PR",
				"Fix bug",
			},
			wantNotContains: []string{
				"Новое ревью", // заголовок другой ветки не должен просочиться
			},
		},
		{
			name: "неизвестный Action — пустой header (документируем текущее поведение)",
			modify: func(ev *types.PRReviewEvent) {
				ev.Action = "edited"
			},
			wantContains: []string{"Fix bug", "@reviewer"},
			// явно фиксируем, что текста заголовка нет:
			wantNotContains: []string{"Новое ревью", "Ревью снято"},
		},

		// ===== ReviewState =====
		{
			name:            "approved — статус «Одобрено»",
			modify:          nil,
			wantContains:    []string{"Одобрено"},
			wantNotContains: []string{"Запрошены изменения", "Статус: 💬 Комментарий"},
		},
		{
			name: "changes_requested — статус «Запрошены изменения»",
			modify: func(ev *types.PRReviewEvent) {
				ev.ReviewState = "changes_requested"
			},
			wantContains:    []string{"Запрошены изменения"},
			wantNotContains: []string{"Одобрено"},
		},
		{
			name: "commented — статус «Комментарий»",
			modify: func(ev *types.PRReviewEvent) {
				ev.ReviewState = "commented"
			},
			wantContains:    []string{"Комментарий"},
			wantNotContains: []string{"Одобрено", "Запрошены изменения"},
		},
		{
			name: "неизвестный ReviewState — эхо-значение",
			modify: func(ev *types.PRReviewEvent) {
				ev.ReviewState = "pending"
			},
			wantContains: []string{"pending"},
		},

		// ===== ReviewBody =====
		{
			name: "пустое тело — «Без комментария»",
			modify: func(ev *types.PRReviewEvent) {
				ev.ReviewBody = ""
			},
			wantContains: []string{"Без комментария"},
		},
		{
			name: "тело > 1000 рун — обрезается",
			modify: func(ev *types.PRReviewEvent) {
				ev.ReviewBody = strings.Repeat("я", 1500) // руны, не байты
			},
			wantContains: []string{
				"(обрезано)",
				strings.Repeat("я", 1000), // ровно 1000 первых рун
			},
			wantNotContains: []string{
				strings.Repeat("я", 1001), // 1001-я руна не должна попасть
			},
		},
		{
			name: "тело ровно 1000 рун — НЕ обрезается",
			modify: func(ev *types.PRReviewEvent) {
				ev.ReviewBody = strings.Repeat("я", 1000)
			},
			wantContains:    []string{strings.Repeat("я", 1000)},
			wantNotContains: []string{"(обрезано)"}, // граница: > 1000, не >=
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := base
			if tt.modify != nil {
				tt.modify(&ev)
			}
			repo, notifier := new(mockUserRepo), new(mockNotifier)
			repo.On("GetTelegramID", mock.Anything, ev.AuthorLogin).Return(chatID, nil)
			sent := captureSend(notifier)

			err := newNotificationService(repo, notifier).NotifyPRReview(context.Background(), ev)

			require.NoError(t, err)
			for _, s := range tt.wantContains {
				assert.Contains(t, *sent, s)
			}
			for _, s := range tt.wantNotContains {
				assert.NotContains(t, *sent, s)
			}
		})
	}
}

func TestNotifyPRReviewComment(t *testing.T) {
	base := types.PRReviewCommentEvent{
		AuthorLogin:    "octocat",
		CommenterLogin: "reviewer",
		PRTitle:        "Fix bug",
		FilePath:       "main.go",
		Line:           42,
		CommentBody:    "nit: rename",
		CommentURL:     "https://github.com/o/r/pull/1#discussion_r1",
	}

	t.Run("успех", func(t *testing.T) {
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(chatID, nil)
		sent := captureSend(notifier)

		err := newNotificationService(repo, notifier).NotifyPRReviewComment(context.Background(), base)

		require.NoError(t, err)
		for _, s := range []string{"Fix bug", "@reviewer", "main.go :42", "nit: rename", base.CommentURL} {
			assert.Contains(t, *sent, s)
		}
	})

	t.Run("пустой комментарий не отправляется", func(t *testing.T) {
		ev := base
		ev.CommentBody = ""
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(chatID, nil)

		err := newNotificationService(repo, notifier).NotifyPRReviewComment(context.Background(), ev)

		require.NoError(t, err)
		notifier.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})

	t.Run("длинный комментарий обрезается", func(t *testing.T) {
		ev := base
		ev.CommentBody = strings.Repeat("я", 1001)
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(chatID, nil)
		sent := captureSend(notifier)

		err := newNotificationService(repo, notifier).NotifyPRReviewComment(context.Background(), ev)

		require.NoError(t, err)
		assert.Contains(t, *sent, "(обрезано)")
		assert.NotContains(t, *sent, strings.Repeat("я", 1001))
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		repo, notifier := new(mockUserRepo), new(mockNotifier)
		repo.On("GetTelegramID", mock.Anything, "octocat").Return(int64(0), repository.ErrUserNotFound)

		err := newNotificationService(repo, notifier).NotifyPRReviewComment(context.Background(), base)

		require.ErrorIs(t, err, repository.ErrUserNotFound)
		notifier.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
	})
}

func TestNotifyPRReviewThread(t *testing.T) {
	base := types.PRReviewThreadEvent{
		AssigneeLogin: "octocat",
		ActorLogin:    "reviewer",
		PRTitle:       "Fix bug",
		PRURL:         "https://github.com/o/r/pull/1",
		Action:        "resolved",
		FilePath:      "main.go",
		ThreadURL:     "https://github.com/o/r/pull/1#discussion_r1",
	}

	tests := []struct {
		name            string
		modify          func(ev *types.PRReviewThreadEvent)
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "resolved",
			wantContains: []string{"Обсуждение закрыто", "«Fix bug»", "@reviewer", "Файл: main.go", "Ссылка: " + base.ThreadURL},
		},
		{
			name:            "unresolved",
			modify:          func(ev *types.PRReviewThreadEvent) { ev.Action = "unresolved" },
			wantContains:    []string{"Обсуждение возобновлено"},
			wantNotContains: []string{"Обсуждение закрыто"},
		},
		{
			name:         "неизвестный action",
			modify:       func(ev *types.PRReviewThreadEvent) { ev.Action = "edited" },
			wantContains: []string{"Изменение в обсуждении", "Статус: edited"},
		},
		{
			name:            "без ThreadURL — ссылка на PR",
			modify:          func(ev *types.PRReviewThreadEvent) { ev.ThreadURL = "" },
			wantContains:    []string{"Ссылка: " + base.PRURL},
			wantNotContains: []string{"discussion_r1"},
		},
		{
			name:            "без файла и ссылок — строк нет",
			modify:          func(ev *types.PRReviewThreadEvent) { ev.FilePath, ev.ThreadURL, ev.PRURL = "", "", "" },
			wantNotContains: []string{"Файл:", "Ссылка:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := base
			if tt.modify != nil {
				tt.modify(&ev)
			}
			repo, notifier := new(mockUserRepo), new(mockNotifier)
			repo.On("GetTelegramID", mock.Anything, "octocat").Return(chatID, nil)
			sent := captureSend(notifier)

			err := newNotificationService(repo, notifier).NotifyPRReviewThread(context.Background(), ev)

			require.NoError(t, err)
			for _, s := range tt.wantContains {
				assert.Contains(t, *sent, s)
			}
			for _, s := range tt.wantNotContains {
				assert.NotContains(t, *sent, s)
			}
		})
	}
}
