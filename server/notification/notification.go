package notification

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
)

var (
	ErrDuplicateProvider   = errors.New("notifcation register called twice for the same provider")
	ErrNilConstructor      = errors.New("notification register constructor is nil")
	ErrUnsupportedProvider = errors.New("unsupported notification provider specified")
)

type Type string

const (
	JobSuccess   = Type("job_success")
	JobFail      = Type("job_fail")
	TokenExpired = Type("token_expired")
)

// Defines the interface that all notification providers must implement
type Provider interface {
	Send(ctx context.Context, notifType Type, message string)
}

// Represents a notification provider's constructor
type ProviderConstructor func(config map[string]any) (Provider, error)

// Represents a notification provider's configuration validator
type ProviderConfigValidator func(config map[string]any, isProduction bool) error

// Contains a notification provider's
// constructor and configuration validator
type ProviderRegistration struct {
	Constructor ProviderConstructor
	Validator   ProviderConfigValidator
}

// A map of notification providers
// The key value is always lower case
var registry = make(map[string]ProviderRegistration)

// Registers a notification provider's constructor and validator
func Register(name string, constructor ProviderConstructor, configValidator ProviderConfigValidator) error {
	if constructor == nil {
		return ErrNilConstructor
	}
	if _, exists := registry[name]; exists {
		return ErrDuplicateProvider
	}

	registry[strings.ToLower(name)] = ProviderRegistration{
		Constructor: constructor,
		Validator:   configValidator,
	}

	return nil
}

// Creates a new provider for the notification provider specified
func NewProvider(name string, config map[string]any) (Provider, error) {
	provider, exists := registry[name]
	if !exists {
		return nil, ErrUnsupportedProvider
	}

	return provider.Constructor(config)
}

// Validates a notification provider's config
func ValidateConfig(name string, config map[string]any, isProduction bool) error {
	provider, exists := registry[name]
	if !exists {
		return ErrUnsupportedProvider
	}

	return provider.Validator(config, isProduction)
}

// Returns a slice of all available of notification providers
func AvailableProviders() []string {
	return slices.Collect(maps.Keys(registry))
}
