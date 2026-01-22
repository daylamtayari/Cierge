package aws

import (
	"context"

	"github.com/daylamtayari/cierge/internal/cloud"
	"github.com/daylamtayari/cierge/internal/config"
)

type AWSProvider struct {
}

func init() {
	cloud.Register(config.CloudProviderAWS, NewAWSProvider)
}

// Returns a new AWS provider
func NewAWSProvider(config map[string]any) (cloud.Provider, error) {
	return &AWSProvider{}, nil
}

func (p *AWSProvider) ScheduleJob(ctx context.Context) error {
	return nil
}

func (p *AWSProvider) CancelJob(ctx context.Context) error {
	return nil
}

func (p *AWSProvider) EncryptData(ctx context.Context) error {
	return nil
}
