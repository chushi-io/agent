package agent

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/chushi-io/agent/internal/auth"
	"github.com/chushi-io/agent/internal/driver"
	"github.com/chushi-io/agent/internal/listener"
	"github.com/chushi-io/agent/internal/server"
	"github.com/chushi-io/agent/types"
	"github.com/dghubble/sling"
	"github.com/goxiaoy/go-eventbus"
	"github.com/hashicorp/go-tfe"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	v1 "k8s.io/api/core/v1"
	"net"
	"net/http"
)

type Agent struct {
	runnerImage           string
	runnerImagePullPolicy string
	logger                *zap.Logger
	sdkResolver           func(event *listener.Event) *tfe.Client
	chushiClient          *sling.Sling
	authorizer            *auth.Auth

	driver driver.Driver
	bus    *eventbus.EventBus
}

func New(events bool, options ...func(*Agent)) *Agent {
	ag := &Agent{
		bus: eventbus.New(),
	}
	for _, o := range options {
		o(ag)
	}

	if events {
		// Register event handlers
		eventbus.Subscribe[*PlanStartedEvent](ag.bus)(func(ctx context.Context, event *PlanStartedEvent) error {
			_, err := ag.chushiClient.Put(fmt.Sprintf("/agents/v1/plans/%s", event.Plan.ID)).BodyJSON(&types.RunStatusUpdate{
				Status: "started",
			}).ReceiveSuccess(nil)
			return err
		})
		eventbus.Subscribe[*PlanCompletedEvent](ag.bus)(func(ctx context.Context, event *PlanCompletedEvent) error {
			_, err := ag.chushiClient.Put(fmt.Sprintf("/agents/v1/plans/%s", event.Plan.ID)).BodyJSON(&types.RunStatusUpdate{
				Status: "finished",
			}).ReceiveSuccess(nil)
			return err
		})
		eventbus.Subscribe[*PlanFailedEvent](ag.bus)(func(ctx context.Context, event *PlanFailedEvent) error {
			_, err := ag.chushiClient.Put(fmt.Sprintf("/agents/v1/plans/%s", event.Plan.ID)).BodyJSON(&types.RunStatusUpdate{
				Status: "errored",
			}).ReceiveSuccess(nil)
			return err
		})
		eventbus.Subscribe[*ApplyStartedEvent](ag.bus)(func(ctx context.Context, event *ApplyStartedEvent) error {
			_, err := ag.chushiClient.Put(fmt.Sprintf("/agents/v1/applies/%s", event.Apply.ID)).BodyJSON(&types.RunStatusUpdate{
				Status: "started",
			}).ReceiveSuccess(nil)
			return err
		})
		eventbus.Subscribe[*ApplyCompletedEvent](ag.bus)(func(ctx context.Context, event *ApplyCompletedEvent) error {
			_, err := ag.chushiClient.Put(fmt.Sprintf("/agents/v1/applies/%s", event.Apply.ID)).BodyJSON(&types.RunStatusUpdate{
				Status: "finished",
			}).ReceiveSuccess(nil)
			return err
		})
		eventbus.Subscribe[*ApplyFailedEvent](ag.bus)(func(ctx context.Context, event *ApplyFailedEvent) error {
			_, err := ag.chushiClient.Put(fmt.Sprintf("/agents/v1/applies/%s", event.Apply.ID)).BodyJSON(&types.RunStatusUpdate{
				Status: "errored",
			}).ReceiveSuccess(nil)
			return err
		})
	}

	return ag
}

func WithSdkResolver(sdkResolver func(event *listener.Event) *tfe.Client) func(agent *Agent) {
	return func(agent *Agent) {
		agent.sdkResolver = sdkResolver
	}
}

func WithDriver(drv driver.Driver) func(agent *Agent) {
	return func(agent *Agent) {
		agent.driver = drv
	}
}

func WithRunnerImage(runnerImage string, pullPolicy string) func(agent *Agent) {
	return func(agent *Agent) {
		agent.runnerImage = runnerImage
		agent.runnerImagePullPolicy = pullPolicy
	}
}

func WithLogger(logger *zap.Logger) func(agent *Agent) {
	return func(agent *Agent) {
		agent.logger = logger
	}
}

func WithChushiClient(client *sling.Sling) func(agent *Agent) {
	return func(agent *Agent) {
		agent.chushiClient = client
	}
}

func WithAuthorizer(authorizer *auth.Auth) func(agent *Agent) {
	return func(agent *Agent) {
		agent.authorizer = authorizer
	}
}

func (a *Agent) Grpc(addr string) error {
	srv := server.New(a.chushiClient, a.authorizer)
	return http.ListenAndServe(addr, srv)
}

func (a *Agent) Run(ad listener.Listener) error {
	ad.Listen(func(event *listener.Event) error {
		ctx := context.Background()

		sdk := a.sdkResolver(event)

		run, err := sdk.Runs.Read(context.TODO(), event.RunId)
		if err != nil {
			return err
		}

		var operation string
		if run.Status == tfe.RunApplyQueued {
			operation = "apply"
		} else if run.Status == tfe.RunPlanQueued {
			operation = "plan"
		} else {
			return errors.New(fmt.Sprintf("invalid operation provided: %s", run.Status))
		}

		if operation == "plan" {
			eventbus.Publish[*PlanStartedEvent](a.bus)(ctx, &PlanStartedEvent{Plan: run.Plan})
		} else {
			eventbus.Publish[*ApplyStartedEvent](a.bus)(ctx, &ApplyStartedEvent{Apply: run.Apply})
		}

		a.logger.Debug("Starting run", zap.String("run", run.ID))
		if err := a.handle(run, sdk, event.OrganizationId); err != nil {
			a.logger.Error("Run failed", zap.Error(err))
			if operation == "plan" {
				eventbus.Publish[*PlanFailedEvent](a.bus)(ctx, &PlanFailedEvent{Plan: run.Plan, Error: err})
			} else {
				eventbus.Publish[*ApplyFailedEvent](a.bus)(ctx, &ApplyFailedEvent{Apply: run.Apply, Error: err})
			}
			return nil
		}
		if operation == "plan" {
			eventbus.Publish[*PlanCompletedEvent](a.bus)(ctx, &PlanCompletedEvent{Plan: run.Plan})
		} else {
			eventbus.Publish[*ApplyCompletedEvent](a.bus)(ctx, &ApplyCompletedEvent{Apply: run.Apply})
		}
		a.logger.Info("Run completed", zap.String("run.id", run.ID))
		return nil
	})
	return nil
}

func (a *Agent) handle(run *tfe.Run, sdk *tfe.Client, organizationId string) error {
	a.logger.Debug("getting workspace data")
	ws, err := sdk.Workspaces.Read(context.TODO(), organizationId, run.Workspace.ID)
	if err != nil {
		return err
	}

	// TODO: Should we just kick off the job, and let the
	// runner itself just fail if its locked?
	if ws.Locked {
		return errors.New("workspace is already locked")
	}

	// Get the TF_TOKEN for the workspace to authenticate to the backend
	type TokenResponse struct {
		Token string `json:"token"`
	}
	var tokenResponse TokenResponse
	_, err = a.chushiClient.Get(fmt.Sprintf("/agents/v1/runs/%s/token", run.ID)).ReceiveSuccess(&tokenResponse)
	if err != nil {
		return err
	}

	// Generate a token for use with the proxy
	proxyToken, err := a.authorizer.GenerateToken(run.ID)
	if err != nil {
		return err
	}

	a.logger.Debug("getting configuration version")
	confVersion, err := sdk.ConfigurationVersions.Read(context.TODO(), run.ConfigurationVersion.ID)
	if err != nil {
		return err
	}

	a.logger.Debug("getting variables")
	variables, err := sdk.Variables.List(context.TODO(), ws.ID, &tfe.VariableListOptions{})
	if err != nil {
		return err
	}

	// Build a job spec
	job := driver.NewJob(&driver.JobSpec{
		OrganizationId: organizationId,
		Image:          a.getRunnerImage(),
		Run:            run,
		Workspace:      ws,
		Token:          tokenResponse.Token,
		ConfigVersion:  confVersion,
		ProxyToken:     proxyToken,
		Variables:      variables.Items,
	})

	a.logger.Debug("starting job")
	_, err = a.driver.Start(job)

	if err != nil {
		return err
	}
	defer a.driver.Cleanup(job)

	a.logger.Debug("waiting for job completion")
	_, err = a.driver.Wait(job)
	if err != nil {
		return err
	}

	a.logger.Debug("updating run as completed")

	return nil
}

func (a *Agent) getRunnerImage() string {
	if a.runnerImage != "" {
		return a.runnerImage
	}
	return "ghcr.io/chushi-io/agent:latest"
}

func (a *Agent) getImagePullPolicy() v1.PullPolicy {
	switch a.runnerImagePullPolicy {
	case "Never":
		return v1.PullNever
	case "Always":
		return v1.PullAlways
	default:
		return v1.PullIfNotPresent
	}
}

func newInsecureClient() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
				// If you're also using this client for non-h2c traffic, you may want
				// to delegate to tls.Dial if the network isn't TCP or the addr isn't
				// in an allowlist.
				return net.Dial(network, addr)
			},
			// Don't forget timeouts!
		},
	}
}

//func (a *Agent) generateToken(workspaceId string, runId string, orgId string) (string, error) {
//	resp, err := a.authClient.GenerateRunnerToken(context.TODO(), connect.NewRequest(&apiv1.GenerateRunnerTokenRequest{
//		WorkspaceId:    workspaceId,
//		RunId:          runId,
//		OrganizationId: orgId,
//	}))
//	if err != nil {
//		return "", err
//	}
//	return resp.Msg.Token, nil
//}
