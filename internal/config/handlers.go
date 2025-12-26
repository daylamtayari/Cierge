package config

import "fmt"

func (c *Config) IsDevelopment() bool {
	return c.Environment == EnvironmentDev
}

func (c *Config) SupportsLocalAuth() bool {
	return c.Auth.Method == AuthMethodLocal
}

func (c *Config) SupportsOIDC() bool {
	return c.Auth.Method == AuthMethodOIDC
}

func (c *Config) GetOIDCProvider(name string) (OIDCProvider, error) {
	provider, ok := c.Auth.OIDCProviders[name]
	if !ok {
		return OIDCProvider{}, fmt.Errorf("no OIDC provider configured with name %q", name)
	}
	return provider, nil
}
