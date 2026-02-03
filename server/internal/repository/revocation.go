package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"gorm.io/gorm"
)

type Revocation struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewRevocation(db *gorm.DB, timeout time.Duration) *Revocation {
	return &Revocation{
		db:      db,
		timeout: timeout,
	}
}

// Get a revocation for a given JTI
func (r *Revocation) GetByJTI(ctx context.Context, jti string) (*model.Revocation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var revocation model.Revocation
	if err := r.db.WithContext(ctx).Where("jti = ?", jti).First(&revocation).Error; err != nil {
		return nil, err
	}
	return &revocation, nil
}

// Create a revocation
func (r *Revocation) Create(ctx context.Context, revocation *model.Revocation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	revocation.RevokedAt = time.Now().UTC()

	return r.db.WithContext(ctx).Create(revocation).Error
}
