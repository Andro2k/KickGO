package service

import (
	"strings"

	"github.com/Andro2k/KickGO/internal/domain"
)

type ChatService struct {
	streamer domain.ChatStreamer
}

func NewChatService(s domain.ChatStreamer) *ChatService {
	return &ChatService{streamer: s}
}

func (cs *ChatService) StartListening(roomID int, ignoredUsers []string) (<-chan domain.ChatMessage, error) {
	msgChannel, err := cs.streamer.Connect(roomID)
	if err != nil {
		return nil, err
	}

	ignoreMap := make(map[string]struct{}, len(ignoredUsers))
	for _, u := range ignoredUsers {
		ignoreMap[strings.ToLower(u)] = struct{}{}
	}

	out := make(chan domain.ChatMessage, 100)

	go func() {
		defer close(out)
		for msg := range msgChannel {
			if _, ignored := ignoreMap[strings.ToLower(msg.Username)]; ignored {
				continue
			}
			out <- msg
		}
	}()

	return out, nil
}
