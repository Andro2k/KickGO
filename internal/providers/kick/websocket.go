package kick

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Andro2k/KickGO/internal/domain"
	"github.com/gorilla/websocket"
)

type pusherEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type KickPusherClient struct {
	cluster    string
	key        string
	conn       *websocket.Conn
	urlRegex   *regexp.Regexp
	emoteRegex *regexp.Regexp
}

func NewKickPusherClient(cluster, key string) *KickPusherClient {
	return &KickPusherClient{
		cluster:    cluster,
		key:        key,
		urlRegex:   regexp.MustCompile(`https?://\S+|www\.\S+`),
		emoteRegex: regexp.MustCompile(`\[emote:[^\]]*\]`),
	}
}

func (k *KickPusherClient) Connect(roomID int) (<-chan domain.ChatMessage, error) {
	url := fmt.Sprintf("wss://ws-%s.pusher.com/app/%s?protocol=7&client=go&version=1.0.0", k.cluster, k.key)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("fallo abriendo socket: %w", err)
	}
	k.conn = conn

	outChannel := make(chan domain.ChatMessage, 100)

	go k.readLoop(roomID, outChannel)

	return outChannel, nil
}

func (k *KickPusherClient) readLoop(roomID int, out chan<- domain.ChatMessage) {
	defer close(out)

	for {
		_, rawBytes, err := k.conn.ReadMessage()
		if err != nil {
			log.Printf("❌ [RED]: Conexión cerrada por el servidor: %v\n", err)
			return
		}

		var frame pusherEvent
		if err := json.Unmarshal(rawBytes, &frame); err != nil {
			continue
		}

		switch frame.Event {
		case "pusher:connection_established":
			log.Println("🔌 [PUSHER]: Handshake aceptado. Suscribiendo a la sala...")
			subMsg := fmt.Sprintf(`{"event":"pusher:subscribe","data":{"channel":"chatrooms.%d.v2"}}`, roomID)
			if err := k.conn.WriteMessage(websocket.TextMessage, []byte(subMsg)); err != nil {
				log.Printf("❌ [RED]: Error enviando suscripción: %v\n", err)
				return
			}

		case "App\\Events\\ChatMessageEvent":
			if msg, ok := k.parseKickMessage(frame.Data); ok {
				out <- msg
			}

		case "pusher:ping":
			_ = k.conn.WriteMessage(websocket.TextMessage, []byte(`{"event":"pusher:pong"}`))
		}
	}
}

func (k *KickPusherClient) parseKickMessage(rawJson string) (domain.ChatMessage, bool) {
	var payload struct {
		ID      string `json:"id"`
		Content string `json:"content"`
		Sender  struct {
			Username string `json:"username"`
			Identity struct {
				Color  string `json:"color"`
				Badges []struct {
					Type string `json:"type"`
				} `json:"badges"`
			} `json:"identity"`
		} `json:"sender"`
	}

	if err := json.Unmarshal([]byte(rawJson), &payload); err != nil || payload.Sender.Username == "" {
		return domain.ChatMessage{}, false
	}

	badgeSet := make(map[string]struct{})
	var uniqueBadges []string

	for _, b := range payload.Sender.Identity.Badges {
		if _, exists := badgeSet[b.Type]; !exists {
			badgeSet[b.Type] = struct{}{}
			uniqueBadges = append(uniqueBadges, b.Type)
		}
	}

	cleanText := k.urlRegex.ReplaceAllString(payload.Content, "")
	cleanText = strings.TrimSpace(k.emoteRegex.ReplaceAllString(cleanText, ""))

	return domain.ChatMessage{
		ID:       payload.ID,
		Username: payload.Sender.Username,
		Content:  cleanText,
		Color:    payload.Sender.Identity.Color,
		Badges:   uniqueBadges,
	}, true
}

func (k *KickPusherClient) Disconnect() error {
	if k.conn != nil {
		return k.conn.Close()
	}
	return nil
}
