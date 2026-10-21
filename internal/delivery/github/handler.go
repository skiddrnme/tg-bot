package github

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/askemblerrr/pr-review-bot/internal/types"
)

const maxBodySize = 1 << 20

type NotificationService interface {
	NotifyPRReview(ctx context.Context, ev types.PRReviewEvent) error
	NotifyPRAssigned(ctx context.Context, ev types.PRAssignedEvent) error
	NotifyPRReviewThread(ctx context.Context, ev types.PRReviewThreadEvent) error
	NotifyPRReviewComment(ctx context.Context, ev types.PRReviewCommentEvent) error
}

type Handler struct {
	notifSvc NotificationService
	secret   string
	username string
	logger   *slog.Logger
}

func NewHandler(notifSvc NotificationService, username, secret string, logger *slog.Logger) *Handler {
	return &Handler{notifSvc: notifSvc, secret: secret, username: username, logger: logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err != nil {
		h.logger.Warn("read body failed", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !VerifySignature(body, r.Header.Get("X-Hub-Signature-256"), h.secret) {
		h.logger.Warn("invalid signature", "remote", r.RemoteAddr)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	h.logger.Info("webhook received", "event", event, "delivery", r.Header.Get("X-GitHub-Delivery"))

	if err := h.route(r.Context(), event, body); err != nil {
		h.logger.Error("handle event failed", "event", event, "err", err)
	}
	w.WriteHeader(http.StatusOK)
}
