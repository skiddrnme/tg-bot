package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/askemblerrr/pr-review-bot/internal/types"
)

type pullRequestReviewPayload struct {
	Action string `json:"action"`
	Review struct {
		State   string `json:"state"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    user   `json:"user"`
	} `json:"review"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    user   `json:"user"`
	} `json:"pull_request"`
}

func (h *Handler) handlePullRequestReview(ctx context.Context, body []byte) error {
	var ev pullRequestReviewPayload
	if err := json.Unmarshal(body, &ev); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if ev.Action != "submitted" && ev.Action != "dismissed" {
		return nil
	}
	if ev.PullRequest.User.Login != h.username {
		return nil
	}
	review := types.PRReviewEvent{
		AuthorLogin:   ev.PullRequest.User.Login,
		ReviewerLogin: ev.Review.User.Login,
		PRTitle:       ev.PullRequest.Title,
		PRURL:         ev.PullRequest.HTMLURL,
		ReviewState:   ev.Review.State,
		ReviewBody:    ev.Review.Body,
		Action:        ev.Action,
	}
	return h.notifSvc.NotifyPRReview(ctx, review)
}
