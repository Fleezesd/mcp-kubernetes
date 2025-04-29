package mcp

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/fleezesd/mcp-kubernetes/pkg/kubernetes"
	"github.com/fleezesd/mcp-kubernetes/pkg/log"
	"github.com/fleezesd/mcp-kubernetes/pkg/version"
	"github.com/mark3labs/mcp-go/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

type Server struct {
	configuration *configuration
	server        *server.MCPServer
	k             *kubernetes.Kubernetes
}

type configuration struct {
	kubeConfig string
}

func NewServer(kubeConfig string) (*Server, error) {
	s := &Server{
		configuration: &configuration{
			kubeConfig: kubeConfig,
		},
		server: server.NewMCPServer(
			"mcp-kubernetes",
			version.Get().String(),
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
		),
	}
	if err := s.reloadKubernetesClient(); err != nil {
		return nil, err
	}
	s.k.WatchKubeConfig(s.reloadKubernetesClient)
	return s, nil
}

func (s *Server) reloadKubernetesClient() error {
	k, err := kubernetes.NewKubernetes(s.configuration.kubeConfig)
	if err != nil {
		return err
	}
	s.k = k
	s.server.SetTools(slices.Concat(
		s.initConfiguration(),
	)...)
	return nil
}

func (s *Server) ServeSse(baseUrl string) *server.SSEServer {
	options := make([]server.SSEOption, 0)
	if baseUrl != "" {
		options = append(options, server.WithBaseURL(baseUrl))
	}
	return server.NewSSEServer(s.server, options...)
}

func (s *Server) Run(SSEBaseURL string, SSEPort int, stopCh <-chan struct{}) {
	log.Infow("Starting mcp kubernetes server")
	if SSEPort > 0 {
		sseServer := s.ServeSse(SSEBaseURL)
		defer func() { _ = sseServer.Shutdown(genericapiserver.SetupSignalContext()) }()
		log.Infow("SSE server starting", "port", SSEPort)
		if err := sseServer.Start(fmt.Sprintf(":%d", SSEPort)); err != nil {
			log.Errorw(err, "Failed to start SSE server")
			return
		}
	}
	if err := server.ServeStdio(s.server); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
	<-stopCh
	s.Stop()
}

func (s *Server) Stop() {
	if s.k != nil {
		s.k.Close()
	}
}
