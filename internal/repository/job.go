package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewJobRepository(db *gorm.DB, timeout time.Duration) *JobRepository {
	return &JobRepository{
		db:      db,
		timeout: timeout,
	}
}

// Gets a job with a given ID
func (r *JobRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var job model.Job
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}
