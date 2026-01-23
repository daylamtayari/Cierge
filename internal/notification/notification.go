package cloud

import (
	"context"
	"errors"

	"github.com/daylamtayari/cierge/internal/config"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported notification provider specified")
)

type NotificationType string

const (
	JobSuccessNotification   = NotificationType("job_success")
	JobFailNotification      = NotificationType("job_fail")
	TokenExpiredNotification = NotificationType("token_expired")
)

// Defines the interface that all notification providers must implement
type Provider interface {
	Send(ctx context.Context, notifType NotificationType, message string)
}

type ProviderConstructor func(config map[string]any) (Provider, error)

var registry = make(map[string]ProviderConstructor)

// Registers a notification provider's constructor
func Register(name string, constructor ProviderConstructor) {
	if constructor == nil {
		panic("notification: register constructor is nil")
	}
	if _, exists := registry[name]; exists {
		panic("notification: register called twice for the same provider: " + name)
	}

	registry[name] = constructor
}

// Creates a new provider for the notification provider specified
func NewProvider(cfg *config.NotificationProvider) (Provider, error) {
	constructor, exists := registry[cfg.Name]
	if !exists {
		return nil, ErrUnsupportedProvider
	}

	return constructor(cfg.Config)
}

// Returns a slice of all available of notification providers
func AvailableProviders() []string {
	providers := make([]string, 0, len(registry))
	for name := range registry {
		providers = append(providers, name)
	}

	return providers
}

