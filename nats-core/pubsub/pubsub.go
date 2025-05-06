package pubsub

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()

type NatsSubject string

type pubsub struct {
	conn        *nats.Conn
	subscribers map[NatsSubject]nats.MsgHandler
}

func NewPubsub(nc *nats.Conn) *pubsub {
	return &pubsub{
		conn: nc,
	}
}

func (p *pubsub) Publisher(ctx context.Context, subject string, message string) error {
	return p.conn.Publish(subject, []byte(message))
}

func (p *pubsub) AddSubscriber(ctx context.Context, subject string, callback nats.MsgHandler) {
	if p.subscribers == nil {
		p.subscribers = make(map[NatsSubject]nats.MsgHandler)
	}
	p.subscribers[NatsSubject(subject)] = callback
}

func (p *pubsub) Listen(ctx context.Context) {
	var wg sync.WaitGroup
	for subject, callback := range p.subscribers {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()

			logger.Info("[NATS] subscriber is running", zap.String("subject", string(subject)))
			sub, err := p.conn.Subscribe(string(subject), callback)
			if err != nil {
				panic(err)
			}
			defer sub.Unsubscribe()

			<-ctx.Done()
			logger.Info("[NATS] subscription done")
		}(ctx)
	}

	wg.Wait()
}
