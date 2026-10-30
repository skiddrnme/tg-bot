package github_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/askemblerrr/pr-review-bot/internal/delivery/github"
	"github.com/askemblerrr/pr-review-bot/internal/types"
)

const (
	testSecret   = "test-secret"
	testUsername = "octocat"
)

type mockNotifSvc struct {
	mock.Mock
}

func (m *mockNotifSvc) NotifyPRAssigned(ctx context.Context, ev types.PRAssignedEvent) error {
	return m.Called(ctx, ev).Error(0)
}

func (m *mockNotifSvc) NotifyPRReview(ctx context.Context, ev types.PRReviewEvent) error {
	return m.Called(ctx, ev).Error(0)
}

func (m *mockNotifSvc) NotifyPRReviewComment(ctx context.Context, ev types.PRReviewCommentEvent) error {
	return m.Called(ctx, ev).Error(0)
}

func (m *mockNotifSvc) NotifyPRReviewThread(ctx context.Context, ev types.PRReviewThreadEvent) error {
	return m.Called(ctx, ev).Error(0)
}

// sign считает подпись так же, как GitHub: "sha256=" + hex(HMAC-SHA256(body, secret)).
func sign(body, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// serve отправляет подписанный testSecret вебхук в Handler и возвращает HTTP-статус ответа.
func serve(svc github.NotificationService, event, body string) int {
	return serveWithSignature(svc, event, body, sign(body, testSecret))
}

func serveWithSignature(svc github.NotificationService, event, body, signature string) int {
	h := github.NewHandler(svc, testUsername, testSecret, slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/github/webhook", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", event)
	if signature != "" {
		req.Header.Set("X-Hub-Signature-256", signature)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestServeHTTP_Signature(t *testing.T) {
	const body = `{"action":"opened"}`

	tests := []struct {
		name      string
		signature string
	}{
		{"без заголовка", ""},
		{"без префикса sha256=", strings.TrimPrefix(sign(body, testSecret), "sha256=")},
		{"не hex", "sha256=zzzz"},
		{"чужой секрет", sign(body, "other-secret")},
		{"подпись другого тела", sign(`{"action":"assigned"}`, testSecret)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockNotifSvc)

			code := serveWithSignature(svc, "pull_request", body, tt.signature)

			assert.Equal(t, http.StatusUnauthorized, code)
			svc.AssertNotCalled(t, "NotifyPRAssigned", mock.Anything, mock.Anything)
		})
	}
}

func TestServeHTTP_AlwaysOK(t *testing.T) {
	t.Run("неизвестное событие", func(t *testing.T) {
		svc := new(mockNotifSvc)
		assert.Equal(t, http.StatusOK, serve(svc, "push", `{}`))
		svc.AssertExpectations(t) // никаких вызовов не ожидалось
	})

	t.Run("битый JSON", func(t *testing.T) {
		svc := new(mockNotifSvc)
		assert.Equal(t, http.StatusOK, serve(svc, "pull_request", `{not json`))
	})

	t.Run("ошибка сервиса не превращается в 5xx", func(t *testing.T) {
		svc := new(mockNotifSvc)
		svc.On("NotifyPRAssigned", mock.Anything, mock.Anything).Return(errors.New("boom"))
		body := `{"action":"assigned","assignee":{"login":"octocat"},"pull_request":{"title":"T","html_url":"U"}}`

		assert.Equal(t, http.StatusOK, serve(svc, "pull_request", body))
		svc.AssertExpectations(t)
	})
}

func TestPullRequest_Filtering(t *testing.T) {
	tests := []struct {
		name string
		body string                 // JSON-payload события pull_request
		want *types.PRAssignedEvent // с чем ожидается вызов NotifyPRAssigned; nil — вызова быть не должно
	}{
		{
			name: "назначили меня",
			body: `{"action":"assigned","assignee":{"login":"octocat"},"pull_request":{"title":"Fix","html_url":"https://x"}}`,
			want: &types.PRAssignedEvent{Assignee: "octocat", PRTitle: "Fix", PRURL: "https://x"},
		},
		{
			name: "PR просто открыли",
			body:  `{"action":"opened","assignee":{"login":"octocat"},"pull_request":{"title":"Fix","html_url":"https://x"}}`,
			want: nil,
		},
		{
			name: "назначили другого",
			body: `{"action":"assigned","assignee":{"login":"alice"},"pull_request":{"title":"Fix","html_url":"https://x"}}`,
			want: nil,
		},
		{
			name: "поля assignee нет вовсе",
			body: `{"action":"assigned","pull_request":{"title":"Fix","html_url":"https://x"}}`,
			want: nil,
		},
		{
			name:  "мой логин большими буквами",
			body: `{"action":"assigned","assignee":{"login":"OctoCat"},"pull_request":{"title":"Fix","html_url":"https://x"}}`,
			want: &types.PRAssignedEvent{Assignee: "OctoCat", PRTitle: "Fix", PRURL: "https://x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockNotifSvc)
			if tt.want != nil {
				svc.On("NotifyPRAssigned", mock.Anything, *tt.want).Return(nil)
			}

			code := serve(svc, "pull_request", tt.body)

			assert.Equal(t, http.StatusOK, code)
			if tt.want == nil {
				svc.AssertNotCalled(t, "NotifyPRAssigned", mock.Anything, mock.Anything)
			}
			svc.AssertExpectations(t)
		})
	}
}
