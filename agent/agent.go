package agent

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/chushi-io/agent/driver"
	"github.com/hashicorp/go-tfe"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	v1 "k8s.io/api/core/v1"
	"net"
	"net/http"
	"time"
)

type Agent struct {
	id                    string
	grpcUrl               string
	runnerImage           string
	runnerImagePullPolicy string
	logger                *zap.Logger
	organizationId        string
	sdk                   *tfe.Client
	driver                driver.Driver
}

func New(options ...func(*Agent)) *Agent {
	agent := &Agent{}
	for _, o := range options {
		o(agent)
	}
	return agent
}

func WithSdk(sdk *tfe.Client) func(agent *Agent) {
	return func(agent *Agent) {
		agent.sdk = sdk
	}
}

func WithDriver(drv driver.Driver) func(agent *Agent) {
	return func(agent *Agent) {
		agent.driver = drv
	}
}

func WithAgentId(agentId string) func(agent *Agent) {
	return func(agent *Agent) {
		agent.id = agentId
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

func (a *Agent) Run(token string) error {
	for {
		runQueue, err := a.sdk.Organizations.ReadRunQueue(context.TODO(), a.organizationId, tfe.ReadRunQueueOptions{})
		if err != nil {
			return err
		}
		for _, run := range runQueue.Items {
			if err := a.handle(run, token); err != nil {
				a.logger.Error(err.Error())
				continue
			}
			a.logger.Info("Run completed", zap.String("run.id", run.ID))
		}
		time.Sleep(time.Second * 2)
	}
}

func (a *Agent) handle(run *tfe.Run, token string) error {

	a.logger.Debug("getting workspace data")
	ws, err := a.sdk.Workspaces.Read(context.TODO(), a.organizationId, run.Workspace.ID)
	if err != nil {
		return err
	}

	// TODO: Should we just kick off the job, and let the
	// runner itself just fail if its locked?
	if ws.Locked {
		return errors.New("workspace is already locked")
	}
	//
	//a.logger.Debug("generating token for runner")
	//token, err := a.generateToken(ws.Msg.Workspace.Id, run.Id, a.organizationId)
	//if err != nil {
	//	return err
	//}

	a.logger.Debug("updating run status")
	//if _, err := a.runsClient.Update(context.TODO(), connect.NewRequest(&apiv1.UpdateRunRequest{
	//	Id:     run.Id,
	//	Status: types.RunStatusRunning,
	//})); err != nil {
	//	return err
	//}

	a.logger.Debug("getting configuration version")
	confVersion, err := a.sdk.ConfigurationVersions.Read(context.TODO(), run.ConfigurationVersion.ID)
	if err != nil {
		return err
	}
	//creds, err := a.wsClient.GetVcsConnection(context.TODO(), connect.NewRequest(&apiv1.GetVcsConnectionRequest{
	//	WorkspaceId:  ws.Msg.Workspace.Id,
	//	ConnectionId: ws.Msg.Workspace.Vcs.ConnectionId,
	//}))
	if err != nil {
		return err
	}

	a.logger.Debug("getting variables")
	//variables, err := a.sdk.Variables.List(context.TODO(), ws.ID, &tfe.VariableListOptions{})
	//if err != nil {
	//	return err
	//}

	// Build a job spec
	job := driver.NewJob(&driver.JobSpec{
		OrganizationId: a.organizationId,
		Image:          a.getRunnerImage(),
		Run:            run,
		Workspace:      ws,
		Token:          token,
		ConfigVersion:  confVersion,
		//Variables:     variables.Items,
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

	//if job.Status.State != "succeeded" {
	//	if _, err := a.runsClient.Update(context.TODO(), connect.NewRequest(&apiv1.UpdateRunRequest{
	//		Id:     run.Id,
	//		Status: types.RunStatusFailed,
	//	})); err != nil {
	//		return err
	//	}
	//	return errors.New("workspace failed")
	//}

	a.logger.Debug("updating run as completed")
	// Lastly, post updates back to the run
	//_, err = a.runsClient.Update(context.TODO(), connect.NewRequest(&apiv1.UpdateRunRequest{
	//	Id:     run.Id,
	//	Status: types.RunStatusCompleted,
	//}))
	return err
}

func (a *Agent) getRunnerImage() string {
	if a.runnerImage != "" {
		return a.runnerImage
	}
	return "ghcr.io/chushi-io/chushi:latest"
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
