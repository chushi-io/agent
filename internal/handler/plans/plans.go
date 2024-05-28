package plans

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	agentv1 "github.com/chushi-io/agent/gen/agent/v1"
	"github.com/chushi-io/agent/gen/agent/v1/agentv1connect"
	"github.com/dghubble/sling"
	"strings"
)

type handler struct {
	httpClient *sling.Sling
}

func New(client *sling.Sling) agentv1connect.PlansHandler {
	return &handler{client}
}

func (h handler) UploadPlan(
	ctx context.Context,
	request *connect.Request[agentv1.UploadPlanRequest],
) (*connect.Response[agentv1.UploadPlanResponse], error) {
	runId := request.Msg.RunId
	planContent := request.Msg.Content

	// Post the plan to HTTP endpoint
	_, err := h.httpClient.Put(fmt.Sprintf("agents/v1/plans/%s", runId)).Body(strings.NewReader(planContent)).ReceiveSuccess(nil)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&agentv1.UploadPlanResponse{Success: true}), nil
}
