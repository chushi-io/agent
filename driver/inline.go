package driver

import (
	"context"
	"github.com/chushi-io/agent/runner"
	"go.uber.org/zap"
	"os"
)

type Inline struct {
	Logger  *zap.Logger
	GrpcUrl string
}

func NewInlineRunner(logger *zap.Logger, grpcUrl string) Inline {
	return Inline{
		Logger:  logger,
		GrpcUrl: grpcUrl,
	}
}

func (i Inline) Start(job *Job) (*Job, error) {
	return job, nil
}

func (i Inline) Wait(job *Job) (*Job, error) {
	rnr := runner.New(
		runner.WithLogger(i.Logger),
		runner.WithGrpc(i.GrpcUrl),
		runner.WithWorkingDirectory(job.Spec.Workspace.WorkingDirectory),
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
	// Noop
	return nil
}
