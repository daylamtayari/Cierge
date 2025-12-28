package repository

import (
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewUserRepository(db *gorm.DB, timeout time.Duration) *UserRepository {
	return &UserRepository{
		db:      db,
		timeout: timeout,
	}
}
