package main

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var (
	upcomingOnly bool

	jobListCmd = &cobra.Command{
		Use:   "list",
		Short: "List jobs",
		Run: func(cmd *cobra.Command, args []string) {
			client := newClient()

			jobs, err := client.GetJobs(upcomingOnly)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to retrieve jobs")
			}

			jt := table.NewWriter()
			jt.SetStyle(table.StyleRounded)
			jt.SetColumnConfigs([]table.ColumnConfig{
				{Number: 5, Align: text.AlignLeft},
			})
			jt.AppendHeader(table.Row{"ID", "Platform", "Scheduled At", "Reservation Date", "Party Size", "Preferred Times", "Status", "Reserved Time", "Confirmation"})

			for _, job := range jobs {
				status := formatJobStatus(job.Status)

				reservedTime := ""
				if job.ReservedTime != nil {
					reservedTime = job.ReservedTime.Format("15:04")
				}

				confirmation := ""
				if job.Confirmation != nil {
					confirmation = *job.Confirmation
				}

				reservationDate, _ := time.Parse("2006-01-02", job.ReservationDate)

				jt.AppendRow(table.Row{
					job.ID,
					job.Platform,
					job.ScheduledAt.Local().Format("02 Jan 2006 at 15:04 MST"),
					reservationDate.Format("02 January 2006"),
					job.PartySize,
					job.PreferredTimes,
					status,
					reservedTime,
					confirmation,
				})
			}

			fmt.Print(jt.Render() + "\n")
		},
	}
)

func initJobListCmd() *cobra.Command {
	jobListCmd.Flags().BoolVar(&upcomingOnly, "upcoming-only", false, "Only output upcoming jobs")
	return jobListCmd
}
