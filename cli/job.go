package main

import "github.com/spf13/cobra"

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage reservation jobs",
}

func initJobCmd() *cobra.Command {
	jobCmd.AddCommand(initJobCreateCmd())
	jobCmd.AddCommand(initJobListCmd())
	return jobCmd
}
