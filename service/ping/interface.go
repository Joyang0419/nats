package ping

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
)

type Producer interface {
	BroadcastRanksUpdated(ctx context.Context, topic string, msg RanksUpdatedMsg) error
	RequestReply(ctx context.Context, topic string, msg string, timeout time.Duration) (replyMsg string, err error)
	EncodedJson(ctx context.Context, topic string, msg any) error
}

type Consumer interface {
	BroadcastPing(ctx context.Context, topic string) (<-chan *RanksUpdatedMsg, error)
	QueueGroup(ctx context.Context, topic, queue string) (<-chan *RanksUpdatedMsg, error)
	Reply(ctx context.Context, topic, replyMsg string) (*nats.Subscription, <-chan string, error)
	DecodedJson(topic string) (*nats.Subscription, error)
}
