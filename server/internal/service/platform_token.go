package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/daylamtayari/cierge/api"
	"github.com/daylamtayari/cierge/resy"
	"github.com/daylamtayari/cierge/server/cloud"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTokenDNE            = errors.New("token does not exist")
	ErrIncorrectPlatform   = errors.New("specified platform is incorrect")
	ErrUnsupportedPlatform = errors.New("unsupported platform was provided")
)

type PlatformToken struct {
	ptRepo        *repository.PlatformToken
	cloudProvider cloud.Provider
}

func NewPlatformToken(platformTokenRepo *repository.PlatformToken, cloudProvider cloud.Provider) *PlatformToken {
	return &PlatformToken{
		ptRepo:        platformTokenRepo,
		cloudProvider: cloudProvider,
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
func (s *PlatformToken) Create(ctx context.Context, userID uuid.UUID, platform string, token any) (*api.PlatformToken, error) {
	tokenString, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	encryptedToken, err := s.cloudProvider.EncryptData(ctx, string(tokenString))
	if err != nil {
		return nil, err
	}

	var existingTokenId *uuid.UUID
	existingToken, err := s.ptRepo.GetByUserAndPlatform(ctx, userID, platform)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if err == nil {
		existingTokenId = &existingToken.ID
	}

	newToken := &model.PlatformToken{
		UserID:         userID,
		Platform:       platform,
		EncryptedToken: encryptedToken,
		CreatedAt:      time.Now().UTC(),
	}

	switch platform {
	case "resy":
		resyToken, ok := token.(resy.Tokens)
		if !ok {
			return nil, ErrIncorrectPlatform
		}

		authExpires, err := resy.GetTokenExpiry(resyToken.Token)
		if err != nil {
			return nil, err
		}
		newToken.ExpiresAt = &authExpires

		refreshExpiresAt, err := resy.GetTokenExpiry(resyToken.Refresh)
		if err != nil {
			return nil, err
		}
		newToken.HasRefresh = true
		newToken.RefreshExpiresAt = &refreshExpiresAt

	case "opentable":
		// TODO: Implement opentable
	default:
		return nil, ErrUnsupportedPlatform
	}

	err = s.ptRepo.Replace(ctx, existingTokenId, newToken)
	if err != nil {
		return nil, err
	}
	return newToken.ToAPI(), nil
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
