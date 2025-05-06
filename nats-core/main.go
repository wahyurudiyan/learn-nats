package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/wahyurudiyan/nats-core/pubsub"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	ctx := context.Background()

	// Listen to operating system kill, interrupt or quit signal
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	defer stop()

	// Init logger
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Drain()

	mode := os.Getenv("MODE")
	switch mode {
	case "pubsub":
		logger.Info("[NATS] running on mode Pubsub!")

		ps := pubsub.NewPubsub(nc)
		ps.AddSubscriber(ctx, "processor.>", func(msg *nats.Msg) { // processor.fadil
			fmt.Printf("Callback-1: message from subject %s is: %s\n", msg.Subject, string(msg.Data))
		})
		ps.AddSubscriber(ctx, "processor.satu", func(msg *nats.Msg) {
			fmt.Printf("Callback-2: message from subject %s is: %s\n", msg.Subject, string(msg.Data))
		})
		ps.Listen(ctx)
		logger.Info("[NATS] application in mode Pubsub is done")
	case "request", "request-reply":
		logger.Info("[NATS] running on mode Request-Reply!")
		nc.Subscribe("say.*", func(msg *nats.Msg) {
			respond := fmt.Sprintf("did you say, %s?", msg.Subject[4:])
			msg.Respond([]byte(respond))
		})

		<-ctx.Done()
		logger.Info("[NATS] application in mode Request-Reply is done")
	case "groups", "queue-groups":

	default:
		logger.Fatal("please specify your mode")
	}
}
