package service

import (
	"context"
	"errors"
	"time"

	"github.com/daylamtayari/cierge/server/cloud"
	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/rs/zerolog"
)

var (
	ErrRefreshExpired = errors.New("refresh token expired")
)

type TokenRenewer struct {
	ptService     *PlatformToken
	jobService    *Job
	cloudProvider cloud.Provider
	logger        zerolog.Logger
	interval      time.Duration
	renewBefore   time.Duration
}

func NewTokenRenewer(ptService *PlatformToken, jobService *Job, cloudProvider cloud.Provider, logger zerolog.Logger, cfg config.PlatformToken) *TokenRenewer {
	return &TokenRenewer{
		ptService:     ptService,
		jobService:    jobService,
		cloudProvider: cloudProvider,
		logger:        logger.With().Str("component", "token_renewer").Logger(),
		interval:      cfg.RenewalInterval.Duration(),
		renewBefore:   cfg.RenewBefore.Duration(),
	}
}

// Start a token renewer goroutine
func (r *TokenRenewer) Start(ctx context.Context) {
	go r.run(ctx)
}

// Runs the token renewer ticker
func (r *TokenRenewer) run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.renewAll(ctx)
		}
	}
}

func (r *TokenRenewer) renewAll(ctx context.Context) {
	tokens, err := r.ptService.GetExpiringWithRefresh(ctx, r.renewBefore)
	if err != nil {
		r.logger.Error().Err(err).Msg("failed to retrieve expiring tokens")
		return
	}

	for _, token := range tokens {
		if err := r.renewToken(ctx, token); err != nil {
			r.logger.Error().Err(err).
				Str("platform", token.Platform).
				Stringer("token_id", token.ID).
				Stringer("user_id", token.UserID).
				Msg("failed to renew token")
			// TODO: Implement notifications if remaining refresh time is short
			continue
		}
		r.logger.Info().
			Str("platform", token.Platform).
			Stringer("token_id", token.ID).
			Stringer("user_id", token.UserID).
			Msg("successfully renewed token")
	}
}

// Renew a token and update the new credentials in all scheduled jobs
func (r *TokenRenewer) renewToken(ctx context.Context, token *model.PlatformToken) error {
	// Check if refresh token is not expired
	if token.RefreshExpiresAt != nil && time.Now().UTC().After(*token.RefreshExpiresAt) {
		return ErrRefreshExpired
	}

	newToken, err := r.ptService.refreshToken(ctx, token)
	if err != nil {
		return err
	}

	// Update all scheduled jobs with the new tokens
	jobs, err := r.jobService.GetScheduledByUserAndPlatform(ctx, token.UserID, token.Platform)
	if err != nil && errors.Is(err, ErrJobDNE) {
		// No existing jobs scheduled
		return nil
	} else if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := r.cloudProvider.UpdateJobCredentials(ctx, job.ID, newToken.EncryptedToken); err != nil {
			r.logger.Error().Err(err).Stringer("job_id", job.ID).Msg("failed to update job credentials")
		}
	}
	return nil
}
