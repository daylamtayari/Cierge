package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/daylamtayari/cierge/reservation"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrJobDNE = errors.New("job does not exist")
)

type Job struct {
	jobRepo *repository.Job
}

func NewJob(jobRepo *repository.Job) *Job {
	return &Job{
		jobRepo: jobRepo,
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
