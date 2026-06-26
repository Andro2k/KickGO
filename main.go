package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Andro2k/KickGO/internal/providers/kick"
	"github.com/Andro2k/KickGO/internal/service"
)

type IPCMessage struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pusherClient := kick.NewKickPusherClient("us2", "32cbd69e4b950bf97679")
	rewardsClient := kick.NewKickRewardsClient()

	chatApp := service.NewChatService(pusherClient)
	rewardsApp := service.NewRewardsService(rewardsClient, 15)

	roomID := 30913450
	listaNegra := []string{"Nightbot", "BotRaro", "StreamElements"}

	log.Println("🚀 Motor KickGO listo en modo Sidecar (NDJSON IPC)")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		chatChan, err := chatApp.StartListening(roomID, listaNegra)
		if err != nil {
			log.Printf("⚠️ Hilo de Chat finalizado: %v\n", err)
			return
		}

		for msg := range chatChan {
			packet, _ := json.Marshal(IPCMessage{Event: "chat", Data: msg})
			fmt.Println(string(packet))
		}
	}()

	go func() {
		defer wg.Done()
		rewardsChan := rewardsApp.StartPolling(ctx)

		for red := range rewardsChan {
			packet, _ := json.Marshal(IPCMessage{Event: "reward", Data: red})
			fmt.Println(string(packet))
		}
	}()

	wg.Wait()
	log.Println("🛑 Motor Sidecar cerrado limpiamente.")
}
