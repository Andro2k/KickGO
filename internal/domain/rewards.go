package domain

import "context"

type Redemption struct {
	ID          string
	UserID      int
	Username    string
	RewardTitle string
	UserInput   string
}

type RewardPoller interface {
	PollPending(ctx context.Context) ([]Redemption, error)
	AcceptBatch(ctx context.Context, ids []string) error
}
