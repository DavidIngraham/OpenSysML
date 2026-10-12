package runtimetools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	pb "github.com/Open-MBEE/OpenSysML/api/proto"
	"github.com/Open-MBEE/OpenSysML/internal/frontend/mcp/toolkit"
)

// ServerInfoInput takes no arguments.
type ServerInfoInput struct{}

// ServerInfo is what the service reports about itself.
type ServerInfo struct {
	Version      string   `json:"version" jsonschema:"build version of the server, informational only"`
	Capabilities []string `json:"capabilities" jsonschema:"named capabilities this build supports; each is a stable string a caller can require"`
}

// ListEnginesInput takes no arguments.
type ListEnginesInput struct{}

func (g *group) registerInfo(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "server_info",
		Description: "Report the OpenSysML runtime's version and the named capabilities this build supports.",
		Annotations: toolkit.ReadOnly("Server info"),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ ServerInfoInput) (*mcp.CallToolResult, ServerInfo, error) {
		resp, err := g.svc.GetServerInfo(ctx, &pb.ServerInfoRequest{})
		if err != nil {
			return nil, ServerInfo{}, toolError(err)
		}
		caps := resp.GetCapabilities()
		if caps == nil {
			caps = []string{}
		}
		return nil, ServerInfo{Version: resp.GetVersion(), Capabilities: caps}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "list_engines",
		Description: "List the analysis engines this build registers (the ListEngines RPC, as the CLI's -engines does): each with the " +
			"questions it answers (evaluate, outcomes, holds, sensitive, satisfiable, sweep, compute), the strongest evidence it can " +
			"produce, the bounds it takes, and whether it can run here. The names are what the verification tools' `engine` argument takes; " +
			"empty there means auto. The result is the ListEnginesResponse as protojson writes it.",
		Annotations: toolkit.ReadOnly("List engines"),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ ListEnginesInput) (*mcp.CallToolResult, Result, error) {
		resp, err := g.svc.ListEngines(ctx, &pb.ListEnginesRequest{})
		if err != nil {
			return nil, nil, toolError(err)
		}
		var b strings.Builder
		for _, e := range resp.GetEngines() {
			state := "ready"
			if !e.GetReady() {
				state = "unavailable: " + e.GetUnavailable()
			}
			fmt.Fprintf(&b, "%s (%s) answers %s; authority %s; %s\n", e.GetName(), e.GetKind(), strings.Join(e.GetAnswers(), ", "), e.GetAuthority(), state)
		}
		return answer(resp, model{}, b.String())
	})
}
