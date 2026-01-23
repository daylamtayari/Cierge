package aws

import (
	"context"

	"github.com/daylamtayari/cierge/internal/cloud"
	"github.com/daylamtayari/cierge/internal/config"
)

type Provider struct {
}

func init() {
	cloud.Register(config.CloudProviderAWS, NewProvider, ValidateConfig)
}

// Returns a new AWS provider
func NewProvider(config map[string]any) (cloud.Provider, error) {
	return &Provider{}, nil
}

func (p *Provider) ScheduleJob(ctx context.Context) error {
	return nil
}

func (p *Provider) CancelJob(ctx context.Context) error {
	return nil
}

func (p *Provider) EncryptData(ctx context.Context) error {
	return nil
}

func ValidateConfig(config map[string]any, isProduction bool) error {
	return nil
}
