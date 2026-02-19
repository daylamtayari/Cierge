package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Job struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewJob(db *gorm.DB, timeout time.Duration) *Job {
	return &Job{
		db:      db,
		timeout: timeout,
	}
}

// Gets a job with a given ID
func (r *Job) GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var job model.Job
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// Updates a job
func (r *Job) Update(ctx context.Context, job *model.Job) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Save(job).Error
}

// Sets the callback secret hash for a specified job
func (r *Job) SetCallbackSecretHash(ctx context.Context, secretHash string, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"callback_secret_hash": secretHash,
		}).Error
}

// Updates the status of a job
func (r *Job) UpdateStatus(ctx context.Context, status model.JobStatus, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.Job{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status": status,
		}).Error
}

// Create a job
func (r *Job) Create(ctx context.Context, job *model.Job) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(job).Error
}
