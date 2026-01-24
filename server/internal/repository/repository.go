package repository

import (
	"time"

	"gorm.io/gorm"
)

type Repositories struct {
	User          *UserRepository
	PlatformToken *PlatformTokenRepository
	Restaurant    *RestaurantRepository
	DropConfig    *DropConfigRepository
	Favourite     *FavouriteRepository
	Job           *JobRepository
	Reservation   *ReservationRepository
	Notification  *NotificationRepository
	Revocation    *RevocationRepository

	// For handlers that are not tied to any repository such as the health handler
	db      *gorm.DB
	timeout time.Duration
}

func New(db *gorm.DB, timeout time.Duration) *Repositories {
	return &Repositories{
		User:          NewUserRepository(db, timeout),
		PlatformToken: NewPlatformTokenRepository(db, timeout),
		Restaurant:    NewRestaurantRepository(db, timeout),
		DropConfig:    NewDropConfigRepository(db, timeout),
		Favourite:     NewFavouriteRepository(db, timeout),
		Job:           NewJobRepository(db, timeout),
		Reservation:   NewReservationRepository(db, timeout),
		Notification:  NewNotificationRepository(db, timeout),
		Revocation:    NewRevocationRepository(db, timeout),

		db:      db,
		timeout: timeout,
	}
}

func (r *Repositories) DB() *gorm.DB {
	return r.db
}

func (r *Repositories) Timeout() time.Duration {
	return r.timeout
}
