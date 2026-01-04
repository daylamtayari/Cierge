package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/internal/model"
	"gorm.io/gorm"
)

type RevocationRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewRevocationRepository(db *gorm.DB, timeout time.Duration) *RevocationRepository {
	return &RevocationRepository{
		db:      db,
		timeout: timeout,
	}
}

// Get a revocation for a given JTI
func (r *RevocationRepository) GetByJTI(ctx context.Context, jti string) (*model.Revocation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var revocation model.Revocation
	if err := r.db.WithContext(ctx).Where("jti = ?", jti).First(&revocation).Error; err != nil {
		return nil, err
	}
	return &revocation, nil
}

// Create a revocation
func (r *RevocationRepository) Create(ctx context.Context, revocation *model.Revocation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(revocation).Error
}
