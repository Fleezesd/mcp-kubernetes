package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initNamespace() []server.ServerTool {
	tools := []server.ServerTool{
		{
			Tool: mcp.NewTool("namespace_list",
				mcp.WithDescription("List all the kubernetes namespaces in the current cluster")),
			Handler: s.namespacesList,
		},
	}
	return tools
}

func (s *Server) namespacesList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := s.k.NamespacesList(ctx)
	if err != nil {
		err = fmt.Errorf("failed to list namespaces: %v", err)
	}
	return NewTextResult(result, err), nil
}
