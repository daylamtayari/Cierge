package main

import (
	"github.com/daylamtayari/cierge/api"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage reservation jobs",
}

func initJobCmd() *cobra.Command {
	jobCmd.AddCommand(initJobCancelCmd())
	jobCmd.AddCommand(initJobCreateCmd())
	jobCmd.AddCommand(initJobGetCmd())
	jobCmd.AddCommand(initJobListCmd())
	return jobCmd
}

// Returns a coloured string representing the status of a job
func formatJobStatus(jobStatus api.JobStatus) string {
	switch jobStatus {
	case api.JobStatusCreated:
		return "Created"
	case api.JobStatusScheduled:
		return color.BlueString("Scheduled")
	case api.JobStatusSuccess:
		return color.GreenString("Succeeded")
	case api.JobStatusFailed:
		return color.RedString("Failed")
	case api.JobStatusCancelled:
		return color.YellowString("Cancelled")
	default:
		return string(jobStatus)
	}
}
