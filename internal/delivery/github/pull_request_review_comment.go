package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/askemblerrr/pr-review-bot/internal/types"
)

type pullRequestReviewCommentPayload struct {
	Action  string `json:"action"`
	Comment struct {
		Body    string `json:"body"`
		Path    string `json:"path"`
		Line    int    `json:"line"`
		HTMLURL string `json:"html_url"`
		User    user   `json:"user"` 
	} `json:"comment"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    user   `json:"user"`
	} `json:"pull_request"`
}

func (h *Handler) handlePullRequestReviewComment(ctx context.Context, body []byte) error {
	var ev pullRequestReviewCommentPayload
	if err := json.Unmarshal(body, &ev); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if ev.Action != "created" {
		return nil
	}
	if ev.PullRequest.User.Login != h.username {
		return nil // не наш PR
	}
	if ev.Comment.User.Login == h.username {
		return nil // свой комментарий — не уведомляем
	}
	comment := types.PRReviewCommentEvent{
		AuthorLogin: ev.PullRequest.User.Login,
		CommenterLogin: ev.Comment.User.Login,
		PRTitle: ev.PullRequest.Title,
		PRURL: ev.PullRequest.HTMLURL,
		FilePath: ev.Comment.Path,
		Line: ev.Comment.Line,
		CommentBody: ev.Comment.Body,
		CommentURL: ev.Comment.HTMLURL,
	}
	return h.notifSvc.NotifyPRReviewComment(ctx, comment)
	
}
