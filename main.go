package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Andro2k/KickGO/internal/providers/kick"
	"github.com/Andro2k/KickGO/internal/service"
)

type MockTTS struct{}

func (m *MockTTS) Speak(t string) error { fmt.Printf("   🔊 [AUDIO]: %s\n", t); return nil }
func (m *MockTTS) SetVolume(l float64)  {}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pusherClient := kick.NewKickPusherClient("us2", "32cbd69e4b950bf97679")
	rewardsClient := kick.NewKickRewardsClient()
	ttsEngine := &MockTTS{}

	chatApp := service.NewChatService(pusherClient, ttsEngine)
	rewardsApp := service.NewRewardsService(rewardsClient, 15)

	roomID := 30913450
	listaNegra := []string{"Nightbot", "BotRaro", "StreamElements"}

	fmt.Println("🚀 Motor KickGO Unificado iniciando. Presiona Ctrl+C para salir.")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := chatApp.StartListening(roomID, listaNegra); err != nil {
			log.Printf("⚠️ Hilo de Chat finalizado: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		rewardsChan := rewardsApp.StartPolling(ctx)

		for red := range rewardsChan {
			fmt.Printf("🏆 [RECOMPENSA] %s canjeó: '%s' (Input: %s)\n",
				red.Username, red.RewardTitle, red.UserInput)
			_ = ttsEngine.Speak(fmt.Sprintf("%s canjeó %s", red.Username, red.RewardTitle))
		}
	}()

	wg.Wait()
	fmt.Println("🛑 Motor cerrado limpiamente. Cero fugas de memoria.")
}
