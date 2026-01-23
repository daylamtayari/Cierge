package cloud

import (
	"context"
	"errors"

	"github.com/daylamtayari/cierge/internal/config"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported cloud provider specified")
)

// Provider defines the interface that all cloud providers must implement
// TODO: Complete this method declarations once their scope has been fully established
type Provider interface {
	ScheduleJob(ctx context.Context) error

	CancelJob(ctx context.Context) error

	EncryptData(ctx context.Context) error
}

// Contains a cloud provider's constructor
// and configuration validator
type ProviderRegistration struct {
	Constructor ProviderConstructor
	Validator   ProviderConfigValidator
}

// Represents a cloud provider's constructor
type ProviderConstructor func(config map[string]any) (Provider, error)

// Represents a cloud provider's configuration validator
type ProviderConfigValidator func(config map[string]any, isProduction bool) error

var registry = make(map[config.CloudProvider]ProviderRegistration)

// Registers a cloud provider's constructor
func Register(name config.CloudProvider, constructor ProviderConstructor, configValidator ProviderConfigValidator) {
	if constructor == nil {
		panic("cloud: register constructor is nil")
	}
	if _, exists := registry[name]; exists {
		panic("cloud: register called twice for the same provider: " + name)
	}

	registry[name] = ProviderRegistration{
		Constructor: constructor,
		Validator:   configValidator,
	}
}

// Creates a new provider for the cloud provider specified in the config
func NewProvider(cfg *config.CloudConfig) (Provider, error) {
	provider, exists := registry[cfg.Provider]
	if !exists {
		return nil, ErrUnsupportedProvider
	}

	return provider.Constructor(cfg.Config)
}

// Validates a cloud provider's config
func ValidateConfig(cfg *config.CloudConfig, isProduction bool) error {
	provider, exists := registry[cfg.Provider]
	if !exists {
		return ErrUnsupportedProvider
	}

	return provider.Validator(cfg.Config, isProduction)
}

// Returns a slice of all available cloud providers
func AvailableProviders() []config.CloudProvider {
	providers := make([]config.CloudProvider, 0, len(registry))
	for name := range registry {
		providers = append(providers, name)
	}

	return providers
}
