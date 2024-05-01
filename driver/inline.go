package driver

import (
	"context"
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
	dir, err := downloadConfigurationVersion(i.Client, job.Spec.Run)
	if err != nil {
		return nil, err
	}

	job.Status.Metadata["git_directory"] = dir
	return job, nil
}

func (i Inline) Wait(job *Job) (*Job, error) {
	rnr := runner.New(
		runner.WithLogger(i.Logger),
		runner.WithGrpc(i.GrpcUrl),
		runner.WithWorkingDirectory(filepath.Join(
			job.Status.Metadata["git_directory"],
			job.Spec.Workspace.WorkingDirectory,
		)),
		runner.WithVersion("1.6.2"),
		runner.WithOperation("plan"),
		runner.WithRunId(job.Spec.Run.ID),
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
