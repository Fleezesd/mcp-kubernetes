package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initConfiguration() []server.ServerTool {
	tools := []server.ServerTool{
		{Tool: mcp.NewTool("configuration_view",
			mcp.WithDescription("Get the current Kubernetes configuration content as a kubeconfig YAML"),
			mcp.WithBoolean("minified", mcp.Description("Return a minified version of the configuration. "+
				"If set to true, keeps only the current-context and the relevant pieces of the configuration for that context. "+
				"If set to false, all contexts, clusters, auth-infos, and users are returned in the configuration. "+
				"(Optional, default true)"))),
			Handler: s.configurationView},
	}
	return tools
}

func (s *Server) configurationView(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	minify := true
	minified := ctr.Params.Arguments["minified"]
	if _, ok := minified.(bool); ok {
		minify = minified.(bool)
	}
	ret, err := s.k.ConfigurationView(minify)
	if err != nil {
		err = fmt.Errorf("failed to get configuration: %v", err)
	}
	return NewTextResult(ret, err), nil
}

func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: err.Error(),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}
