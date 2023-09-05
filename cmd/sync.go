package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

func SyncStart() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("NATS 连接失败")
	}
	defer nc.Close()

	// Simple Publisher
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			nc.Publish("foo", []byte("Hello World"))
		}
	}()
	sub, _ := nc.SubscribeSync("foo")
	go func() {
		// Simple Sync Subscriber
		for {
			m, errM := sub.NextMsg(10000 * time.Second)
			if m != nil {
				fmt.Println(string(m.Data))
			}
			fmt.Println(errM)
		}

	}()

	<-signalCh
	sub.Drain()
}
