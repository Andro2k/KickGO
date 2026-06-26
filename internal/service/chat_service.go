package service

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Andro2k/KickGO/internal/domain"
)

type ChatService struct {
	streamer domain.ChatStreamer
	tts      domain.TTSProvider
}

func NewChatService(s domain.ChatStreamer, t domain.TTSProvider) *ChatService {
	return &ChatService{streamer: s, tts: t}
}

func (cs *ChatService) StartListening(roomID int, ignoredUsers []string) error {
	msgChannel, err := cs.streamer.Connect(roomID)
	if err != nil {
		return err
	}

	ignoreMap := make(map[string]struct{}, len(ignoredUsers))
	for _, u := range ignoredUsers {
		ignoreMap[strings.ToLower(u)] = struct{}{}
	}

	var m runtime.MemStats

	for msg := range msgChannel {
		startProcessing := time.Now()

		if _, ignored := ignoreMap[strings.ToLower(msg.Username)]; ignored {
			continue
		}

		if cs.tts != nil && msg.Content != "" {
			_ = cs.tts.Speak(fmt.Sprintf("%s dice: %s", msg.Username, msg.Content))
		}

		elapsedMicro := time.Since(startProcessing).Microseconds()

		runtime.ReadMemStats(&m)
		ramMB := float64(m.Alloc) / 1024 / 1024

		fmt.Printf("   ⚡ [TELEMETRÍA] Latencia interna: %d µs | RAM Activa: %.2f MB | Objetos vivos: %d\n\n",
			elapsedMicro, ramMB, m.Mallocs-m.Frees)
	}

	return nil
}
