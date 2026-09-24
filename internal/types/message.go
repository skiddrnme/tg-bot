package types

type Message struct {
	Text     string
	UserID   int64
	ChatID   int64
	Username string   // telegram username (для логов)
	Command  string   // "start", "set_github", "" если не команда
	Args     []string // аргументы команды
}
