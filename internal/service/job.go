package service

import (
	"context"
	"errors"

	"github.com/daylamtayari/cierge/internal/model"
	"github.com/daylamtayari/cierge/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrJobDNE = errors.New("job does not exist")
)

type JobService struct {
	jobRepo *repository.JobRepository
}

func NewJobService(jobRepo *repository.JobRepository) *JobService {
	return &JobService{
		jobRepo: jobRepo,
	}
}

// Retrieves a job from a given UUID
func (s *JobService) GetByID(ctx context.Context, jobID uuid.UUID) (*model.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrJobDNE
	} else if err != nil {
		return nil, err
	}

	return job, nil
}
