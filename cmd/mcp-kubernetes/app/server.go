package app

import (
	"github.com/fleezesd/mcp-kubernetes/cmd/mcp-kubernetes/app/options"
	mcpkubernetes "github.com/fleezesd/mcp-kubernetes/internal/mcp-kubernetes"
	"github.com/fleezesd/mcp-kubernetes/pkg/app"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

const commandDesc = `
  Kubernetes Model Context Protocol (MCP) Server

  Usage:
    kubernetes-mcp-server [flags]

  Available Commands:
    -h, --help        Display help information
    --version         Display version information
    
  Server Options:
    --sse-port        Port number for SSE server (e.g. 8080, 8443)
    --sse-base-url    Base URL for HTTPS host (e.g. https://example.com:8443)

  Examples:
    # Start STDIO server
    kubernetes-mcp-server

    # Start SSE server on port 8080
    kubernetes-mcp-server --sse-port 8080

    # Start SSE server on port 8443 with HTTPS
    kubernetes-mcp-server --sse-port 8443 --sse-base-url https://example.com:8443
`

func NewApp() *app.App {
	opts := options.NewOptions()

	application := app.NewApp("mcp-kubernetes-server", "Kubernetes Model Context Protocol (MCP) server",
		app.WithDescription(commandDesc),
		app.WithOptions(opts),
		app.WithRunFunc(run(opts)),
	)
	return application
}

func run(opts *options.Options) app.RunFunc {
	return func() error {
		cfg, err := opts.Config()
		if err != nil {
			return err
		}
		return Run(cfg, genericapiserver.SetupSignalHandler())
	}
}

func Run(c *mcpkubernetes.Config, stopCh <-chan struct{}) error {
	mcpServer, err := c.Complete().New()
	if err != nil {
		return err
	}

	mcpServer.Run(c.SSEBaseURL, c.SSEPort, stopCh)
	return nil
}
