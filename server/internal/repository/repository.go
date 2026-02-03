package repository

import (
	"time"

	"gorm.io/gorm"
)

type Repositories struct {
	User          *User
	PlatformToken *PlatformToken
	Restaurant    *Restaurant
	DropConfig    *DropConfig
	Favourite     *Favourite
	Job           *Job
	Reservation   *Reservation
	Notification  *Notification
	Revocation    *Revocation

	// For handlers that are not tied to any repository such as the health handler
	db      *gorm.DB
	timeout time.Duration
}

func New(db *gorm.DB, timeout time.Duration) *Repositories {
	return &Repositories{
		User:          NewUser(db, timeout),
		PlatformToken: NewPlatformToken(db, timeout),
		Restaurant:    NewRestaurant(db, timeout),
		DropConfig:    NewDropConfig(db, timeout),
		Favourite:     NewFavourite(db, timeout),
		Job:           NewJob(db, timeout),
		Reservation:   NewReservation(db, timeout),
		Notification:  NewNotification(db, timeout),
		Revocation:    NewRevocation(db, timeout),

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
