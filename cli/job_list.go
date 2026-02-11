package main

import (
	"fmt"

	"github.com/daylamtayari/cierge/api"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	upcomingOnly bool

	jobListCmd = &cobra.Command{
		Use:   "list",
		Short: "List jobs",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := api.NewClient(nil, cfg.HostURL, cfg.ApiKey)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to create API client")
			}

			jobs, err := client.GetJobs(upcomingOnly)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to retrieve jobs")
			}

			jt := table.NewWriter()
			jt.AppendHeader(table.Row{"ID", "Platform", "Scheduled Date", "Reservation Date", "Party Size", "Preferred Times", "Status", "Reserved Time", "Confirmation"})

			for _, job := range jobs {
				var status string
				switch job.Status {
				case api.JobStatusCreated:
					status = "Created"
				case api.JobStatusScheduled:
					status = color.BlueString("Scheduled")
				case api.JobStatusSuccess:
					status = color.GreenString("Succeeded")
				case api.JobStatusFailed:
					status = color.RedString("Failed")
				case api.JobStatusCancelled:
					status = color.YellowString("Cancelled")
				default:
					status = string(job.Status)
				}

				reservedTime := ""
				if job.ReservedTime != nil {
					reservedTime = job.ReservedTime.Format("15:04")
				}

				confirmation := ""
				if job.Confirmation != nil {
					confirmation = *job.Confirmation
				}

				jt.AppendRow(table.Row{
					job.ID,
					job.Platform,
					job.ScheduledAt.Local(),
					job.ReservationDate.Format("2006-01-02"),
					job.PartySize,
					job.PreferredTimes,
					status,
					reservedTime,
					confirmation,
				})
				jt.SetStyle(table.StyleRounded)

				fmt.Print(jt.Render())
			}
		},
	}
)

func initJobListCmd() *cobra.Command {
	jobListCmd.Flags().BoolVar(&upcomingOnly, "upcoming-only", false, "Only output upcoming jobs")
	return jobListCmd
}
