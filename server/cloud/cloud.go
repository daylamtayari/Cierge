package cloud

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/daylamtayari/cierge/reservation"
	"github.com/google/uuid"
)

var (
	ErrDuplicateProvider   = errors.New("cloud register called twice for the same provider")
	ErrNilConstructor      = errors.New("cloud register constructor is nil")
	ErrUnsupportedProvider = errors.New("unsupported cloud provider specified")
	ErrJobNotFound         = errors.New("job was not found")
)

// Provider defines the interface that all cloud providers must implement
type Provider interface {
	// Creates a scheduled invocation of a reservation
	ScheduleJob(ctx context.Context, event reservation.Event) error

	// Updates the encrypted platform token in an already-scheduled job
	// For when a user updates their credentials but they have already scheduled jobs
	UpdateJobCredentials(ctx context.Context, jobID uuid.UUID, encryptedToken string) error

	// Deletes the scheduled invocation for a given job
	// Returned error will be nil if the job already executed or does not exist
	CancelJob(ctx context.Context, jobID uuid.UUID) error

	// Encrypts plaintext and returns base64-encoded ciphertext
	EncryptData(ctx context.Context, plaintext string) (string, error)
}

// Represents a cloud provider's constructor
type ProviderConstructor func(config map[string]any) (Provider, error)

// Represents a cloud provider's configuration validator
type ProviderConfigValidator func(config map[string]any, isProduction bool) error

// Contains a cloud provider's constructor
// and configuration validator
type ProviderRegistration struct {
	Constructor ProviderConstructor
	Validator   ProviderConfigValidator
}

// A map of cloud providers
// The key value is always lower case
var registry = make(map[string]ProviderRegistration)

// Registers a cloud provider's constructor and validator
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
	return slices.Collect(maps.Keys(registry))
}
