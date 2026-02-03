package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlatformTokenRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewPlatformTokenRepository(db *gorm.DB, timeout time.Duration) *PlatformTokenRepository {
	return &PlatformTokenRepository{
		db:      db,
		timeout: timeout,
	}
}

// Get a platform token from its ID
func (r *PlatformTokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformToken model.PlatformToken
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&platformToken).Error; err != nil {
		return nil, err
	}
	return &platformToken, nil
}

// Get all platform tokens for a given user
func (r *PlatformTokenRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformTokens []*model.PlatformToken
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&platformTokens).Error; err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Get platform token for a given user and a given platform
func (r *PlatformTokenRepository) GetByUserAndPlatform(ctx context.Context, userID uuid.UUID, platform model.Platform) (*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformToken model.PlatformToken
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Where("platform = ?", platform).First(&platformToken).Error; err != nil {
		return nil, err
	}
	return &platformToken, nil
}

// Get all tokens that are expired
func (r *PlatformTokenRepository) GetExpired(ctx context.Context) ([]*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformTokens []*model.PlatformToken
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Find(&platformTokens).Error; err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Get all tokens expiring within a given duration
func (r *PlatformTokenRepository) GetExpiringWithin(ctx context.Context, duration time.Duration) ([]*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformTokens []*model.PlatformToken
	if err := r.db.WithContext(ctx).Where("expires_at < ?", (time.Now().UTC()).Add(duration)).Find(&platformTokens).Error; err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Create platform token
func (r *PlatformTokenRepository) Create(ctx context.Context, platformToken *model.PlatformToken) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(platformToken).Error
}

// Delete platform token
func (r *PlatformTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.PlatformToken{}, "id = ?", id).Error
}

// Delete tokens for a given user and platform
func (r *PlatformTokenRepository) DeleteByUserAndPlatform(ctx context.Context, userID uuid.UUID, platform model.Platform) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.PlatformToken{}, "user_id = ? AND platform = ?", userID, platform).Error
}
