package kick

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Andro2k/KickGO/internal/domain"
)

type KickRewardsClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewKickRewardsClient() *KickRewardsClient {
	return &KickRewardsClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    "http://127.0.0.1:8090/api/proxy",
	}
}

func (c *KickRewardsClient) PollPending(ctx context.Context) ([]domain.Redemption, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/pending_rewards", c.baseURL), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("servidor local de Python no responde en el puerto 8090: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("puente local retornó status: %d", resp.StatusCode)
	}

	var raw struct {
		Data []struct {
			Reward struct {
				Title string `json:"title"`
			} `json:"reward"`
			Redemptions []struct {
				ID        string `json:"id"`
				UserInput string `json:"user_input"`
				Redeemer  struct {
					UserID   int    `json:"user_id"`
					Username string `json:"username"`
				} `json:"redeemer"`
			} `json:"redemptions"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var parsed []domain.Redemption
	for _, group := range raw.Data {
		for _, r := range group.Redemptions {
			parsed = append(parsed, domain.Redemption{
				ID:          r.ID,
				UserID:      r.Redeemer.UserID,
				Username:    r.Redeemer.Username,
				RewardTitle: group.Reward.Title,
				UserInput:   r.UserInput,
			})
		}
	}

	return parsed, nil
}

func (c *KickRewardsClient) AcceptBatch(ctx context.Context, ids []string) error {
	payload, _ := json.Marshal(map[string][]string{"ids": ids})
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accept_rewards", c.baseURL), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
