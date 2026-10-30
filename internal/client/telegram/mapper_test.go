package telegram_test

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/askemblerrr/pr-review-bot/internal/client/telegram"
)

func update(text string, commandLen int) tgbotapi.Update {
	m := &tgbotapi.Message{
		Text: text,
		Chat: &tgbotapi.Chat{ID: 100},
		From: &tgbotapi.User{ID: 42, UserName: "asker"},
	}
	if commandLen > 0 {
		m.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: commandLen}}
	}
	return tgbotapi.Update{Message: m}
}

func TestToDomain(t *testing.T) {
	t.Run("команда с аргументом", func(t *testing.T) {
		msg := telegram.ToDomain(update("/set_github octocat", len("/set_github")))

		require.NotNil(t, msg)
		assert.Equal(t, "set_github", msg.Command)
		assert.Equal(t, []string{"octocat"}, msg.Args)
		assert.Equal(t, int64(42), msg.UserID)
		assert.Equal(t, int64(100), msg.ChatID)
		assert.Equal(t, "asker", msg.Username)
	})

	t.Run("обычный текст — не команда", func(t *testing.T) {
		msg := telegram.ToDomain(update("привет", 0))

		require.NotNil(t, msg)
		assert.Empty(t, msg.Command)
		assert.Empty(t, msg.Args)
	})

	t.Run("подпись к фото вместо текста", func(t *testing.T) {
		u := update("", 0)
		u.Message.Caption = "фото"

		msg := telegram.ToDomain(u)

		require.NotNil(t, msg)
		assert.Equal(t, "фото", msg.Text)
	})

	t.Run("без текста — отбрасывается", func(t *testing.T) {
		assert.Nil(t, telegram.ToDomain(update("", 0)))
	})

	t.Run("не сообщение — отбрасывается", func(t *testing.T) {
		assert.Nil(t, telegram.ToDomain(tgbotapi.Update{}))
	})
}
