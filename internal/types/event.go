package types

type PRAssignedEvent struct {
	Assignee string
	PRTitle  string
	PRURL    string
}

type PRReviewEvent struct {
	AuthorLogin   string // pull_request.user.login
	ReviewerLogin string // review.user.login
	PRTitle       string // pull_request.title
	PRURL         string // pull_request.html_url
	ReviewState   string // review.state: approved / changes_requested / commented
	ReviewBody    string // review.body (может быть пустой)
	Action        string // submitted / dismissed
}

type PRReviewCommentEvent struct {
	AuthorLogin    string // pull_request.user.login
	CommenterLogin string
	PRTitle        string // pull_request.title
	PRURL          string // pull_request.html_url
	FilePath       string
	Line           int
	CommentBody    string
	CommentURL     string
}

type PRReviewThreadEvent struct {
	AssigneeLogin string // кого уведомляем (совпал с pull_request.assignees)
	ActorLogin    string // sender.login — кто закрыл/переоткрыл ветку
	PRTitle       string // pull_request.title
	PRURL         string // pull_request.html_url
	Action        string // resolved / unresolved
	FilePath      string // thread.comments[0].path (может быть пустым)
	ThreadURL     string // thread.comments[0].html_url (может быть пустым)
}
