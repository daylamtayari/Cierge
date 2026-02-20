package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

			restaurantNames := make(map[uuid.UUID]string)
			for _, job := range jobs {
				if _, ok := restaurantNames[job.RestaurantID]; !ok {
					restaurant, err := client.GetRestaurant(job.RestaurantID)
					if err != nil {
						restaurantNames[job.RestaurantID] = job.RestaurantID.String()
					} else {
						restaurantNames[job.RestaurantID] = restaurant.Name
					}
				}
			}

			jt := table.NewWriter()
			jt.SetStyle(table.StyleRounded)
			jt.SetColumnConfigs([]table.ColumnConfig{
				{Number: 5, Align: text.AlignLeft},
			})
			jt.AppendHeader(table.Row{"ID", "Platform", "Restaurant", "Scheduled At", "Reservation Date", "Party Size", "Preferred Times", "Status", "Reserved Time"})

			for _, job := range jobs {
				status := formatJobStatus(job.Status)

				reservedTime := ""
				if job.ReservedTime != nil {
					reservedTime = job.ReservedTime.Format("02 Jan 2006 15:04")
				}

				reservationDate, _ := time.Parse("2006-01-02", job.ReservationDate)

				jt.AppendRow(table.Row{
					job.ID,
					cases.Title(language.Und).String(job.Platform),
					restaurantNames[job.RestaurantID],
					job.ScheduledAt.Local().Format("02 Jan 2006 at 15:04 MST"),
					reservationDate.Format("02 January 2006"),
					job.PartySize,
					job.PreferredTimes,
					status,
					reservedTime,
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
