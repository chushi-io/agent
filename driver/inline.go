package driver

import (
	"context"
	"fmt"
	"github.com/chushi-io/agent/runner"
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
	fmt.Printf("Git Directory: %s", dir)
	return job, nil
}

func (i Inline) Wait(job *Job) (*Job, error) {
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
		runner.WithOperation("plan"),
		runner.WithRunId(job.Spec.Run.ID),
		runner.WithBackendToken(job.Spec.Token),
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
	if err := os.RemoveAll(job.Status.Metadata["git_directory"]); err != nil {
		i.Logger.Warn(err.Error())
	}
	return nil
}
