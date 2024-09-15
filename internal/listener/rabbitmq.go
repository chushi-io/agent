package listener

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Queue      string
	Logger     *zap.Logger
}

func (r RabbitMQ) Listen(handler runHandler) {
	channel, err := r.Connection.Channel()
	if err != nil {
		panic(err)
	}
	defer channel.Close()

	msgs, err := channel.Consume(
		r.Queue, // queue
		"",      // consumer
		true,    // auto ack
		false,   // exclusive
		false,   // no local
		false,   // no wait
		nil,     //args
	)
	if err != nil {
		panic(err)
	}

	// print consumed messages from queue
	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			var event Event
			if err = json.Unmarshal(msg.Body, &event); err != nil {
				r.Logger.Error("failed unmarshaling event", zap.Error(err))
			}
			if err = handler(&event); err != nil {
				r.Logger.Error("failed handling run", zap.Error(err))
			}
		}
	}()

	r.Logger.Info("Waiting for messages...")
	<-forever
}
