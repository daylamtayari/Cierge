package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
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

// Gets a reservation from a given ID
func (r *ReservationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var reservation model.Reservation
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&reservation).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

// Gets all reservations for a given user
func (r *ReservationRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.Reservation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var reservations []*model.Reservation
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}

// Get all upcoming reservations for a given user
func (r *ReservationRepository) GetByUserUpcoming(ctx context.Context, userID uuid.UUID) ([]*model.Reservation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var reservations []*model.Reservation
	if err := r.db.WithContext(ctx).Where("user_id = ? AND reservation_at > ?", userID, time.Now().UTC()).Find(&reservations).Error; err != nil {
		return nil, err
	}
	return reservations, nil
}

// Create a reservation
func (r *ReservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(reservation).Error
}

// Update a reservation
func (r *ReservationRepository) Update(ctx context.Context, reservation *model.Reservation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Save(reservation).Error
}

// Delete a reservation
func (r *ReservationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.Reservation{}, "id = ?", id).Error
}
