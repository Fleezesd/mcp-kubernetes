package mcpkubernetes

import (
	"github.com/fleezesd/mcp-kubernetes/internal/mcp-kubernetes/mcp"
)

type Config struct {
	SSEBaseURL string
	SSEPort    int
	KubeConfig string
}

type CompletedConfig struct {
	*Config
}

func (c *Config) Complete() CompletedConfig {
	return CompletedConfig{c}
}

func (c *Config) New() (*mcp.Server, error) {
	return mcp.NewServer(c.KubeConfig)
}
