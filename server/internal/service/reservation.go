package service

import (
	"context"
	"errors"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrReservationDNE = errors.New("reservation does not exist")
)

type Reservation struct {
	reservationRepo *repository.Reservation
}

func NewReservation(reservationRepo *repository.Reservation) *Reservation {
	return &Reservation{
		reservationRepo: reservationRepo,
	}
}

// Retrieve a reservation from a given UUID
func (s *Reservation) GetByID(ctx context.Context, reservationID uuid.UUID) (*model.Reservation, error) {
	reservation, err := s.reservationRepo.GetByID(ctx, reservationID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReservationDNE
	} else if err != nil {
		return nil, err
	}
	return reservation, nil
}

// Retrieve all reservations for a given user
func (s *Reservation) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.Reservation, error) {
	reservations, err := s.reservationRepo.GetByUser(ctx, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReservationDNE
	} else if err != nil {
		return nil, err
	}
	return reservations, nil
}

// Retrieve all upcoming reservations for a given user
func (s *Reservation) GetByUserUpcoming(ctx context.Context, userID uuid.UUID) ([]*model.Reservation, error) {
	reservations, err := s.reservationRepo.GetByUserUpcoming(ctx, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReservationDNE
	} else if err != nil {
		return nil, err
	}
	return reservations, nil
}

// Create a reservation from a provided job
func (s *Reservation) CreateFromJob(ctx context.Context, job *model.Job) (*model.Reservation, error) {
	timezone := time.UTC
	if job.Restaurant.Timezone != nil {
		timezone = job.Restaurant.Timezone.Location
	} else { //nolint:staticcheck
		// TODO: Attempt to fetch the timezone of the restaurant, store it and use it here (sounds like a function for the restaurant service)
	}

	parsedDate, _ := time.Parse("2006-01-02", string(job.ReservationDate))

	res := model.Reservation{
		JobID:        &job.ID,
		UserID:       job.UserID,
		RestaurantID: job.RestaurantID,
		Platform:     job.Platform,
		Confirmation: job.Confirmation,
		ReservationAt: time.Date(
			parsedDate.Year(), parsedDate.Month(), parsedDate.Day(),
			job.ReservedTime.Hour(), job.ReservedTime.Minute(), job.ReservedTime.Second(), job.ReservedTime.Nanosecond(),
			timezone,
		),
		PartySize: job.PartySize,
	}

	return &res, s.reservationRepo.Create(ctx, &res)
}
