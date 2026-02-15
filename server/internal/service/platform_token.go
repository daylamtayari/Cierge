package service

import (
	"context"
	"errors"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTokenDNE = errors.New("token does not exist")
)

type PlatformToken struct {
	platformTokenRepo *repository.PlatformToken
}

func NewPlatformToken(platformTokenRepo *repository.PlatformToken) *PlatformToken {
	return &PlatformToken{
		platformTokenRepo: platformTokenRepo,
	}
}

// Gets a platform token from a given ID
func (s *PlatformToken) GetByID(ctx context.Context, tokenID uuid.UUID) (*model.PlatformToken, error) {
	platformToken, err := s.platformTokenRepo.GetByID(ctx, tokenID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenDNE
	} else if err != nil {
		return nil, err
	}
	return platformToken, nil
}

// Gets all platorm tokens for a given user
func (s *PlatformToken) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.PlatformToken, error) {
	platformTokens, err := s.platformTokenRepo.GetByUser(ctx, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenDNE
	} else if err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Gets a platform token for a speciffied user and token
func (s *PlatformToken) GetByUserAndPlatform(ctx context.Context, userID uuid.UUID, platform string) (*model.PlatformToken, error) {
	platformToken, err := s.platformTokenRepo.GetByUserAndPlatform(ctx, userID, platform)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenDNE
	} else if err != nil {
		return nil, err
	}
	return platformToken, nil
}

// Replaces the token for a specific user and platform with a new one
func (s *PlatformToken) Replace(ctx context.Context, newToken *model.PlatformToken) error {
	userId := newToken.UserID
	platform := newToken.Platform

	oldToken, err := s.platformTokenRepo.GetByUserAndPlatform(ctx, userId, platform)
	if err != nil {
		return err
	}

	return s.platformTokenRepo.Replace(ctx, oldToken.ID, newToken)
}

// Delete's a specified token
func (s *PlatformToken) Delete(ctx context.Context, tokenId uuid.UUID) error {
	err := s.platformTokenRepo.Delete(ctx, tokenId)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrTokenDNE
	} else if err != nil {
		return err
	}
	return nil
}
