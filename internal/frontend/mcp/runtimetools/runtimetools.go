// Package runtimetools is the runtime group of sysml-mcp tools: the one-shot
// operations of the sysml-grpc service (parse, evaluate, instantiate, execute,
// verify, query, render, convert), each tool one RPC with the RPC's contract.
package runtimetools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	sysmlgrpc "github.com/Open-MBEE/OpenSysML/internal/frontend/grpc"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/mcp/toolkit"
)

// ServerInfoInput takes no arguments.
type ServerInfoInput struct{}

// ServerInfo is what the service reports about itself.
type ServerInfo struct {
	Version      string   `json:"version" jsonschema:"build version of the server, informational only"`
	Capabilities []string `json:"capabilities" jsonschema:"named capabilities this build supports; each is a stable string a caller can require"`
}

// Register adds the runtime tools to s over svc.
func Register(s *mcp.Server, svc *sysmlgrpc.Service) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "server_info",
		Description: "Report the OpenSysML runtime's version and the named capabilities this build supports.",
		Annotations: toolkit.ReadOnly("Server info"),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ ServerInfoInput) (*mcp.CallToolResult, ServerInfo, error) {
		resp, err := svc.GetServerInfo(ctx, &pb.ServerInfoRequest{})
		if err != nil {
			return nil, ServerInfo{}, err
		}
		caps := resp.GetCapabilities()
		if caps == nil {
			caps = []string{}
		}
		return nil, ServerInfo{Version: resp.GetVersion(), Capabilities: caps}, nil
	})
}
