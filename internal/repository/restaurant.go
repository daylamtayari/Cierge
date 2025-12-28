package repository

import (
	"time"

	"gorm.io/gorm"
)

type RestaurantRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewRestaurantRepository(db *gorm.DB, timeout time.Duration) *RestaurantRepository {
	return &RestaurantRepository{
		db:      db,
		timeout: timeout,
	}
}
