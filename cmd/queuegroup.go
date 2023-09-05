package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"nats/service/ping"

	"github.com/nats-io/nats.go"
)

const (
	QueueGroup = "QueueGroup"
)

func QueueGroupStart() {
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

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if err = pingProducer.BroadcastRanksUpdated(context.Background(), QueueGroup, ping.RanksUpdatedMsg{
				UpdatedRanks: []string{strconv.Itoa(rand.Intn(10))},
			}); err != nil {
				panic(err)
			}
		}
	}()

	pingConsumer := ping.NewConsumer(nc, c)

	go func() {
		ch, errBroadcastPing := pingConsumer.QueueGroup(context.Background(), QueueGroup, "workers")
		if errBroadcastPing != nil {
			panic(errBroadcastPing)
		}
		for ranksUpdatedMsg := range ch {
			fmt.Printf("consumer 1 receive: %v\n", ranksUpdatedMsg.UpdatedRanks)
		}
	}()

	go func() {
		ch, errBroadcastPing := pingConsumer.QueueGroup(context.Background(), QueueGroup, "workerss")
		if errBroadcastPing != nil {
			panic(errBroadcastPing)
		}
		for ranksUpdatedMsg := range ch {
			fmt.Printf("consumer 2 receive: %v\n", ranksUpdatedMsg.UpdatedRanks)
		}
	}()

	<-signalCh
}
