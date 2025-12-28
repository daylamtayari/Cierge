package repository

import (
	"time"

	"gorm.io/gorm"
)

type FavouriteRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewFavouriteRepository(db *gorm.DB, timeout time.Duration) *FavouriteRepository {
	return &FavouriteRepository{
		db:      db,
		timeout: timeout,
	}
}
