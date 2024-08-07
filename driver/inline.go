package driver

import (
	"context"
	"fmt"
	"github.com/chushi-io/agent/runner"
	"github.com/dghubble/sling"
	"github.com/hashicorp/go-tfe"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type Inline struct {
	Logger  *zap.Logger
	GrpcUrl string
	Client  *tfe.Client
}

func NewInlineRunner(logger *zap.Logger, grpcUrl string, client *tfe.Client) Inline {
	return Inline{
		Logger:  logger,
		GrpcUrl: grpcUrl,
		Client:  client,
	}
}

func (i Inline) Start(job *Job) (*Job, error) {
	// Create a temp directory
	fmt.Println("Starting inline job")
	dir, err := downloadConfigurationVersion(i.Client, job.Spec.Run)
	if err != nil {
		return nil, err
	}

	job.Status.Metadata["git_directory"] = dir
	return job, nil
}

func (i Inline) Wait(job *Job) (*Job, error) {
	tfeClient, err := tfe.NewClient(&tfe.Config{
		Address:           "http://localhost:3000",
		BasePath:          "/api/v1",
		Token:             os.Getenv("RUNNER_TOKEN"),
		RetryServerErrors: true,
	})
	if err != nil {
		return job, err
	}

	var operation string
	if job.Spec.Run.Status == "plan_queued" {
		operation = "plan"
	} else if job.Spec.Run.Status == "apply_queued" {
		operation = "apply"
	}

	for _, variable := range job.Spec.Variables {
		if variable.HCL == false {
			os.Setenv(fmt.Sprintf("TF_VAR_%s", variable.Key), variable.Value)
		}
	}

	os.Setenv("TF_WORKSPACE", job.Spec.Workspace.Name)

	rnr := runner.New(
		runner.WithLogger(i.Logger),
		runner.WithGrpc(i.GrpcUrl, os.Getenv("RUNNER_TOKEN")),
		runner.WithWorkingDirectory(filepath.Join(
			job.Status.Metadata["git_directory"],
			job.Spec.Workspace.WorkingDirectory,
		)),
		// TODO: Replace with appropriate version
		runner.WithVersion("1.6.2"),
		// TODO: Replace with appropriate operation
		runner.WithOperation(operation),
		runner.WithRunId(job.Spec.Run.ID),
		runner.WithBackendToken(job.Spec.Token),
		runner.WithClient(sling.New().Base("http://localhost:3000").Set("Authorization", fmt.Sprintf("Bearer %s", job.Spec.Token))),
		runner.WithSdk(tfeClient),
	)

	job.Status.State = JobStateRunning

	if err := rnr.Run(context.Background(), os.Stdout); err != nil {
		job.Status.State = JobStateFailed
		return job, err
	}
	job.Status.State = JobStateComplete
	return job, nil
}

func (i Inline) Cleanup(job *Job) error {
	for _, variable := range job.Spec.Variables {
		if variable.HCL == false {
			os.Unsetenv(fmt.Sprintf("TF_VAR_%s", variable.Key))
		}
	}

	os.Unsetenv("TF_WORKSPACE")

	if err := os.RemoveAll(job.Status.Metadata["git_directory"]); err != nil {
		i.Logger.Warn(err.Error())
	}
	return nil
}
