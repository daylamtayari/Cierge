package repository

import (
	"time"

	"gorm.io/gorm"
)

type Repositories struct {
	User         *UserRepository
	Token        *TokenRepository
	Restaurant   *RestaurantRepository
	DropConfig   *DropConfigRepository
	Favourite    *FavouriteRepository
	Job          *JobRepository
	Reservation  *ReservationRepository
	Notification *NotificationRepository

	db      *gorm.DB // For handlers that are not tied to any repository such as the health handler
	timeout time.Duration
}

func (r *Repositories) DB() *gorm.DB {
	return r.db
}

func (r *Repositories) Timeout() time.Duration {
	return r.timeout
}
