package telegram

import (
    "context"
    "errors"
    "log/slog"

    "github.com/askemblerrr/pr-review-bot/internal/service"
    "github.com/askemblerrr/pr-review-bot/internal/types"
)

type Sender interface {
    Send(chatID int64, text string) error
}

type Handler struct {
    userSvc *service.UserService
    sender  Sender
    logger  *slog.Logger
}

func NewHandler(userSvc *service.UserService, sender Sender, logger *slog.Logger) *Handler {
    return &Handler{userSvc: userSvc, sender: sender, logger: logger}
}

func (h *Handler) Handle(ctx context.Context, msg *types.Message) {
    h.logger.Info("incoming", "user_id", msg.UserID, "cmd", msg.Command)
    h.route(ctx, msg)
}

func (h *Handler) handleStart(ctx context.Context, msg *types.Message) {
    h.reply(ctx, msg.ChatID, "Привет! Отправь /set_github <твой_github_login>")
}

func (h *Handler) handleSetGitHub(ctx context.Context, msg *types.Message) {
    if len(msg.Args) != 1 {
        h.reply(ctx, msg.ChatID, "Использование: /set_github <github_login>")
        return
    }
    err := h.userSvc.SetGitHubLogin(ctx, msg.UserID, msg.Args[0])
    switch {
    case err == nil:
        h.reply(ctx, msg.ChatID, "✅ GitHub-логин сохранён")
    case errors.Is(err, service.ErrInvalidLogin), errors.Is(err, service.ErrEmptyLogin):
        h.reply(ctx, msg.ChatID, "❌ Неверный формат GitHub-логина")
    default:
        h.logger.Error("set github login failed", "err", err)
        h.reply(ctx, msg.ChatID, "❌ Внутренняя ошибка")
    }
}

func (h *Handler) reply(ctx context.Context, chatID int64, text string) {
    if err := h.sender.Send(chatID, text); err != nil {
        h.logger.Error("send reply failed", "err", err, "chat_id", chatID)
    }
}