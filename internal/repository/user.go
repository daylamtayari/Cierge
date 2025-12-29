package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewUserRepository(db *gorm.DB, timeout time.Duration) *UserRepository {
	return &UserRepository{
		db:      db,
		timeout: timeout,
	}
}

// -----------------
// Retrieval methods
// -----------------

// Gets a user with a given ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Gets a user with a given email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Checks if a user exists with a given email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// List all users
func (r *UserRepository) List(ctx context.Context) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var users []*models.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

// Get all admin users
func (r *UserRepository) GetAdmins(ctx context.Context) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var users []*models.User
	err := r.db.WithContext(ctx).Where("is_admin = true").Find(&users).Error
	return users, err
}

// Get user notification preferences
func (r *UserRepository) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*models.UserNotificationPreferences, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var notificationPreferences models.UserNotificationPreferences
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&notificationPreferences).Error; err != nil {
		return nil, err
	}
	return &notificationPreferences, nil
}

// -----------------
// Mutating methods
// -----------------

// Create user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(user).Error
}

// Update user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Save(user).Error
}

// Update user password, including the password changed timestamp
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password_hash":       passwordHash,
			"password_changed_at": time.Now().UTC(),
		}).Error
}

// Update a user's email
func (r *UserRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("email", email).Error
}

// Update an API key value
func (r *UserRepository) UpdateAPIKey(ctx context.Context, id uuid.UUID, apiKey string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("api_key", apiKey).Error
}

// Account lockout handling
// Updates the failed login attempts counter and if lockUntil is provided, updates the locked until value
func (r *UserRepository) RecordFailedLogin(ctx context.Context, id uuid.UUID, lockUntil *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	updates := map[string]any{
		"failed_login_attempts": gorm.Expr("failed_login_attempts + 1"),
	}

	if lockUntil != nil {
		updates["locked_until"] = lockUntil
	}

	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepository) RecordSuccessfulLogin(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_login_at":         time.Now().UTC(),
			"failed_login_attempts": 0,
		}).
		Error
}

func (r *UserRepository) UpdateAdminStatus(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("is_admin", isAdmin).Error
}

// User notification preferences
func (r *UserRepository) CreateNotificationPreferences(ctx context.Context, prefs *models.UserNotificationPreferences) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(prefs).Error
}

func (r *UserRepository) UpdateNotificationPreferences(ctx context.Context, prefs *models.UserNotificationPreferences) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Save(prefs).Error
}

// Delete user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}
