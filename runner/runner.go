package runner

import (
	"connectrpc.com/connect"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	agentv1 "github.com/chushi-io/agent/gen/agent/v1"
	"github.com/chushi-io/agent/gen/agent/v1/agentv1connect"
	"github.com/chushi-io/agent/internal/auth"
	"github.com/chushi-io/agent/runner/installer"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-exec/tfexec"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
)

type Runner struct {
	logger           *zap.Logger
	grpcUrl          string
	workingDirectory string
	version          string
	operation        string
	token            string
	runId            string
	isSpeculative    bool
	backendToken     string

	writer io.Writer
}

func New(options ...func(*Runner)) *Runner {
	runner := &Runner{
		isSpeculative: false,
	}
	for _, o := range options {
		o(runner)
	}
	return runner
}

func WithLogger(logger *zap.Logger) func(*Runner) {
	return func(runner *Runner) {
		runner.logger = logger
	}
}

func WithBackendToken(backendToken string) func(*Runner) {
	return func(runner *Runner) {
		runner.backendToken = backendToken
	}
}

func WithGrpc(grpcUrl string, token string) func(*Runner) {
	return func(runner *Runner) {
		runner.grpcUrl = grpcUrl
		runner.token = token
	}
}

func WithWorkingDirectory(workingDirectory string) func(runner *Runner) {
	return func(runner *Runner) {
		runner.workingDirectory = workingDirectory
	}
}

func WithVersion(version string) func(runner *Runner) {
	return func(runner *Runner) {
		runner.version = version
	}
}

func WithOperation(operation string) func(runner *Runner) {
	return func(runner *Runner) {
		runner.operation = operation
	}
}

func WithRunId(runId string) func(runner *Runner) {
	return func(runner *Runner) {
		runner.runId = runId
	}
}

func (r *Runner) Run(ctx context.Context, out io.Writer) error {
	interceptors := connect.WithInterceptors(
		interceptor(r.runId, r.token),
	)

	adapter := newLogAdapter(
		agentv1connect.NewLogsClient(
			newInsecureClient(),
			r.grpcUrl,
			connect.WithGRPC(),
			interceptors,
		),
		r.runId,
	)
	r.writer = io.MultiWriter(adapter, out)

	r.logger.Info("installing tofu", zap.String("version", r.version))
	ver, err := version.NewVersion(r.version)
	if err != nil {
		return err
	}
	install, err := installer.Install(ver, r.workingDirectory, r.logger)
	if err != nil {
		return err
	}

	tf, err := tfexec.NewTerraform(r.workingDirectory, install)
	if err != nil {
		return err
	}

	// Copy our token to the filesystem
	pwd, _ := os.Getwd()
	err = os.WriteFile(filepath.Join(pwd, ".terraformrc"), []byte(fmt.Sprintf(`
credentials "caring-foxhound-whole.ngrok-free.app" {
  token = "%s"
}
`, r.backendToken)), 0644)
	defer os.Remove(filepath.Join(pwd, ".terraformrc"))
	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stdout)
	tf.SetEnv(map[string]string{
		"TF_FORCE_LOCAL_BACKEND": "1",
	})
	r.logger.Info("intializing tofu")
	err = tf.Init(ctx, tfexec.Upgrade(false))
	if err != nil {
		return err
	}

	r.logger.Info("tofu initialized")
	var hasChanges bool

	switch r.operation {
	case "plan":
		args := []tfexec.PlanOption{
			tfexec.Out("tfplan"),
		}
		if r.isSpeculative {
			args = append(args, tfexec.Lock(false))
		}
		r.logger.Info("Running plan")
		hasChanges, err = tf.PlanJSON(ctx, r.writer, args...)
	case "apply":
		r.logger.Info("Starting apply")
		err = tf.ApplyJSON(ctx, r.writer)
	case "destroy":
		r.logger.Info("Starting destroy")
		err = tf.DestroyJSON(ctx, r.writer)
	case "refresh_only":
	case "empty_apply":
	default:
		err = errors.New("command not found")
	}

	if err != nil {
		return err
	}

	if err = adapter.Flush(); err != nil {
		r.logger.Warn(err.Error())
	}

	if r.operation == "plan" && hasChanges {

		data, err := os.ReadFile(filepath.Join(r.workingDirectory, "tfplan"))
		if err != nil {
			return err
		}

		if err := r.uploadPlan(data); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) uploadPlan(p []byte) error {
	planClient := agentv1connect.NewPlansClient(
		newInsecureClient(),
		r.grpcUrl,
		connect.WithGRPC(),
	)
	_, err := planClient.UploadPlan(context.TODO(), connect.NewRequest(&agentv1.UploadPlanRequest{
		Content: base64.StdEncoding.EncodeToString(p),
		RunId:   r.runId,
	}))
	return err
}

func interceptor(runId string, token string) connect.UnaryInterceptorFunc {
	int := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			req.Header().Set(auth.TokenHeader, token)
			req.Header().Set(auth.RunIdHeader, runId)
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(int)
}
