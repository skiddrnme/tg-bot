package types

type PRAssignedEvent struct {
    Assignee      string
    PRTitle       string
    PRURL         string
}

type PRReviewEvent struct {
    AuthorLogin   string  // pull_request.user.login
    ReviewerLogin string  // review.user.login
    PRTitle       string  // pull_request.title
    PRURL         string  // pull_request.html_url
    ReviewState   string  // review.state: approved / changes_requested / commented
    ReviewBody    string  // review.body (может быть пустой)
    Action        string  // submitted / dismissed
}