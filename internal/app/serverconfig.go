package app

import (
	"os"
	"strings"

	"github.com/nomos/nomos/internal/storage"
	"gopkg.in/yaml.v3"
)

// ServerConfig is the optional, file-based identity of a Nomos server, stored at
// .nomos/server.yml. It lets a deployment declare its public domain and display
// label in the workspace instead of relying solely on environment variables.
// The file is non-authoritative operator config: NOMOS_DOMAIN still overrides
// Domain, and a DNS domain is only honored once provably owned (see
// localDisplayEndpointHint).
type ServerConfig struct {
	Domain string `yaml:"domain,omitempty" json:"domain,omitempty"`
	Label  string `yaml:"label,omitempty" json:"label,omitempty"`
}

// LoadServerConfig reads .nomos/server.yml for the workspace. The file is
// optional: a missing or unreadable file yields a zero ServerConfig rather than
// an error, so server startup never depends on it.
func LoadServerConfig(path string) ServerConfig {
	var cfg ServerConfig
	data, err := os.ReadFile(storage.ServerConfigFile(path))
	if err != nil {
		return ServerConfig{}
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ServerConfig{}
	}
	cfg.Domain = strings.TrimSpace(cfg.Domain)
	cfg.Label = strings.TrimSpace(cfg.Label)
	return cfg
}
