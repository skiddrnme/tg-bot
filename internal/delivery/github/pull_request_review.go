package github

import (
	"context"
	"encoding/json"
	"fmt"
)

type PullRequestReviewEvent struct {
	Action string `json:"action"`
	Review struct {
		State   string `json:"state"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"review"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
}

func (h *Handler) handlePullRequestReview(ctx context.Context,  body []byte) error{
	var rev PullRequestReviewEvent
	if err := json.Unmarshal(body, &rev); err != nil{
		return fmt.Errorf("unmarshal: %w", err)
	}
	if rev.Action != "submitted" && rev.Action != "dismissed" {
		return nil
	}
	if rev.PullRequest.User.Login != h.username{
		return nil
	}
	return h.notifSvc.NotifyPR_Review(ctx, rev.PullRequest.User.Login, rev.Review.User.Login, rev.PullRequest.Title, rev.PullRequest.HTMLURL, rev.Review.State, rev.Review.Body, rev.Action)
}