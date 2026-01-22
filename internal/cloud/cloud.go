package cloud

import (
	"context"
	"errors"

	"github.com/daylamtayari/cierge/internal/config"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported provider specified")
)

// Provider defines the interface that all cloud providers must implement
// TODO: Complete this method declarations once their scope has been fully established
type Provider interface {
	ScheduleJob(ctx context.Context) error

	CancelJob(ctx context.Context) error

	EncryptData(ctx context.Context) error
}

type ProviderConstructor func(config map[string]any) (Provider, error)

var registry = make(map[config.CloudProvider]ProviderConstructor)

// Registers a cloud provider's constructor
func Register(name config.CloudProvider, constructor ProviderConstructor) {
	if constructor == nil {
		panic("cloud: register constructor is nil")
	}
	if _, exists := registry[name]; exists {
		panic("cloud: register called twice for the same provider: " + name)
	}

	registry[name] = constructor
}

// Creates a new provider for the cloud provider specified in the config
func NewProvider(cfg *config.CloudConfig) (Provider, error) {
	constructor, exists := registry[cfg.Provider]
	if !exists {
		return nil, ErrUnsupportedProvider
	}

	return constructor(cfg.Config)
}

// Returns a slice of all available cloud providers
func AvailableProviders() []config.CloudProvider {
	providers := make([]config.CloudProvider, 0, len(registry))
	for name := range registry {
		providers = append(providers, name)
	}

	return providers
}

