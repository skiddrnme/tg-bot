package github

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/askemblerrr/pr-review-bot/internal/types"
)

type pullRequestEvent struct {
    Action      string `json:"action"`
    Assignee    *user  `json:"assignee"`
    PullRequest struct {
        Title   string `json:"title"`
        HTMLURL string `json:"html_url"`
        Number  int    `json:"number"`
    } `json:"pull_request"`
    Repository struct {
        FullName string `json:"full_name"`
    } `json:"repository"`
}

type user struct {
    Login string `json:"login"`
}

func (h *Handler) handlePullRequest(ctx context.Context, body []byte) error {
    var ev pullRequestEvent
    if err := json.Unmarshal(body, &ev); err != nil {
        return fmt.Errorf("unmarshal: %w", err)
    }
    if ev.Action != "assigned" {
        return nil
    }
    if ev.Assignee == nil || ev.Assignee.Login != h.username {
        return nil
    }
    assigned := types.PRAssignedEvent{
        Assignee: ev.Assignee.Login,
        PRTitle: ev.PullRequest.Title,
        PRURL: ev.PullRequest.HTMLURL,
    }
    return h.notifSvc.NotifyPRAssigned(ctx, assigned)
}