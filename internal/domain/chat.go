package domain

type ChatMessage struct {
	ID       string
	Username string
	Content  string
	Color    string
	Badges   []string
}

type ChatStreamer interface {
	Connect(roomID int) (<-chan ChatMessage, error)
	Disconnect() error
}

type TTSProvider interface {
	Speak(text string) error
	SetVolume(level float64)
}
