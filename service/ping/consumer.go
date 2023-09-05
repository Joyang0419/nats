package ping

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

func NewConsumer(nc *nats.Conn, encodedConn *nats.EncodedConn) Consumer {
	return &consumer{nc: nc, encodedConn: encodedConn}
}

type consumer struct {
	nc          *nats.Conn
	encodedConn *nats.EncodedConn
}

func (c consumer) DecodedJson(topic string) (*nats.Subscription, error) {
	return c.encodedConn.Subscribe(topic, func(msg *RanksUpdatedMsg) {
		fmt.Println(msg)
	})
}

func (c consumer) Reply(ctx context.Context, topic, replyMsg string) (*nats.Subscription, <-chan string, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("consumer Reply ctx.Err() err: %w", err)
	}
	ch := make(chan string)
	sub, err := c.nc.Subscribe(topic, func(m *nats.Msg) {
		ch <- string(m.Data)
		if errPublish := c.nc.Publish(m.Reply, []byte(replyMsg)); errPublish != nil {
			panic(errPublish)
		}
	})
	if err != nil {
		return nil, nil, fmt.Errorf("consumer Reply nc.Publish err: %w", err)
	}

	return sub, ch, err
}

func (c consumer) QueueGroup(ctx context.Context, topic, queue string) (<-chan *RanksUpdatedMsg, error) {
	var ch = make(chan *RanksUpdatedMsg, 1)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("consumer QueueGroup ctx.Err: %w", err)
	}
	go func() {
		var data RanksUpdatedMsg
		var errUnmarshal error
		if _, err := c.nc.QueueSubscribe(topic, queue, func(msg *nats.Msg) {
			if errUnmarshal = json.Unmarshal(msg.Data, &data); errUnmarshal != nil {
				close(ch)
				panic(errUnmarshal)
			}
			ch <- &data
		}); err != nil {
			close(ch)
			panic(err)
		}
	}()

	return ch, nil
}

func (c consumer) BroadcastPing(ctx context.Context, topic string) (<-chan *RanksUpdatedMsg, error) {
	var ch = make(chan *RanksUpdatedMsg, 1)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("consumer BroadcastPing ctx.Err: %w", err)
	}

	go func() {
		var data RanksUpdatedMsg
		var errUnmarshal error
		if _, err := c.nc.Subscribe(topic, func(msg *nats.Msg) {
			if errUnmarshal = json.Unmarshal(msg.Data, &data); errUnmarshal != nil {
				close(ch)
				panic(errUnmarshal)
			}
			ch <- &data
		}); err != nil {
			close(ch)
			panic(err)
		}
	}()

	return ch, nil
}
