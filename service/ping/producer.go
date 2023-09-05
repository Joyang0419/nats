package ping

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

func NewProducer(nc *nats.Conn, encodedConn *nats.EncodedConn) Producer {
	return &producer{nc: nc, encodedConn: encodedConn}
}

type producer struct {
	nc          *nats.Conn
	encodedConn *nats.EncodedConn
}

func (p producer) EncodedJson(ctx context.Context, topic string, msg any) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("producer EncodedJson ctx.Err err: %w", err)
	}
	return p.encodedConn.Publish(topic, msg)
}

func (p producer) RequestReply(ctx context.Context, topic string, msg string, timeout time.Duration) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("producer RequestReply ctx.Err err: %w", err)
	}

	natsMsg, err := p.nc.Request(topic, []byte(msg), timeout)
	if err != nil {
		return "", fmt.Errorf("producer RequestReply nc.Request err: %w", err)
	}

	return string(natsMsg.Data), err
}

func (p producer) BroadcastRanksUpdated(ctx context.Context, topic string, msg RanksUpdatedMsg) error {
	var (
		errCh = make(chan error)
	)
	ctx, cancelFunc := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancelFunc()

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("producer BroadcastPing nc.Publish err: %w", err)
	}

	go func() {
		errCh <- p.nc.Publish(topic, jsonBytes)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		if err != nil {
			return fmt.Errorf("producer BroadcastPing nc.Publish err: %w", err)
		}
	}

	return nil
}
