package repository

import (
	"time"

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
