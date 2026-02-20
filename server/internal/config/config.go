package config

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Environment    Environment            `json:"environment" default:"dev"`
	LogLevel       zerolog.Level          `json:"log_level" default:"info"`
	Server         Server                 `json:"server"`
	Database       Database               `json:"database"`
	Auth           Auth                   `json:"auth"`
	TokenStorePath string                 `json:"token_store_path" default:"./data/token_store"`
	Cloud          Cloud                  `json:"cloud"`
	Notification   []NotificationProvider `json:"notification"`
	DefaultAdmin   User                   `json:"default_admin"`
}

type Environment string

const (
	EnvironmentDev  Environment = "dev"
	EnvironmentProd Environment = "prod"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Necessary for parsing of defaults from tags
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// HTTP server configuration
type Server struct {
	Host           string   `json:"host" default:"localhost"`
	ExternalHost   string   `json:"external_host"`
	Port           int      `json:"port" default:"8080"`
	TLS            TLS      `json:"tls"`
	TrustedProxies []string `json:"trusted_proxies"`
	CORSOrigins    []string `json:"cors_origins"`
}

type TLS struct {
	Enabled  bool   `json:"enabled" default:"false"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s Server) URL() string {
	scheme := "http"
	if s.TLS.Enabled {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/", scheme, s.Host, s.Port)
}

func (s Server) ExternalURL() string {
	scheme := "http"
	if s.TLS.Enabled {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/", scheme, s.ExternalHost, s.Port)
}

// Database configuration
type Database struct {
	Host        string   `json:"host" default:"localhost"`
	Port        int      `json:"port" default:"5432"`
	User        string   `json:"user" default:"cierge"`
	Password    string   `json:"password"`
	Database    string   `json:"database" default:"cierge"`
	SSLMode     string   `json:"ssl_mode" default:"disable"`
	AutoMigrate bool     `json:"auto_migrate" default:"true"`
	Timeout     Duration `json:"timeout" default:"30s"`
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode,
	)
}

// Auth configuration
type AuthMethod string

const (
	AuthMethodLocal AuthMethod = "local"
	AuthMethodOIDC  AuthMethod = "oidc"
)

type Auth struct {
	Method        AuthMethod              `json:"method" default:"local"`
	OIDCProviders map[string]OIDCProvider `json:"oidc_providers"`

	// JWT configuration
	JWTSecret          string   `json:"jwt_secret"`
	JWTIssuer          string   `json:"jwt_issuer" default:"cierge"`
	AccessTokenExpiry  Duration `json:"access_token_expiry" default:"15m"`
	RefreshTokenExpiry Duration `json:"refresh_token_expiry" default:"168h"`

	// Rate limiting for local authentication - not used for OIDC
	RateLimitRequests int      `json:"rate_limit_requests" default:"3"`
	RateLimitWindow   Duration `json:"rate_limit_window" default:"5m"`
}

type OIDCProvider struct {
	ClientID          string   `json:"client_id"`
	ClientSecret      string   `json:"client_secret"`
	IssuerURL         string   `json:"issuer_url"`
	RedirectURL       string   `json:"redirect_url"`
	Scopes            []string `json:"scopes"`
	BackchannelLogout bool     `json:"backchannel_logout"`
}

// Cloud configuration

type Cloud struct {
	Provider string         `json:"provider" default:"aws"`
	Config   map[string]any `json:"config"`
}

// Notification configuration

// Represents a notification provider and it's config
// The name should match the notification provider name
type NotificationProvider struct {
	Name    string         `json:"name"`
	Enabled bool           `json:"enabled"`
	Config  map[string]any `json:"config"`
}
