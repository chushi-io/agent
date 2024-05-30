package main

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Testing redacted JSON",
	Long: `
The Chushi runner is responsible for the actual plan / apply / destroy 
executions occuring for Chushi workspaces.'
`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	mainCmd.AddCommand(testCmd)
}
