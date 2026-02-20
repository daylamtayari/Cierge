package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/daylamtayari/cierge/api"
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var (
	jobGetId string

	jobGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get details about a job",
		Run: func(cmd *cobra.Command, args []string) {
			client := newClient()
			var selectedJob *api.Job

			jobs, err := client.GetJobs(false)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to retrieve jobs")
			}

			if cmd.Flags().Changed("job") {
				uid, err := uuid.Parse(jobGetId)
				if err != nil {
					logger.Error().Err(err).Msgf("%q is not a valid UUID", jobGetId)
				} else {
					for _, job := range jobs {
						if job.ID == uid {
							selectedJob = &job
						}
					}
				}
			}

			if selectedJob == nil {
				var selectedJobId uuid.UUID
				options := make([]huh.Option[uuid.UUID], 0, len(jobs))
				for _, job := range jobs {
					date, _ := time.Parse("2006-01-02", job.ReservationDate)
					label := fmt.Sprintf("On %s for %s - %s", job.ScheduledAt.Format("02 Jan"), date.Format("02 Jan"), job.Status)
					options = append(options, huh.NewOption(label, job.ID))
				}
				err := runHuh(huh.NewSelect[uuid.UUID]().
					Title("Select job:").Options(options...).Value(&selectedJobId))
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for job")
				}
				for _, job := range jobs {
					if job.ID == selectedJobId {
						selectedJob = &job
					}
				}
			}

			restaurantName := selectedJob.RestaurantID.String()
			restaurant, err := client.GetRestaurant(selectedJob.RestaurantID)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to get restaurant")
			} else {
				restaurantName = restaurant.Name
			}

			jt := table.NewWriter()
			jt.SetStyle(table.StyleLight)
			jt.Style().Options.DrawBorder = false
			jt.Style().Options.SeparateColumns = false
			jt.SetColumnConfigs([]table.ColumnConfig{
				{Number: 2, WidthMax: 80},
			})
			jt.AppendRows([]table.Row{
				{"ID", selectedJob.ID.String()},
				{"Platform", strings.Title(selectedJob.Platform)},
				{"Restaurant", restaurantName},
				{"Status", formatJobStatus(selectedJob.Status)},
				{"Reservation Date", selectedJob.ReservationDate},
				{"Party Size", selectedJob.PartySize},
				{"Preferred Times", selectedJob.PreferredTimes},
				{"Scheduled At", selectedJob.ScheduledAt.Format("02 Jan 2006")},
			})

			if selectedJob.Status == api.JobStatusSuccess || selectedJob.Status == api.JobStatusFailed {
				jt.AppendRows([]table.Row{
					{"Started At", selectedJob.StartedAt.Format("2006-01-02 15:04:05")},
					{"Completed At", selectedJob.CompletedAt.Format("2006-01-02 15:04:05")},
				})
			}
			if selectedJob.Status == api.JobStatusSuccess {
				jt.AppendRows([]table.Row{
					{"Reserved Time", selectedJob.ReservedTime.Format("02 Jan 2006 15:04:05 MST")},
					{"Confirmation", *selectedJob.Confirmation},
				})
			}
			if selectedJob.Logs != nil {
				jt.AppendRow(table.Row{"Logs", *selectedJob.Logs})
			}
			jt.AppendRow(table.Row{"Created At", selectedJob.CreatedAt.Format("2006-01-02 15:04")})

			fmt.Print(jt.Render() + "\n")
		},
	}
)

func initJobGetCmd() *cobra.Command {
	jobGetCmd.Flags().StringVar(&jobGetId, "job", "", "UUID of the job to retrieve")
	return jobGetCmd
}
