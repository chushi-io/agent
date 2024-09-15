package main

import (
	"fmt"
	"github.com/chushi-io/agent/agent"
	"github.com/chushi-io/agent/internal/auth"
	"github.com/chushi-io/agent/internal/driver"
	"github.com/chushi-io/agent/internal/listener"
	"github.com/dghubble/sling"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"log"
	"os"
)

// RabbitMQ usage

// Note: The sidekiq executor is only meant to be used for
// local development. This should *not* be ran in production
var rabbitCmd = &cobra.Command{
	Use:   "rabbitmq",
	Short: "Run the rabbitmq based manager for default plans",
	Run:   runRabbitMq,
}

func init() {
	rabbitCmd.Flags().String("amqp-url", "amqp://guest:guest@localhost:5672/", "URL for RabbitMQ")
	rabbitCmd.Flags().String("queue", "operations", "Sidekiq queue to process")
	rabbitCmd.Flags().String("server-url", "https://chushi.io/api/v1", "Chushi Server URL")
}

func runRabbitMq(cmd *cobra.Command, args []string) {
	serverUrl, _ := cmd.Flags().GetString("server-url")
	queue, _ := cmd.Flags().GetString("queue")
	amqpUrl, _ := cmd.Flags().GetString("amqp-url")

	tfeClient, err := tfe.NewClient(&tfe.Config{
		Address:           serverUrl,
		BasePath:          "/api/v1",
		Token:             os.Getenv("CHUSHI_TOKEN"),
		RetryServerErrors: true,
	})
	logger, _ := zap.NewDevelopment()
	connection, err := amqp.Dial(amqpUrl)
	if err != nil {
		logger.Fatal("failed connecting to rabbitmq", zap.Error(err))
	}
	defer connection.Close()

	rabbitmq := listener.RabbitMQ{
		Connection: connection,
		Queue:      queue,
		Logger:     logger,
	}

	drv := driver.NewInlineRunner(logger, ":8082", tfeClient)
	ag := agent.New(
		agent.WithLogger(logger),
		agent.WithChushiClient(sling.New().Base(serverUrl).Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TFE_TOKEN")))),
		agent.WithAuthorizer(auth.New(auth.NewMemoryStore())),
		agent.WithDriver(drv),
	)

	if err != nil {
		log.Fatal(err)
	}

	if err := ag.Run(rabbitmq); err != nil {
		logger.Fatal("Failed", zap.Error(err))
	}
}
