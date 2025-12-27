package config

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Environment  Environment        `json:"environment" default:"dev"`
	LogLevel     zerolog.Level      `json:"log_level" default:"info"`
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	Auth         AuthConfig         `json:"auth"`
	Cloud        CloudConfig        `json:"cloud"`
	Notification NotificationConfig `json:"notification"`
}

type Environment string

const (
	EnvironmentDev  Environment = "dev"
	EnvironmentProd Environment = "prod"
)

// Necessary for parsing of defaults from tags
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// HTTP server configuration
type ServerConfig struct {
	Host           string    `json:"host" default:"localhost"`
	Port           int       `json:"port" default:"8080"`
	TLS            TLSConfig `json:"tls"`
	TrustedProxies []string  `json:"trusted_proxies"`
	CORSOrigins    []string  `json:"cors_origins"`
}

type TLSConfig struct {
	Enabled  bool   `json:"enabled" default:"false"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Database configuration
type DatabaseConfig struct {
	Host        string `json:"host" default:"localhost"`
	Port        int    `json:"port" default:"5432"`
	User        string `json:"user" default:"cierge"`
	Password    string `json:"password"`
	Database    string `json:"database" default:"cierge"`
	SSLMode     string `json:"ssl_mode" default:"disable"`
	AutoMigrate bool   `json:"auto_migrate" default:"true"`
}

func (d DatabaseConfig) DSN() string {
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

type AuthConfig struct {
	Method        AuthMethod              `json:"method" default:"local"`
	OIDCProviders map[string]OIDCProvider `json:"oidc_providers"`

	// JWT configuration
	JWTSecret          string   `json:"jwt_secret"`
	JWTIssuer          string   `json:"jwt_issuer" default:"cierge"`
	AccessTokenExpiry  Duration `json:"access_token_expiry" default:"15m"`
	RefreshTokenExpiry Duration `json:"refresh_token_expiry" default:"7d"`

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
type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderAzure CloudProvider = "azure"
	CloudProviderGCP   CloudProvider = "gcp"
)

type CloudConfig struct {
	Provider CloudProvider  `json:"provider" default:"aws"`
	Config   map[string]any `json:"config"`
}

// Notification configuration
type NotificationConfig struct {
	Email   NotificationChannelConfig `json:"email"`
	SMS     NotificationChannelConfig `json:"sms"`
	Webhook NotificationChannelConfig `json:"webhook"`
}

type NotificationChannelConfig struct {
	Enabled     bool `json:"enabled" default:"false"`
	TokenExpiry bool `json:"token_expiry" default:"true"`
	JobStarted  bool `json:"job_started" default:"false"`
	JobSuccess  bool `json:"job_success" deffault:"true"`
	JobFailed   bool `json:"job_failed" default:"true"`
}
