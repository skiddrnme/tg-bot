package github

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/askemblerrr/pr-review-bot/internal/types"
)

type pullRequestReviewThreadPayload struct {
	Action      string `json:"action"` // resolved / unresolved
	PullRequest struct {
		Title     string `json:"title"`
		HTMLURL   string `json:"html_url"`
		Assignees []user `json:"assignees"`
	} `json:"pull_request"`
	Thread struct {
		Comments []struct {
			Path    string `json:"path"`
			HTMLURL string `json:"html_url"`
		} `json:"comments"`
	} `json:"thread"`
	Sender user `json:"sender"`
}

func (h *Handler) handlePullRequestReviewThread(ctx context.Context, body []byte) error {
	var ev pullRequestReviewThreadPayload
	if err := json.Unmarshal(body, &ev); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if ev.Action != "resolved" && ev.Action != "unresolved" {
		return nil
	}

	assigned := slices.ContainsFunc(ev.PullRequest.Assignees, func(a user) bool {
		return strings.EqualFold(h.username, a.Login)
	})
	if !assigned {
		return nil
	}

	thread := types.PRReviewThreadEvent{
		AssigneeLogin: h.username,
		ActorLogin:    ev.Sender.Login,
		PRTitle:       ev.PullRequest.Title,
		PRURL:         ev.PullRequest.HTMLURL,
		Action:        ev.Action,
	}
	if len(ev.Thread.Comments) > 0 {
		thread.FilePath = ev.Thread.Comments[0].Path
		thread.ThreadURL = ev.Thread.Comments[0].HTMLURL
	}
	return h.notifSvc.NotifyPRReviewThread(ctx, thread)
}
