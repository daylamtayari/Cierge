package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlatformToken struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewPlatformToken(db *gorm.DB, timeout time.Duration) *PlatformToken {
	return &PlatformToken{
		db:      db,
		timeout: timeout,
	}
}

// Get a platform token from its ID
func (r *PlatformToken) GetByID(ctx context.Context, id uuid.UUID) (*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformToken model.PlatformToken
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&platformToken).Error; err != nil {
		return nil, err
	}
	return &platformToken, nil
}

// Get all platform tokens for a given user
func (r *PlatformToken) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformTokens []*model.PlatformToken
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&platformTokens).Error; err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Get platform token for a given user and a given platform
func (r *PlatformToken) GetByUserAndPlatform(ctx context.Context, userID uuid.UUID, platform string) (*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformToken model.PlatformToken
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Where("platform = ?", platform).First(&platformToken).Error; err != nil {
		return nil, err
	}
	return &platformToken, nil
}

// Get all tokens that are expired
func (r *PlatformToken) GetExpired(ctx context.Context) ([]*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformTokens []*model.PlatformToken
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Find(&platformTokens).Error; err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Get all tokens expiring within a given duration
func (r *PlatformToken) GetExpiringWithin(ctx context.Context, duration time.Duration) ([]*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformTokens []*model.PlatformToken
	if err := r.db.WithContext(ctx).Where("expires_at < ?", (time.Now().UTC()).Add(duration)).Find(&platformTokens).Error; err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Get all tokens expirigin with a given duration that have a refresh token
func (r *PlatformToken) GetExpiringWithinWithRefresh(ctx context.Context, duration time.Duration) ([]*model.PlatformToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var platformTokens []*model.PlatformToken
	if err := r.db.WithContext(ctx).Where("expires_at < ?", (time.Now().UTC()).Add(duration)).Where("has_refresh = true").Find(&platformTokens).Error; err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Create platform token
func (r *PlatformToken) Create(ctx context.Context, platformToken *model.PlatformToken) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(platformToken).Error
}

// Replace platform token
func (r *PlatformToken) Replace(ctx context.Context, oldTokenId uuid.UUID, newToken *model.PlatformToken) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete the old token
		if err := tx.Delete(&model.PlatformToken{}, "id = ?", oldTokenId).Error; err != nil {
			return err
		}

		// Create the new token
		if err := tx.Create(newToken).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete platform token
func (r *PlatformToken) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.PlatformToken{}, "id = ?", id).Error
}

// Delete tokens for a given user and platform
func (r *PlatformToken) DeleteByUserAndPlatform(ctx context.Context, userID uuid.UUID, platform string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.PlatformToken{}, "user_id = ? AND platform = ?", userID, platform).Error
}
