package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/daylamtayari/cierge/api"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	jobCancelIdInput []string
	jobCancelIds     []uuid.UUID

	jobCancelCmd = &cobra.Command{
		Use:   "cancel",
		Short: "Cancel reservation jobs",
		Run: func(cmd *cobra.Command, args []string) {
			client := newClient()
			userJobs, err := client.GetJobs(true)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to get a user's jobs")
			}
			if len(userJobs) == 0 {
				logger.Fatal().Msg("No upcoming jobs to cancel")
			}

			userJobSet := make(map[uuid.UUID]struct{}, len(userJobs))
			for _, job := range userJobs {
				userJobSet[job.ID] = struct{}{}
			}

			if cmd.Flags().Changed("id") {
				for _, inputId := range jobCancelIdInput {
					uid, err := uuid.Parse(inputId)
					if err != nil {
						logger.Error().Err(err).Msgf("%q is not a valid UUID", inputId)
					} else if _, ok := userJobSet[uid]; !ok {
						logger.Error().Msgf("ID %q does not correspond to any valid jobs", uid.String())
					} else {
						jobCancelIds = append(jobCancelIds, uid)
					}
				}
			}

			if len(jobCancelIds) == 0 {
				// Fetch restaurant names, preventing duplicate requests
				restaurantNames := make(map[uuid.UUID]string)
				for _, job := range userJobs {
					if _, ok := restaurantNames[job.RestaurantID]; !ok {
						restaurant, err := client.GetRestaurant(job.RestaurantID)
						if err != nil {
							restaurantNames[job.RestaurantID] = job.RestaurantID.String()
						} else {
							restaurantNames[job.RestaurantID] = restaurant.Name
						}
					}
				}

				options := make([]huh.Option[uuid.UUID], 0, len(userJobs))
				for _, job := range userJobs {
					if job.Status == api.JobStatusScheduled {
						date, _ := time.Parse("2006-01-02", job.ReservationDate)
						label := fmt.Sprintf("%s for %s, party of %d", restaurantNames[job.RestaurantID], date.Format("02 Jan 2006"), job.PartySize)
						options = append(options, huh.NewOption(label, job.ID))
					}
				}

				err := runHuh(huh.NewMultiSelect[uuid.UUID]().
					Title("Select jobs to cancel:").
					Options(options...).
					Value(&jobCancelIds))
				if err != nil {
					logger.Fatal().Err(err).Msg("Failed to prompt user for jobs to cancel")
				}
			}

			for _, id := range jobCancelIds {
				if err := client.CancelJob(id); err != nil {
					logger.Error().Err(err).Msgf("Failed to cancel job %s", id.String())
				} else {
					logger.Info().Msgf("Cancelled job %s", id.String())
				}
			}
		},
	}
)

func initJobCancelCmd() *cobra.Command {
	jobCancelCmd.Flags().StringSliceVar(&jobCancelIdInput, "id", nil, "IDs of the jobs to cancel (one or multiple)")
	return jobCancelCmd
}
