package main

import (
	"flag"
	"github.com/chushi-io/agent/internal/poller"
	"github.com/chushi-io/chushi-go-sdk"
	"github.com/hashicorp/go-tfe"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
)

func main() {
	sdk, err := chushi.New(tfe.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	client, err := tfe.NewClient(tfe.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	logger, _ := zap.NewDevelopment()
	p := poller.New(sdk, clientset, client, logger)

	if err = p.Poll(os.Getenv("TFE_AGENT_POOL_ID")); err != nil {
		log.Fatal(err)
	}
}
