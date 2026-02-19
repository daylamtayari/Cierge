package local

import (
	"context"

	"github.com/daylamtayari/cierge/reservation"
	"github.com/daylamtayari/cierge/server/cloud"
	"github.com/google/uuid"
)

type Provider struct {
}

func NewProvider(config map[string]any) (cloud.Provider, error) {
	return &Provider{}, nil
}

func (p *Provider) ScheduleJob(ctx context.Context, event reservation.Event) error {
	return nil
}

func (p *Provider) UpdateJobCredentials(ctx context.Context, jobID uuid.UUID, encryptedToken string) error {
	return nil
}

func (p *Provider) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	return nil
}

func (p *Provider) EncryptData(ctx context.Context, plaintext string) (string, error) {
	return "", nil
}

func (p *Provider) DecryptData(ctx context.Context, ciphertext string) (string, error) {
	return "", nil
}

func ValidateConfig(config map[string]any, isProduction bool) error {
	return nil
}
