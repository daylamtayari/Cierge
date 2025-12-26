package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Collects all validation errors that were identified
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return "config validation failed: " + strings.Join(msgs, "; ")
}

func (e ValidationErrors) MarshalJSON() ([]byte, error) {
	return json.Marshal([]ValidationError(e))
}

func (c *Config) Validate() error {
	var errs ValidationErrors

	// Server validation
	if c.Server.Host == "" {
		// There is a default value of localhost if not specified but if it is explicitly set to an empty string, error
		errs = append(errs, ValidationError{"server.host", "server host value is required"})
	}
	if err := portValidation(c.Server.Port); err != nil {
		errs = append(errs, ValidationError{"server.port", err.Error()})
	}
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" {
			errs = append(errs, ValidationError{"server.tls.cert_file", "path to certificate is required when TLS is enabled"})
		}
		if c.Server.TLS.KeyFile == "" {
			errs = append(errs, ValidationError{"server.tls.key_file", "path to key file is required when TLS is enabled"})
		}
	}

	// Database validation
	if c.Database.Host == "" {
		errs = append(errs, ValidationError{"database.host", "database host value is required"})
	}
	if err := portValidation(c.Database.Port); err != nil {
		errs = append(errs, ValidationError{"server.port", err.Error()})
	}
	if c.Database.User == "" {
		errs = append(errs, ValidationError{"database.user", "database user value is required"})
	}
	if c.Database.Database == "" {
		errs = append(errs, ValidationError{"database.database", "database name is required"})
	}

	// Auth validation
	if !c.IsDevelopment() && c.Auth.JWTSecret == "" {
		errs = append(errs, ValidationError{"auth.jwt_secret", "JWT secret is required in production"})
	}
	if len(c.Auth.JWTSecret) > 0 && len(c.Auth.JWTSecret) < 64 {
		errs = append(errs, ValidationError{"auth.jwt_secret", "JWT secret must be at least 64 characters"})
	}
	// Validate auth method
	switch c.Auth.Method {
	case AuthMethodLocal, AuthMethodOIDC:
		// Valid
	case "":
		errs = append(errs, ValidationError{"auth.method", "auth method must be specified"})
	default:
		errs = append(errs, ValidationError{"auth.method", "auth method must be 'local' or 'oidc'"})
	}
	// Validate OIDC
	if c.SupportsOIDC() {
		if len(c.Auth.OIDCProviders) == 0 {
			errs = append(errs, ValidationError{"auth.oidc_providers", "at least one provider is required when using OIDC"})
		}
		for name, provider := range c.Auth.OIDCProviders {
			prefix := fmt.Sprintf("auth.oidc.providers.%s", name)

			if provider.ClientID == "" {
				errs = append(errs, ValidationError{prefix + ".client_id", "client ID is required"})
			}
			if provider.ClientSecret == "" {
				errs = append(errs, ValidationError{prefix + ".client_secret", "client secret is required"})
			}
			if provider.IssuerURL == "" {
				errs = append(errs, ValidationError{prefix + ".issuer_url", "issuer URL is required"})
			} else if _, err := url.Parse(provider.IssuerURL); err != nil {
				errs = append(errs, ValidationError{prefix + ".issuer_url", "issuer URL must be a valid URL"})
			}
			if provider.RedirectURL == "" {
				errs = append(errs, ValidationError{prefix + ".redirect_url", "redirect URL is required"})
			}
		}
	}

	// Cloud validation
	switch c.Cloud.Provider {
	case CloudProviderAWS:
		// Valid
	case CloudProviderAzure, CloudProviderGCP:
		errs = append(errs, ValidationError{"cloud.provider", "GCP and Azure are not currently supported cloud providers"})
	case "":
		errs = append(errs, ValidationError{"cloud.provider", "cloud provider must be specified"})
	default:
		errs = append(errs, ValidationError{"cloud.provider", "cloud provider must be 'aws', 'azure', or 'gcp'"})
	}
	if c.Cloud.Config == nil {
		errs = append(errs, ValidationError{"cloud.config", "cloud configuration must be provided"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func portValidation(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port value must be between 1 and 65535")
	}
	return nil
}
