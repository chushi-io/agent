package main

import (
	"flag"
	"fmt"
	"github.com/chushi-io/agent/agent"
	"github.com/chushi-io/agent/driver"
	"github.com/chushi-io/agent/internal/auth"
	"github.com/dghubble/sling"
	"github.com/docker/docker/client"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start the manager process",
	Run:   runManager,
}

func init() {
	managerCmd.Flags().String("agent-id", "", "ID of the agent")
	managerCmd.Flags().String("grpc-address", ":8082", "Address to bind the GRPC server to")
	managerCmd.Flags().String("server-url", "https://chushi.io/api/v1", "Chushi Server URL")
	managerCmd.Flags().String("driver", "kubernetes", "Driver to execute runners with")

	_ = managerCmd.MarkFlagRequired("agent-id")

	mainCmd.AddCommand(managerCmd)
}

func runManager(cmd *cobra.Command, args []string) {
	agentId, _ := cmd.Flags().GetString("agent-id")
	grpcAddress, _ := cmd.Flags().GetString("grpc-address")
	serverUrl, _ := cmd.Flags().GetString("server-url")
	driverType, _ := cmd.Flags().GetString("driver")

	tfeConfig := &tfe.Config{
		Address:           serverUrl,
		BasePath:          "/api/v1",
		Token:             os.Getenv("CHUSHI_TOKEN"),
		RetryServerErrors: true,
	}

	tfeClient, err := tfe.NewClient(tfeConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, _ := zap.NewDevelopment()

	opts := []func(a *agent.Agent){
		agent.WithAgentId(agentId),
		agent.WithLogger(logger),
		agent.WithSdk(tfeClient),
		agent.WithOrganizationId(os.Getenv("ORGANIZATION_ID")),
		agent.WithChushiClient(sling.New().Base(serverUrl).Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TFE_TOKEN")))),
		agent.WithAuthorizer(auth.New(auth.NewMemoryStore())),
	}

	var drv driver.Driver
	switch driverType {
	case "kubernetes":
		configFile := flag.String("kubeconfig", "", "Kubernetes configuration location")
		kubeClient, err := getKubeClient(*configFile)
		if err != nil {
			logger.Fatal(err.Error())
		}
		opts = append(opts, agent.WithDriver(driver.Kubernetes{
			Client: kubeClient,
		}))
	case "docker":
		cli, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			logger.Fatal(err.Error())
		}
		drv = driver.Docker{Client: cli, Sdk: tfeClient}
	default:
		drv = driver.NewInlineRunner(logger, grpcAddress, tfeClient)
	}

	opts = append(opts, agent.WithDriver(drv))

	runnerImage := flag.String("runner-image", "", "Image for runner")
	pullPolicy := flag.String("runner-image-pull-policy", "", "Image Pull Policy")
	if runnerImage != nil || pullPolicy != nil {
		opts = append(opts, agent.WithRunnerImage(*runnerImage, *pullPolicy))
	}

	ag := agent.New(opts...)

	go func() {
		if err := ag.Grpc(grpcAddress); err != nil {
			logger.Fatal(
				"failed running grpc server",
				zap.Error(err),
			)
		}
	}()
	if err := ag.Run(os.Getenv("CHUSHI_TOKEN")); err != nil {
		logger.Fatal("Failed", zap.Error(err))
	}
}

func getKubeClient(configFile string) (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		if configFile == "" {
			configFile = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		}
		config, err = clientcmd.BuildConfigFromFlags("", configFile)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}
