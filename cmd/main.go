package main

import (
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	//QueueGroupStart()
	//RequestReplyStart()
	//JsonEncodedStart()
	ChannelStart()
	js, _ := jetstream.New(nc)

	// create a consumer (this is an idempotent operation)
	cons, _ := js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
		Durable:   "foo",
		AckPolicy: jetstream.AckExplicitPolicy,
	})

	// create an ephemeral pull consumer by not providing `Durable`
	ephemeral, _ := js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
		AckPolicy: jetstream.AckExplicitPolicy,
	})

	// get consumer handle
	cons, _ = js.Consumer(ctx, "ORDERS", "foo")
}
