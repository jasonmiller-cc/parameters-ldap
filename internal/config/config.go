// Package config holds the LDAP service configuration, extending BaseConfig
// from parameters-core with LDAP-specific fields.
package config

import (
	"github.com/jasonmiller-cc/parameters-core/pkg/config"
)

// LDAPConfig holds connection parameters for the LDAP/FreeIPA directory.
type LDAPConfig struct {
	URL           string `yaml:"url"`
	BindDN        string `yaml:"bind_dn"`
	BindPassword  string `yaml:"bind_password"`
	BaseDN        string `yaml:"base_dn"`
	UserBaseDN    string `yaml:"user_base_dn"`
	GroupBaseDN   string `yaml:"group_base_dn"`
	TLSSkipVerify bool   `yaml:"tls_skip_verify"`
}

// Config is the full service configuration.
type Config struct {
	config.BaseConfig `yaml:",inline"`
	LDAP              LDAPConfig `yaml:"ldap"`
}

// Load reads config from path (or discovers it via PARAMS_LDAP_CONFIG / standard
// locations when path is empty) and returns a ready-to-use *Config.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if err := config.Load(path, "PARAMS_LDAP", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
