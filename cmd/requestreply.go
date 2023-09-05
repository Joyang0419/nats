package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nats/service/ping"

	"github.com/nats-io/nats.go"
)

const (
	TopicHelp = "TopicHelp"
)

func RequestReplyStart() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("NATS 连接失败")
	}
	defer nc.Close()

	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal("NATS 连接失败")
	}
	defer c.Close()

	pingProducer := ping.NewProducer(nc, c)

	go func() {
		receiveMsg, errRequestReply := pingProducer.RequestReply(
			context.Background(),
			TopicHelp,
			"help me",
			100*time.Millisecond,
		)
		if errRequestReply != nil {
			panic(errRequestReply)
		}
		fmt.Println(receiveMsg)
	}()

	pingConsumer := ping.NewConsumer(nc, c)
	go func() {
		_, ch, errBroadcastPing := pingConsumer.Reply(
			context.Background(),
			TopicHelp,
			"consumer 1 can help",
		)
		if errBroadcastPing != nil {
			panic(errBroadcastPing)
		}

		for receiveStr := range ch {
			fmt.Printf("consumer 1 receive: %v\n", receiveStr)
		}
	}()

	go func() {
		_, ch, errBroadcastPing := pingConsumer.Reply(
			context.Background(),
			TopicHelp,
			"consumer 2 can help",
		)
		if errBroadcastPing != nil {
			panic(errBroadcastPing)
		}

		for receiveStr := range ch {
			fmt.Printf("consumer 2 receive: %v\n", receiveStr)
		}
	}()

	<-signalCh
}
