package cloud

import (
	"context"
	"errors"
	"strings"
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

// A map of cloud providers
// The key value is always lower case
var registry = make(map[string]ProviderRegistration)

// Registers a cloud provider's constructor
func Register(name string, constructor ProviderConstructor, configValidator ProviderConfigValidator) {
	if constructor == nil {
		panic("cloud: register constructor is nil")
	}
	if _, exists := registry[name]; exists {
		panic("cloud: register called twice for the same provider: " + name)
	}

	registry[strings.ToLower(name)] = ProviderRegistration{
		Constructor: constructor,
		Validator:   configValidator,
	}
}

// Creates a new provider for the cloud provider specified in the config
func NewProvider(name string, config map[string]any) (Provider, error) {
	provider, exists := registry[name]
	if !exists {
		return nil, ErrUnsupportedProvider
	}

	return provider.Constructor(config)
}

// Validates a cloud provider's config
func ValidateConfig(name string, config map[string]any, isProduction bool) error {
	provider, exists := registry[name]
	if !exists {
		return ErrUnsupportedProvider
	}

	return provider.Validator(config, isProduction)
}

// Returns a slice of all available cloud providers
func AvailableProviders() []string {
	providers := make([]string, 0, len(registry))
	for name := range registry {
		providers = append(providers, name)
	}

	return providers
}
