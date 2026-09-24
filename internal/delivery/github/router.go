package github

import "context"

const (
    eventPullRequest          = "pull_request"
    eventPRReview             = "pull_request_review"
    eventPRReviewComment      = "pull_request_review_comment"
    eventPRReviewThread       = "pull_request_review_thread"
)

func (h *Handler) route(ctx context.Context, event string, body []byte) error {
    switch event {
    case eventPullRequest:
        return h.handlePullRequest(ctx, body)
    case eventPRReview, eventPRReviewComment, eventPRReviewThread:
        h.logger.Info("event ignored (not implemented)", "event", event)
        return nil
    default:
        h.logger.Debug("event ignored", "event", event)
        return nil
    }
}