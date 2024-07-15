package adapter

import (
	"context"
	"github.com/hashicorp/go-tfe"
	"github.com/jrallison/go-workers"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
)

type Sidekiq struct {
	Queue    string
	RedisUrl string
	Sdk      *tfe.Client
	Logger   *zap.Logger
}

func (s Sidekiq) Listen(handler runHandler) {
	workers.Configure(map[string]string{
		"server":   s.RedisUrl,
		"database": "0",
		"pool":     "10",
		"process":  strconv.Itoa(rand.Intn(10000)),
	})

	workers.Process(s.Queue, func(message *workers.Msg) {
		args, _ := message.Args().Array()
		runId := args[0]
		run, err := s.Sdk.Runs.Read(context.TODO(), runId.(string))
		if err != nil {
			s.Logger.Error("Failed getting run data", zap.Error(err))
			return
		}
		handler(run)
	}, 1)

	workers.Run()
}
