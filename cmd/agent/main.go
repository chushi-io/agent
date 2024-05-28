package main

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var mainCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run the Chushi agent",
}

func main() {
	cmd, _, err := mainCmd.Find(os.Args[1:])
	if err != nil || cmd == nil {
		args := append([]string{"manager"}, os.Args[1:]...)
		mainCmd.SetArgs(args)
	}

	if err := mainCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
