package main

import (
	"fmt"
	"github.com/chushi-io/agent/adapter"
	"github.com/chushi-io/agent/agent"
	"github.com/chushi-io/agent/driver"
	"github.com/chushi-io/agent/internal/auth"
	"github.com/dghubble/sling"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"os"
)

// Note: The sidekiq executor is only meant to be used for
// local development. This should *not* be ran in production
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Run the sidekiq based execution for development",
	Run:   runSidekiq,
}

func init() {
	devCmd.Flags().String("redis-url", "localhost:6379/0", "URL for redis")
	devCmd.Flags().String("queue", "operations", "Sidekiq queue to process")
	devCmd.Flags().String("server-url", "https://chushi.io/api/v1", "Chushi Server URL")
}

func runSidekiq(cmd *cobra.Command, args []string) {
	serverUrl, _ := cmd.Flags().GetString("server-url")
	redisUrl, _ := cmd.Flags().GetString("redis-url")
	queue, _ := cmd.Flags().GetString("queue")

	tfeClient, err := tfe.NewClient(&tfe.Config{
		Address:           serverUrl,
		BasePath:          "/api/v1",
		Token:             os.Getenv("CHUSHI_TOKEN"),
		RetryServerErrors: true,
	})

	logger, _ := zap.NewDevelopment()
	drv := driver.NewInlineRunner(logger, ":8082", tfeClient)
	ag := agent.New(
		agent.WithAgentId(""),
		agent.WithLogger(logger),
		agent.WithSdk(tfeClient),
		agent.WithOrganizationId(""),
		agent.WithChushiClient(sling.New().Base(serverUrl).Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TFE_TOKEN")))),
		agent.WithAuthorizer(auth.New(auth.NewMemoryStore())),
		agent.WithDriver(drv),
		agent.WithAdapter(adapter.Sidekiq{
			Queue:    queue,
			RedisUrl: redisUrl,
			Sdk:      tfeClient,
			Logger:   logger,
		}),
	)

	if err != nil {
		log.Fatal(err)
	}

	if err := ag.Run(); err != nil {
		logger.Fatal("Failed", zap.Error(err))
	}
}
