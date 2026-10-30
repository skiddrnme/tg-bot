package github_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/askemblerrr/pr-review-bot/internal/types"
)

func TestPullRequestReview_Filtering(t *testing.T) {
	const body = `{"action":"%s","review":{"state":"approved","body":"LGTM","user":{"login":"reviewer"}},` +
		`"pull_request":{"title":"Fix","html_url":"https://x","user":{"login":"%s"}}}`

	tests := []struct {
		name   string
		action string
		author string
		want   *types.PRReviewEvent
	}{
		{
			name: "ревью на мой PR", action: "submitted", author: "octocat",
			want: &types.PRReviewEvent{
				AuthorLogin: "octocat", ReviewerLogin: "reviewer", PRTitle: "Fix", PRURL: "https://x",
				ReviewState: "approved", ReviewBody: "LGTM", Action: "submitted",
			},
		},
		{
			name: "ревью снято с моего PR", action: "dismissed", author: "octocat",
			want: &types.PRReviewEvent{
				AuthorLogin: "octocat", ReviewerLogin: "reviewer", PRTitle: "Fix", PRURL: "https://x",
				ReviewState: "approved", ReviewBody: "LGTM", Action: "dismissed",
			},
		},
		{name: "ревью отредактировано", action: "edited", author: "octocat"},
		{name: "чужой PR", action: "submitted", author: "alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockNotifSvc)
			if tt.want != nil {
				svc.On("NotifyPRReview", mock.Anything, *tt.want).Return(nil)
			}

			code := serve(svc, "pull_request_review", fmt.Sprintf(body, tt.action, tt.author))

			assert.Equal(t, http.StatusOK, code)
			if tt.want == nil {
				svc.AssertNotCalled(t, "NotifyPRReview", mock.Anything, mock.Anything)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestPullRequestReviewComment_Filtering(t *testing.T) {
	const body = `{"action":"%s","comment":{"body":"nit","path":"main.go","line":42,"html_url":"https://c",` +
		`"user":{"login":"%s"}},"pull_request":{"title":"Fix","html_url":"https://x","user":{"login":"%s"}}}`

	tests := []struct {
		name      string
		action    string
		commenter string
		author    string
		want      *types.PRReviewCommentEvent
	}{
		{
			name: "комментарий к моему PR", action: "created", commenter: "reviewer", author: "octocat",
			want: &types.PRReviewCommentEvent{
				AuthorLogin: "octocat", CommenterLogin: "reviewer", PRTitle: "Fix", PRURL: "https://x",
				FilePath: "main.go", Line: 42, CommentBody: "nit", CommentURL: "https://c",
			},
		},
		{name: "комментарий отредактирован", action: "edited", commenter: "reviewer", author: "octocat"},
		{name: "чужой PR", action: "created", commenter: "reviewer", author: "alice"},
		{name: "мой собственный комментарий", action: "created", commenter: "octocat", author: "octocat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockNotifSvc)
			if tt.want != nil {
				svc.On("NotifyPRReviewComment", mock.Anything, *tt.want).Return(nil)
			}

			code := serve(svc, "pull_request_review_comment", fmt.Sprintf(body, tt.action, tt.commenter, tt.author))

			assert.Equal(t, http.StatusOK, code)
			if tt.want == nil {
				svc.AssertNotCalled(t, "NotifyPRReviewComment", mock.Anything, mock.Anything)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestPullRequestReviewThread_Filtering(t *testing.T) {
	const withComment = `{"action":"%s","pull_request":{"title":"Fix","html_url":"https://x","assignees":%s},` +
		`"thread":{"comments":[{"path":"main.go","html_url":"https://t"}]},"sender":{"login":"reviewer"}}`
	const noComments = `{"action":"resolved","pull_request":{"title":"Fix","html_url":"https://x",` +
		`"assignees":[{"login":"octocat"}]},"thread":{"comments":[]},"sender":{"login":"reviewer"}}`

	full := func(action string) *types.PRReviewThreadEvent {
		return &types.PRReviewThreadEvent{
			AssigneeLogin: "octocat", ActorLogin: "reviewer", PRTitle: "Fix", PRURL: "https://x",
			Action: action, FilePath: "main.go", ThreadURL: "https://t",
		}
	}

	tests := []struct {
		name string
		body string
		want *types.PRReviewThreadEvent
	}{
		{name: "resolved, я в assignees", body: fmt.Sprintf(withComment, "resolved", `[{"login":"alice"},{"login":"octocat"}]`), want: full("resolved")},
		{name: "unresolved", body: fmt.Sprintf(withComment, "unresolved", `[{"login":"octocat"}]`), want: full("unresolved")},
		{name: "регистр логина не важен", body: fmt.Sprintf(withComment, "resolved", `[{"login":"OctoCat"}]`), want: full("resolved")},
		{name: "меня нет в assignees", body: fmt.Sprintf(withComment, "resolved", `[{"login":"alice"}]`)},
		{name: "assignees пуст", body: fmt.Sprintf(withComment, "resolved", `[]`)},
		{name: "другой action", body: fmt.Sprintf(withComment, "edited", `[{"login":"octocat"}]`)},
		{
			name: "без комментариев — файла и ссылки нет",
			body: noComments,
			want: &types.PRReviewThreadEvent{
				AssigneeLogin: "octocat", ActorLogin: "reviewer", PRTitle: "Fix", PRURL: "https://x", Action: "resolved",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockNotifSvc)
			if tt.want != nil {
				svc.On("NotifyPRReviewThread", mock.Anything, *tt.want).Return(nil)
			}

			code := serve(svc, "pull_request_review_thread", tt.body)

			assert.Equal(t, http.StatusOK, code)
			if tt.want == nil {
				svc.AssertNotCalled(t, "NotifyPRReviewThread", mock.Anything, mock.Anything)
			}
			svc.AssertExpectations(t)
		})
	}
}
