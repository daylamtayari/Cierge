package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/daylamtayari/cierge/resy"
	"github.com/daylamtayari/cierge/server/cloud"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTokenDNE          = errors.New("token does not exist")
	ErrIncorrectPlatform = errors.New("specified platform is incorrect")
)

type PlatformToken struct {
	ptRepo *repository.PlatformToken
}

func NewPlatformToken(platformTokenRepo *repository.PlatformToken) *PlatformToken {
	return &PlatformToken{
		ptRepo: platformTokenRepo,
	}
}

// Gets a platform token from a given ID
func (s *PlatformToken) GetByID(ctx context.Context, tokenID uuid.UUID) (*model.PlatformToken, error) {
	platformToken, err := s.ptRepo.GetByID(ctx, tokenID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenDNE
	} else if err != nil {
		return nil, err
	}
	return platformToken, nil
}

// Gets all platorm tokens for a given user
func (s *PlatformToken) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.PlatformToken, error) {
	platformTokens, err := s.ptRepo.GetByUser(ctx, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenDNE
	} else if err != nil {
		return nil, err
	}
	return platformTokens, nil
}

// Gets a platform token for a speciffied user and token
func (s *PlatformToken) GetByUserAndPlatform(ctx context.Context, userID uuid.UUID, platform string) (*model.PlatformToken, error) {
	platformToken, err := s.ptRepo.GetByUserAndPlatform(ctx, userID, platform)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenDNE
	} else if err != nil {
		return nil, err
	}
	return platformToken, nil
}

// Creates a new token, replacing any existing one
// Encrypts the token string and adds expiry and refresh values
func (s *PlatformToken) Create(ctx context.Context, cloudProvider cloud.Provider, userID uuid.UUID, platform string, token any) error {
	tokenString, err := json.Marshal(token)
	if err != nil {
		return err
	}
	encryptedToken, err := cloudProvider.EncryptData(ctx, string(tokenString))
	if err != nil {
		return err
	}

	existingToken, err := s.ptRepo.GetByUserAndPlatform(ctx, userID, platform)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var existingTokenId *uuid.UUID
	if existingToken != nil {
		existingTokenId = &existingToken.ID
	}

	newToken := &model.PlatformToken{
		UserID:         userID,
		Platform:       platform,
		EncryptedToken: encryptedToken,
	}

	switch platform {
	case "resy":
		resyToken, ok := token.(resy.Tokens)
		if !ok {
			return ErrIncorrectPlatform
		}

		authExpires, err := resy.GetTokenExpiry(resyToken.Token)
		if err != nil {
			return err
		}
		newToken.ExpiresAt = &authExpires

		refreshExpiresAt, err := resy.GetTokenExpiry(resyToken.Refresh)
		if err != nil {
			return err
		}
		newToken.HasRefresh = true
		newToken.RefreshExpiresAt = &refreshExpiresAt

	case "opentable":
		// TODO: Implement opentable
	}

	return s.ptRepo.Replace(ctx, existingTokenId, newToken)
}

// Replaces the token for a specific user and platform with a new one
func (s *PlatformToken) Replace(ctx context.Context, newToken *model.PlatformToken) error {
	userId := newToken.UserID
	platform := newToken.Platform

	oldToken, err := s.ptRepo.GetByUserAndPlatform(ctx, userId, platform)
	if err != nil {
		return err
	}

	return s.ptRepo.Replace(ctx, &oldToken.ID, newToken)
}

// Delete's a specified token
func (s *PlatformToken) Delete(ctx context.Context, tokenId uuid.UUID) error {
	err := s.ptRepo.Delete(ctx, tokenId)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrTokenDNE
	} else if err != nil {
		return err
	}
	return nil
}
