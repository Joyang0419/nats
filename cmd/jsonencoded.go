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

func JsonEncodedStart() {
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
			if err = pingProducer.EncodedJson(context.Background(), QueueGroup, ping.RanksUpdatedMsg{
				UpdatedRanks: []string{strconv.Itoa(rand.Intn(10))},
			}); err != nil {
				panic(err)
			}
		}
	}()

	pingConsumer := ping.NewConsumer(nc, c)

	go func() {
		sub, errDecodedJson := pingConsumer.DecodedJson(QueueGroup)
		if errDecodedJson != nil {
			panic(errDecodedJson)
		}
		for {
			time.Sleep(10 * time.Second)
			msg, errNextMsg := sub.NextMsg(10000000 * time.Millisecond)
			if errNextMsg != nil {
				panic(errNextMsg)
			}
			fmt.Printf("%v", msg)
		}
	}()

	<-signalCh
}
