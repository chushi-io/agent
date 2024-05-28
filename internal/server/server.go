package server

import (
	"connectrpc.com/connect"
	"github.com/chushi-io/agent/gen/agent/v1/agentv1connect"
	"github.com/chushi-io/agent/internal/auth"
	"github.com/chushi-io/agent/internal/handler/logs"
	"github.com/chushi-io/agent/internal/handler/plans"
	"github.com/chushi-io/agent/internal/handler/runs"
	"github.com/dghubble/sling"
	"net/http"
)

func New(
	httpClient *sling.Sling,
	auth *auth.Auth,
) *http.ServeMux {
	interceptors := connect.WithInterceptors(auth.Interceptor())
	mux := http.NewServeMux()
	mux.Handle(agentv1connect.NewRunsHandler(runs.New(httpClient), interceptors))
	mux.Handle(agentv1connect.NewPlansHandler(plans.New(httpClient), interceptors))
	mux.Handle(agentv1connect.NewLogsHandler(logs.New(httpClient), interceptors))

	return mux
}
