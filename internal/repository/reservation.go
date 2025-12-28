package repository

import (
	"time"

	"gorm.io/gorm"
)

type ReservationRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewReservationRepository(db *gorm.DB, timeout time.Duration) *ReservationRepository {
	return &ReservationRepository{
		db:      db,
		timeout: timeout,
	}
}
