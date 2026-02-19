package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Reservation struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewReservation(db *gorm.DB, timeout time.Duration) *Reservation {
	return &Reservation{
		db:      db,
		timeout: timeout,
	}
}

// Gets a reservation from a given ID
func (r *Reservation) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var reservation model.Reservation
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&reservation).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

// Gets all reservations for a given user
func (r *Reservation) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.Reservation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var reservations []*model.Reservation
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}

// Get all upcoming reservations for a given user
func (r *Reservation) GetByUserUpcoming(ctx context.Context, userID uuid.UUID) ([]*model.Reservation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var reservations []*model.Reservation
	if err := r.db.WithContext(ctx).Where("user_id = ? AND reservation_at > ?", userID, time.Now().UTC()).Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}

// Create a reservation
func (r *Reservation) Create(ctx context.Context, reservation *model.Reservation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(reservation).Error
}

// Update a reservation
func (r *Reservation) Update(ctx context.Context, reservation *model.Reservation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Save(reservation).Error
}

// Delete a reservation
func (r *Reservation) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.Reservation{}, "id = ?", id).Error
}
