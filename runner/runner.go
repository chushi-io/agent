package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chushi-io/agent/runner/installer"
	"github.com/dghubble/sling"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Plan struct {
	PlanFormatVersion  string                     `json:"plan_format_version"`
	OutputChanges      map[string]*tfjson.Change  `json:"output_changes"`
	ResourceChanges    []*tfjson.ResourceChange   `json:"resource_changes"`
	ResourceDrift      []*tfjson.ResourceChange   `json:"resource_drift"`
	RelevantAttributes []tfjson.ResourceAttribute `json:"relevant_attributes"`

	ProviderFormatVersion string                            `json:"provider_format_version"`
	ProviderSchemas       map[string]*tfjson.ProviderSchema `json:"provider_schemas"`
}

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
	client           *sling.Sling

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

func WithClient(client *sling.Sling) func(runner *Runner) {
	return func(runner *Runner) {
		runner.client = client
	}
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
	logStreamer := newLogAdapter(r.client, r.runId)
	logger := io.MultiWriter(out, logStreamer)

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
	//tf.SetStdout(os.Stdout)
	//tf.SetStderr(os.Stdout)
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
		hasChanges, err = tf.PlanJSON(ctx, logger, args...)
	case "apply":
		r.logger.Info("Starting apply")
		err = tf.ApplyJSON(ctx, logger)
	case "destroy":
		r.logger.Info("Starting destroy")
		err = tf.DestroyJSON(ctx, logger)
	case "refresh_only":
	case "empty_apply":
	default:
		err = errors.New("command not found")
	}

	if err != nil {
		return err
	}

	if err := logStreamer.Flush(); err != nil {
		return err
	}

	if r.operation == "plan" && hasChanges {

		providerSchemas, err := tf.ProvidersSchema(ctx)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(filepath.Join(r.workingDirectory, "tfplan"))
		if err != nil {
			return err
		}

		if err := r.uploadPlan(data); err != nil {
			return err
		}

		plan, err := tf.ShowPlanFile(ctx, filepath.Join(r.workingDirectory, "tfplan"))
		if err != nil {
			return err
		}

		jplan := &Plan{
			ProviderSchemas:       providerSchemas.Schemas,
			ProviderFormatVersion: providerSchemas.FormatVersion,
			OutputChanges:         plan.OutputChanges,
			ResourceChanges:       plan.ResourceChanges,
			ResourceDrift:         plan.ResourceDrift,
			RelevantAttributes:    plan.RelevantAttributes,
		}

		if err := r.uploadPlanJson(jplan); err != nil {
			return err
		}

		structured, err := tf.ShowPlanFileRaw(ctx, filepath.Join(r.workingDirectory, "tfplan"))
		if err != nil {
			return err
		}
		if err := r.uploadStructuredPlan(structured); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) uploadPlan(p []byte) error {
	_, err := r.client.
		Post(fmt.Sprintf("/api/v1/plans/%s/upload", r.runId)).
		Set("Content-Type", "application/octet-stream").
		Body(strings.NewReader(string(p))).
		ReceiveSuccess(nil)
	return err
}

func (r *Runner) uploadStructuredPlan(input string) error {
	_, err := r.client.
		Post(fmt.Sprintf("/api/v1/plans/%s/upload_structured", r.runId)).
		Set("Content-Type", "application/octet-stream").
		Body(strings.NewReader(input)).
		ReceiveSuccess(nil)
	return err
}

func (r *Runner) uploadPlanJson(plan *Plan) error {
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}

	_, err = r.client.
		Post(fmt.Sprintf("/api/v1/plans/%s/upload_json", r.runId)).
		//Set("Content-Type", "application/octet-stream").
		Body(strings.NewReader(string(data))).
		ReceiveSuccess(nil)
	return err
}
