package adapter

import (
	"context"
	"github.com/hashicorp/go-tfe"
	"go.uber.org/zap"
	"time"
)

type Queue struct {
	Sdk            *tfe.Client
	OrganizationId string
	Logger         *zap.Logger
}

func (q Queue) Listen(handler runHandler) {
	for {
		runQueue, err := q.Sdk.Organizations.ReadRunQueue(context.TODO(), q.OrganizationId, tfe.ReadRunQueueOptions{})
		if err != nil {
			q.Logger.Error("Failed fetching runs", zap.Error(err))
			continue
		}

		for _, run := range runQueue.Items {
			handler(run)
		}

		time.Sleep(time.Second * 1)
	}
}
