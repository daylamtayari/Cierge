package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/daylamtayari/cierge/api"
	"github.com/daylamtayari/cierge/reservation"
	"github.com/daylamtayari/cierge/server/cloud"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrJobDNE = errors.New("job does not exist")
)

type Job struct {
	jobRepo       *repository.Job
	ptService     *PlatformToken
	tokenService  *Token
	cloudProvider cloud.Provider
	serverURL     string
}

func NewJob(jobRepo *repository.Job, ptService *PlatformToken, tokenService *Token, cloudProvider cloud.Provider, serverURL string) *Job {
	return &Job{
		jobRepo:       jobRepo,
		ptService:     ptService,
		tokenService:  tokenService,
		cloudProvider: cloudProvider,
		serverURL:     serverURL,
	}
}

// Retrieves a job from a given UUID
func (s *Job) GetByID(ctx context.Context, jobID uuid.UUID) (*model.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrJobDNE
	} else if err != nil {
		return nil, err
	}

	return job, nil
}

// Retrieves all jobs for a given user
func (s *Job) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.Job, error) {
	jobs, err := s.jobRepo.GetByUser(ctx, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return jobs, ErrJobDNE
	} else if err != nil {
		return jobs, err
	}
	return jobs, nil
}

// Updates a job record from a callback request. Updates the various fields of the Job object
// and if successful, stores the success output and creates a reservation.
func (s *Job) UpdateFromCallback(ctx context.Context, job *model.Job, callback reservation.Output) (*model.Job, error) {
	job.Callbacked = true
	completedAt := callback.StartTime.Add(callback.Duration)
	job.StartedAt = &callback.StartTime
	job.CompletedAt = &completedAt
	job.ErrorMessage = &callback.Error

	callbackJson, err := json.Marshal(callback)
	if err != nil {
		dbErr := s.jobRepo.Update(ctx, job)
		if dbErr != nil {
			return nil, dbErr
		}
		return nil, err
	}
	callbackLog := string(callbackJson)
	job.Logs = &callbackLog

	if callback.Success {
		job.Status = model.JobStatusSuccess
		job.ReservedTime = &callback.ReservationTime

		platformConfirmation, err := json.Marshal(callback.PlatformConfirmation)
		if err != nil {
			dbErr := s.jobRepo.Update(ctx, job)
			if dbErr != nil {
				return nil, dbErr
			}
			return nil, err
		}
		confirmation := string(platformConfirmation)
		job.Confirmation = &confirmation
	} else {
		job.Status = model.JobStatusFailed
	}

	return job, s.jobRepo.Update(ctx, job)
}

// Update the status of a job
func (s *Job) UpdateStatus(ctx context.Context, status model.JobStatus, id uuid.UUID) error {
	return s.jobRepo.UpdateStatus(ctx, status, id)
}

// Create a new job and returns the job and an error that is nil if successful
func (s *Job) Create(ctx context.Context, jobCreationRequest *api.JobCreationRequest, restaurant *model.Restaurant, dropConfig *model.DropConfig) (*model.Job, error) {
	reservationDate, _ := time.Parse("2006-01-02", jobCreationRequest.ReservationDate)
	scheduledAtDate := reservationDate.Add(-time.Duration(dropConfig.DaysInAdvance) * 24 * time.Hour)
	scheduledAtTime, _ := time.Parse("15:04", dropConfig.DropTime)
	scheduledAtLoc := time.UTC
	if restaurant.Timezone != nil {
		scheduledAtLoc = restaurant.Timezone.Location
	}
	scheduledAt := time.Date(scheduledAtDate.Year(), scheduledAtDate.Month(), scheduledAtDate.Day(), scheduledAtTime.Hour(), scheduledAtTime.Minute(), 0, 0, scheduledAtLoc)

	job := model.Job{
		UserID:          appctx.UserID(ctx),
		RestaurantID:    restaurant.ID,
		Platform:        restaurant.Platform,
		ReservationDate: model.DateString(jobCreationRequest.ReservationDate),
		PartySize:       jobCreationRequest.PartySize,
		PreferredTimes:  jobCreationRequest.PreferredTimes,
		ScheduledAt:     scheduledAt,
		DropConfigID:    jobCreationRequest.DropConfigID,
		Status:          model.JobStatusCreated,
	}

	return &job, s.jobRepo.Create(ctx, &job)
}

// Schedule a job and return an error if unsuccessful
// Includes getting the platform token and generating and
// encrypting the callback secret and scheduling the job
func (s *Job) Schedule(ctx context.Context, job *model.Job, restaurant *model.Restaurant) error {
	platformToken, err := s.ptService.GetByUserAndPlatform(ctx, job.UserID, job.Platform)
	if err != nil {
		return err
	}
	callbackSecret, err := s.tokenService.GenerateCallbackSecret(ctx, job.ID)
	if err != nil {
		return err
	}
	encryptedCallbackSecret, err := s.cloudProvider.EncryptData(ctx, callbackSecret)
	if err != nil {
		return err
	}
	event := reservation.Event{
		JobID:                   job.ID,
		Platform:                job.Platform,
		PlatformVenueId:         restaurant.PlatformID,
		EncryptedToken:          platformToken.EncryptedToken,
		EncryptedCallbackSecret: encryptedCallbackSecret,
		ReservationDate:         string(job.ReservationDate),
		PartySize:               job.PartySize,
		PreferredTimes:          job.PreferredTimes,
		DropTime:                job.ScheduledAt,
		ServerEndpoint:          s.serverURL,
		Callback:                true,
	}
	err = s.cloudProvider.ScheduleJob(ctx, event)
	if err != nil {
		return err
	}
	return nil
}

// Cancels a job
func (s *Job) Cancel(ctx context.Context, jobId uuid.UUID) error {
	err := s.cloudProvider.CancelJob(ctx, jobId)
	if err != nil {
		return err
	}
	return nil
}
