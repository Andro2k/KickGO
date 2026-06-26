package service

import (
	"context"
	"time"

	"github.com/Andro2k/KickGO/internal/domain"
)

type RewardsService struct {
	poller    domain.RewardPoller
	interval  time.Duration
	processed map[string]struct{}
}

func NewRewardsService(p domain.RewardPoller, intervalSeconds int) *RewardsService {
	return &RewardsService{
		poller:    p,
		interval:  time.Duration(intervalSeconds) * time.Second,
		processed: make(map[string]struct{}),
	}
}

func (rs *RewardsService) StartPolling(ctx context.Context) <-chan domain.Redemption {
	out := make(chan domain.Redemption, 50)
	ticker := time.NewTicker(rs.interval)

	go func() {
		defer close(out)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				rs.executeCycle(ctx, out)
			}
		}
	}()

	return out
}

func (rs *RewardsService) executeCycle(ctx context.Context, out chan<- domain.Redemption) {
	redemptions, err := rs.poller.PollPending(ctx)
	if err != nil {
		return
	}

	var batchToAccept []string

	for _, red := range redemptions {
		if _, exists := rs.processed[red.ID]; exists {
			continue
		}

		rs.processed[red.ID] = struct{}{}
		batchToAccept = append(batchToAccept, red.ID)
		out <- red
	}

	if len(batchToAccept) > 0 {
		_ = rs.poller.AcceptBatch(ctx, batchToAccept)
	}
}
