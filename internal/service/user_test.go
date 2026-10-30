package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/askemblerrr/pr-review-bot/internal/service"
)

var errDB = errors.New("db is down")
var errSaveFailed = errors.New("save failed")
func TestUserService_SetGitHubLogin(t *testing.T) {
	const telegramID int64 = 42

	tests := []struct {
		name      string
		login     string // что пользователь передал в /set_github
		wantSaved string // с каким логином ожидается вызов repo.Save; "" — Save не должен вызываться
		saveErr   error  // что вернёт repo.Save
		wantErr   error  // ожидаемая ошибка (проверяется через errors.Is); nil — успех
	}{
		{name: "валидный", login: "octocat", wantSaved: "octocat"},
        {name: "дефис и цифры", login: "my-login-123", wantSaved: "my-login-123"},
        {name: "макс длина 39", login: strings.Repeat("a", 39), wantSaved: strings.Repeat("a", 39)},

        {name: "пустой", login: "", wantErr: service.ErrEmptyLogin},
        {name: "с дефиса", login: "-octocat", wantErr: service.ErrInvalidLogin},
        {name: "40 символов", login: strings.Repeat("a", 40), wantErr: service.ErrInvalidLogin},
        {name: "подчёркивание", login: "user_name", wantErr: service.ErrInvalidLogin},
        {name: "кириллица", login: "юзер", wantErr: service.ErrInvalidLogin},

        {
            name: "ошибка репозитория",
            login: "octocat", wantSaved: "octocat",
            saveErr: errSaveFailed, wantErr: errSaveFailed,
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockUserRepo)
			if tt.wantSaved != "" {
				repo.On("Save", mock.Anything, telegramID, tt.wantSaved).Return(tt.saveErr)
			}
			svc := service.NewUserService(repo, discardLogger())

			err := svc.SetGitHubLogin(context.Background(), telegramID, tt.login)

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
			if tt.wantSaved == "" {
				repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything)
			}
			repo.AssertExpectations(t)
		})
	}
}
