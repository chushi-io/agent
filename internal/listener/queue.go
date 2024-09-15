package listener

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
			if err = handler(&Event{
				RunId:          run.ID,
				OrganizationId: q.OrganizationId,
				WorkspaceId:    run.Workspace.ID,
			}); err != nil {
				q.Logger.Error("failed handling run", zap.Error(err))
			}
		}

		time.Sleep(time.Second * 1)
	}
}

func (q Queue) Client() *tfe.Client {
	return q.Sdk
}
