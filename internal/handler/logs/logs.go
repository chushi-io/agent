package logs

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

func New(client *sling.Sling) agentv1connect.LogsHandler {
	return &handler{client}
}

func (h handler) StreamLogs(
	ctx context.Context,
	request *connect.Request[agentv1.StreamLogsRequest],
) (*connect.Response[agentv1.StreamLogsResponse], error) {
	return nil, nil
}

func (h handler) UploadLogs(
	ctx context.Context,
	req *connect.Request[agentv1.UploadLogsRequest],
) (*connect.Response[agentv1.UploadLogsResponse], error) {
	return nil, nil
}
