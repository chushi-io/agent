package main

import (
	"flag"
	"github.com/chushi-io/agent/agent"
	"github.com/chushi-io/agent/driver"
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
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	managerCmd.Flags().String("agent-id", "", "ID of the agent")
	managerCmd.Flags().String("grpc-address", "", "Address to bind the GRPC server to")
	managerCmd.Flags().String("server-url", "https://chushi.io/api/v1", "Chushi Server URL")
	managerCmd.Flags().String("driver", "kubernetes", "Driver to execute runners with")
}

func runManager(cmd *cobra.Command, args []string) {
	agentId, _ := cmd.Flags().GetString("agent-id")
	grpcAddress, _ := cmd.Flags().GetString("grpc-address")
	serverUrl, _ := cmd.Flags().GetString("server-url")
	driverType, _ := cmd.Flags().GetString("driver")

	tfeConfig := &tfe.Config{
		Address:           serverUrl,
		Token:             os.Getenv("CHUSHI_TOKEN"),
		RetryServerErrors: true,
	}

	tfeClient, err := tfe.NewClient(tfeConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger := zap.L()

	opts := []func(a *agent.Agent){
		agent.WithAgentId(agentId),
		agent.WithLogger(logger),
		agent.WithSdk(tfeClient),
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
		drv = driver.NewInlineRunner(logger, grpcAddress)
	}

	opts = append(opts, agent.WithDriver(drv))

	runnerImage := flag.String("runner-image", "", "Image for runner")
	pullPolicy := flag.String("runner-image-pull-policy", "", "Image Pull Policy")
	if runnerImage != nil || pullPolicy != nil {
		opts = append(opts, agent.WithRunnerImage(*runnerImage, *pullPolicy))
	}

	ag := agent.New(opts...)
	if err := ag.Run(os.Getenv("CHUSHI_TOKEN")); err != nil {
		logger.Fatal(err.Error())
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
