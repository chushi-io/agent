package runs

import (
	"connectrpc.com/connect"
	"context"
	agentv1 "github.com/chushi-io/agent/gen/agent/v1"
	"github.com/chushi-io/agent/gen/agent/v1/agentv1connect"
	"github.com/dghubble/sling"
)

type handler struct {
	httpClient *sling.Sling
}

func New(client *sling.Sling) agentv1connect.RunsHandler {
	return &handler{client}
}

func (h handler) UpdateStatus(
	ctx context.Context,
	request *connect.Request[agentv1.UpdateRunStatusRequest],
) (*connect.Response[agentv1.Run], error) {

	return nil, nil
}
