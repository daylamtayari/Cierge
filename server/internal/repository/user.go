package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewUser(db *gorm.DB, timeout time.Duration) *User {
	return &User{
		db:      db,
		timeout: timeout,
	}
}

// Gets a user with a given ID
func (r *User) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Gets a user with a given email
func (r *User) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Gets a user with a given API key
func (r *User) GetByApiKey(ctx context.Context, apiKey string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var user model.User
	if err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Checks if a user exists with a given email
func (r *User) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Checks if a user exists with a given API key
func (r *User) ExistsByApiKey(ctx context.Context, apiKey string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("api_key = ?", apiKey).Count(&count).Error
	return count > 0, err
}

// List all users
func (r *User) List(ctx context.Context) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var users []model.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

// Get all admin users
func (r *User) GetAdmins(ctx context.Context) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var users []model.User
	err := r.db.WithContext(ctx).Where("is_admin = true").Find(&users).Error
	return users, err
}

// Create user
func (r *User) Create(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(user).Error
}

// Update user
func (r *User) Update(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Save(user).Error
}

// Update user password, including the password changed timestamp
func (r *User) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password_hash":       passwordHash,
			"password_changed_at": time.Now().UTC(),
		}).Error
}

// Update a user's email
func (r *User) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("email", email).Error
}

// Update an API key value and timestamp
func (r *User) UpdateAPIKey(ctx context.Context, id uuid.UUID, apiKey string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"api_key":              apiKey,
			"api_key_last_created": time.Now().UTC(),
		}).Error
}

// Account lockout handling
// Updates the failed login attempts counter and if lockUntil is provided, updates the locked until value
func (r *User) RecordFailedLogin(ctx context.Context, id uuid.UUID, lockUntil *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	updates := map[string]any{
		"failed_login_attempts": gorm.Expr("failed_login_attempts + 1"),
		"last_failed_login":     gorm.Expr("now()"),
	}

	if lockUntil != nil {
		updates["locked_until"] = lockUntil
	}

	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *User) RecordSuccessfulLogin(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_login_at":         time.Now().UTC(),
			"failed_login_attempts": 0,
		}).
		Error
}

func (r *User) UpdateAdminStatus(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("is_admin", isAdmin).Error
}

// Delete user
func (r *User) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}
