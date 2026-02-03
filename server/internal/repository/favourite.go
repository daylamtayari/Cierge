package repository

import (
	"time"

	"gorm.io/gorm"
)

type Favourite struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewFavourite(db *gorm.DB, timeout time.Duration) *Favourite {
	return &Favourite{
		db:      db,
		timeout: timeout,
	}
}
