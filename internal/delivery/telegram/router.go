package telegram

import (
    "context"

    "github.com/askemblerrr/pr-review-bot/internal/types"
)

func (h *Handler) route(ctx context.Context, msg *types.Message) {
    switch msg.Command {
    case "start":
        h.handleStart(ctx, msg)
    case "set_github":
        h.handleSetGitHub(ctx, msg)
    default:
        h.reply(ctx, msg.ChatID, "Неизвестная команда. Используй /set_github <login>")
    }
}