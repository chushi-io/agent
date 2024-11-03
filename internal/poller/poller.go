package poller

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chushi-io/chushi-go-sdk"
	"github.com/hashicorp/go-tfe"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type Poller struct {
	Sdk       *chushi.Sdk
	Client    *tfe.Client
	ClientSet *kubernetes.Clientset
	Logger    *zap.Logger
}

func New(sdk *chushi.Sdk, clientSet *kubernetes.Clientset, tfeClient *tfe.Client, logger *zap.Logger) *Poller {
	return &Poller{Sdk: sdk, ClientSet: clientSet, Client: tfeClient, Logger: logger}
}

func (p *Poller) Poll(agentId string) error {
	for {
		p.Logger.Debug("queuing jobs")
		jobs, err := p.Sdk.Jobs.List(agentId)
		if err != nil {
			p.Logger.Error("failed processing job", zap.Error(err))
			time.Sleep(time.Second * 1)
		}

		for _, job := range jobs.Items {
			p.Logger.Debug("processing job", zap.String("job.id", job.ID))
			// Lock / ack the job
			if jobErr := p.processManager(job.ID); jobErr != nil {
				p.Logger.Error("failed processing job", zap.Error(jobErr))
				continue
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func (p *Poller) processManager(jobId string) error {
	ctx := context.TODO()
	// Lock the job
	logger := p.Logger.With(zap.String("job.id", jobId))
	logger.Debug("locking job", zap.String("job.id", jobId))
	if _, err := p.Sdk.Jobs.Lock(jobId, os.Getenv("HOSTNAME")); err != nil {
		return err
	}

	// Requery the job, which ensures that we have a locked job. This
	// is probably not 100%, but its what we're doing for now
	logger.Debug("reading job")
	job, err := p.Sdk.Jobs.Read(jobId)
	if err != nil {
		return err
	}
	// Update as running
	logger.Debug("updating job as running")
	if _, err := p.Sdk.Jobs.Update(jobId, "running"); err != nil {
		return err
	}

	// Job is locked, we can process it
	logger.Debug("starting job processing")
	if err := p.process(ctx, job); err != nil {
		logger.Error("failed processing job", zap.Error(err))
		// Update as failed
		if _, updateErr := p.Sdk.Jobs.Update(jobId, "errored"); updateErr != nil {
			logger.Error("failed setting job as errored", zap.Error(updateErr))
			return updateErr
		}
		return err
	}
	// Update as completed
	//_, err = p.Sdk.Jobs.Update(jobId, "completed")
	return err
}

func (p *Poller) process(ctx context.Context, job *chushi.Job) error {
	namespace := os.Getenv("RUNNER_NAMESPACE")
	if namespace == "" {
		namespace = "chushi"
	}
	run, err := p.Client.Runs.Read(ctx, job.Run.ID)
	if err != nil {
		return err
	}

	workspace, err := p.Client.Workspaces.Read(ctx, "chushi", run.Workspace.ID)
	if err != nil {
		return err
	}

	run.Workspace = workspace

	oidcToken, err := p.Sdk.Runs.OidcToken(run.ID)
	if err != nil {
		return err
	}

	oidcSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "run-identity-",
			Namespace:    namespace,
			Labels:       labelsForRun(run),
		},
		Data: map[string][]byte{"token": []byte(oidcToken)},
	}

	if oidcSecret, err = p.ClientSet.CoreV1().Secrets(namespace).Create(ctx, oidcSecret, metav1.CreateOptions{}); err != nil {
		return err
	}

	runToken, err := p.Sdk.Runs.Token(run.ID)
	if err != nil {
		return err
	}

	secretVars := map[string]*tfe.Variable{}
	for _, variable := range run.Workspace.Variables {
		secretVars[variable.Key] = variable
	}

	secretData, terraformTfvars := preloadSecrets(run)

	fileMappings := map[string]string{
		".terraform_environment": ".terraform/environment",
		"terraform.tfvars":       filepath.Join(run.Workspace.WorkingDirectory, "terraform.tfvars"),
	}
	configMapData := map[string]string{
		".terraform_environment": run.Workspace.Name,
		"terraform.tfvars":       terraformTfvars,
	}

	hostname := strings.Replace(
		strings.TrimPrefix(p.Sdk.Address, "https://"),
		".", "_", -1,
	)
	// TODO: This should be a run-specific token and not the agent token, but eh.
	secretData[fmt.Sprintf("TF_TOKEN_%s", hostname)] = []byte(runToken)

	runSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "run-",
			Namespace:    namespace,
			Labels:       labelsForRun(run),
		},
		Data: secretData,
	}

	if runSecret, err = p.ClientSet.CoreV1().Secrets(namespace).Create(ctx, runSecret, metav1.CreateOptions{}); err != nil {
		return err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "run-",
			Namespace:    namespace,
			Labels:       labelsForRun(run),
		},
		Data: configMapData,
	}

	if configMap, err = p.ClientSet.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return err
	}

	podSpec, err := podForRun(
		job,
		run,
		runToken,
		fileMappings,
		oidcSecret.Name,
		runSecret.Name,
		configMap.Name,
	)
	if err != nil {
		return err
	}

	if podSpec, err = p.ClientSet.CoreV1().Pods(namespace).Create(ctx, podSpec, metav1.CreateOptions{}); err != nil {
		return err
	}

	// TODO: Poll and watch the pod?
	go func() {
		//labelSelector := &metav1.LabelSelector{
		//	MatchLabels: podSpec.Labels,
		//}
		//selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		//if err != nil {
		//	p.Logger.Error("failed generating selector", zap.Error(err))
		//	return
		//}
		//p.Logger.Info(selector.String())
		watch, err := p.ClientSet.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(metav1.ObjectNameField, podSpec.Name).String(),
		})

		if err != nil {
			p.Logger.Error("failed creating watch listener", zap.Error(err))
			return
		}
		for event := range watch.ResultChan() {
			//fmt.Printf("Type: %v\n", event.Type)
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				log.Fatal("unexpected type")
			}
			p.Logger.Info(
				"pod status change",
				zap.String("job.id", job.ID),
				zap.String("pod.phase", string(pod.Status.Phase)),
				zap.String("pod.name", pod.Name),
			)
			switch pod.Status.Phase {
			case corev1.PodSucceeded:
				p.Logger.Info("updating job as completed", zap.String("job.id", job.ID))
				if _, err := p.Sdk.Jobs.Update(job.ID, "completed"); err != nil {
					p.Logger.Error("failed updating job", zap.String("job.id", job.ID))
				}
				return
			case corev1.PodFailed:
				p.Logger.Info("updating job as failed", zap.String("job.id", job.ID))
				if _, err := p.Sdk.Jobs.Update(job.ID, "errored"); err != nil {
					p.Logger.Error("failed updating job", zap.String("job.id", job.ID))
				}
				return
			}
		}
		time.Sleep(time.Second)
	}()

	return nil
}

func podForRun(
	job *chushi.Job,
	run *tfe.Run,
	token string,
	fileMappings map[string]string,
	identitySecret string,
	runSecret string,
	configMap string,
) (*corev1.Pod, error) {
	t := template.Must(template.New("init-script").Parse(`curl -Lv -H "Authorization: Bearer ${TFE_TOKEN}" {{ .Host }}/api/v2/configuration-versions/{{ .ConfigurationVersionId }}/download --output config.tgz
tar -xvf config.tgz -C /workspace

{{range $item, $path := .FileMappings}}
mkdir -p /workspace/{{ $.WorkingDirectory }}/$(dirname {{ $path }}) 
cat /configuration/{{ $item }} > /workspace/{{ $.WorkingDirectory }}/{{ $path }}
{{end}}
`))
	var doc bytes.Buffer
	if err := t.Execute(&doc, struct {
		Host                   string
		ConfigurationVersionId string
		FileMappings           map[string]string
		WorkingDirectory       string
	}{
		Host:                   os.Getenv("TFE_ADDRESS"),
		ConfigurationVersionId: run.ConfigurationVersion.ID,
		FileMappings:           fileMappings,
		WorkingDirectory:       run.Workspace.WorkingDirectory,
	}); err != nil {
		return nil, err
	}

	runnerArgs := []string{
		fmt.Sprintf("--directory=/workspace/%s", run.Workspace.WorkingDirectory),
		// TODO: Pull this from configuration somewhere
		fmt.Sprintf("--log-stream-url=%s", "http://host.minikube.internal:8080"),
		fmt.Sprintf("--run-id=%s", run.ID),
		"--debug",
	}

	if run.IsDestroy {
		runnerArgs = append(runnerArgs, "--destroy")
	}
	for linkName, linkUrl := range job.Links {
		runnerArgs = append(runnerArgs, fmt.Sprintf("--%s=%s", linkName, linkUrl))
	}

	if run.Workspace.TerraformVersion != "" {
		runnerArgs = append(runnerArgs, fmt.Sprintf("--version=%s", run.Workspace.TerraformVersion))
	}
	// Add any additional arguments
	runnerArgs = append(runnerArgs, job.Operation)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "chushi-run-",
			Namespace:    os.Getenv("RUNNER_NAMESPACE"),
			Labels:       labelsForRun(run),
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			InitContainers: []corev1.Container{{
				Name:       "runner-init",
				Image:      "alpine/curl",
				WorkingDir: filepath.Join("/workspace", run.Workspace.WorkingDirectory),
				Command: []string{
					"sh",
					"-c",
					doc.String(),
				},
				Env: []corev1.EnvVar{{
					Name:  "TFE_TOKEN",
					Value: token,
				}},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "temp",
					MountPath: "/workspace",
				}, {
					Name:      "config",
					MountPath: "/configuration",
				}},
			}},
			Containers: []corev1.Container{{
				Name:            "runner",
				Image:           os.Getenv("RUNNER_IMAGE"),
				ImagePullPolicy: corev1.PullNever,
				Args:            runnerArgs,
				WorkingDir:      filepath.Join("/workspace", run.Workspace.WorkingDirectory),
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "temp",
					MountPath: "/workspace",
				}},
				Env: []corev1.EnvVar{{
					Name:  "TF_FORCE_LOCAL_BACKEND",
					Value: "1",
				}, {
					Name:  "TFE_TOKEN",
					Value: token,
				}},
				EnvFrom: []corev1.EnvFromSource{{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: runSecret,
						},
					},
				}},
			}},
			Volumes: []corev1.Volume{{
				Name: "temp",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}, {
				Name: "identity",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: identitySecret,
					},
				},
			}, {
				Name: "secret",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: runSecret,
					},
				},
			}, {
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMap,
						},
					},
				},
			}},
		},
	}

	return pod, nil
}

func labelsForRun(run *tfe.Run) map[string]string {
	return map[string]string{
		"agent.chushi.io/run-id": run.ID,
	}
}

func preloadSecrets(run *tfe.Run) (map[string][]byte, string) {
	secretVars := map[string]*tfe.Variable{}
	terraformTfVars := map[string]string{}
	for _, variable := range run.Workspace.Variables {
		secretVars[variable.Key] = variable
	}

	generatedSecrets := map[string][]byte{}

	if val, ok := secretVars["TFC_AWS_PROVIDER_AUTH"]; ok && string(val.Value) == "true" {
		// We handle AWS authentication
		delete(secretVars, "TFC_AWS_PROVIDER_AUTH")
	}
	if val, ok := secretVars["TFC_GCP_PROVIDER_AUTH"]; ok && string(val.Value) == "true" {
		// Setup GCP authentication

		delete(secretVars, "TFC_GCP_PROVIDER_AUTH")
	}
	if val, ok := secretVars["TFC_AZURE_PROVIDER_AUTH"]; ok && string(val.Value) == "true" {
		// Setup Azure authentication

		delete(secretVars, "TFC_AZURE_PROVIDER_AUTH")
	}
	if val, ok := secretVars["TFC_KUBERNETES_PROVIDER_AUTH"]; ok && string(val.Value) == "true" {
		// Handle kubernetes authentication

		delete(secretVars, "TFC_KUBERNETES_PROVIDER_AUTH")
	}
	if val, ok := secretVars["TFC_VAULT_PROVIDER_AUTH"]; ok && string(val.Value) == "true" {
		// Handle vault authentication

		delete(secretVars, "TFC_VAULT_PROVIDER_AUTH")
	}

	tfVarsOutput := []string{}
	for tfVarKey, tfVarValue := range terraformTfVars {
		tfVarsOutput = append(tfVarsOutput, fmt.Sprintf("%s=%s", tfVarKey, tfVarValue))
	}

	for key, value := range secretVars {
		// TODO: Add support for HCL variables?
		if value.Category == tfe.CategoryEnv {
			generatedSecrets[key] = []byte(value.Value)
		} else if value.Category == tfe.CategoryTerraform {
			generatedSecrets[fmt.Sprintf("TF_VAR_%s", key)] = []byte(value.Value)
		}
	}
	return generatedSecrets, strings.Join(tfVarsOutput, "\n")
}
