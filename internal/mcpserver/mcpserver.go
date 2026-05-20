// Package mcpserver implements a Model Context Protocol server for Nomos.
// It exposes Nomos domains, services, decisions and validation as MCP tools.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/version"
)

// New builds an MCP server wired to the given cosmos workspace path.
func New(cosmosPath string) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "nomos",
		Version: version.Version,
	}, &mcp.ServerOptions{
		Instructions: "Nomos Cosmos-Workspace assistant. Read tools: explore services, decisions, blueprints, validate. Write tools: add services, create products/processes/decisions/instances, add process steps with capability and decision references.",
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_cosmos_info",
		Description: "Get top-level cosmos metadata (name, ID, version, service count, decision count).",
	}, toolCosmosInfo(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_list_services",
		Description: "List all services in the cosmos.",
	}, toolListServices(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_get_service",
		Description: "Get detailed information about a service including capabilities and methods.",
	}, toolGetService(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_validate",
		Description: "Validate the entire cosmos and return a list of findings (errors and warnings).",
	}, toolValidate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_list_decisions",
		Description: "List all decision tables in the cosmos.",
	}, toolListDecisions(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_evaluate_decision",
		Description: "Evaluate a decision table with given input values and return the output.",
	}, toolEvaluateDecision(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_list_blueprints",
		Description: "List all product and service blueprints in the cosmos catalog.",
	}, toolListBlueprints(cosmosPath))

	registerWriteTools(srv, cosmosPath)
	registerCapabilityMethodTools(srv, cosmosPath)

	return srv
}

// ── tool input/output types ──────────────────────────────────────────────────

type emptyIn struct{}

type cosmosInfoOut struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	Owner         string `json:"owner"`
	ServiceCount  int    `json:"service_count"`
	DecisionCount int    `json:"decision_count"`
}

func toolCosmosInfo(path string) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
		c, err := app.GetCosmos(path)
		if err != nil {
			return nil, nil, err
		}
		out := cosmosInfoOut{
			ID: c.ID, Name: c.Name, Version: c.Version,
			Status: c.Status, Owner: c.Owner,
			ServiceCount: c.ServiceCount, DecisionCount: c.DecisionCount,
		}
		return textResult(out)
	}
}

func toolListServices(path string) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
		dto, err := app.ListServices(path)
		if err != nil {
			return nil, nil, err
		}
		return textResult(dto)
	}
}

type getServiceIn struct {
	Service string `json:"service" jsonschema:"service name"`
}

func toolGetService(path string) func(context.Context, *mcp.CallToolRequest, getServiceIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getServiceIn) (*mcp.CallToolResult, any, error) {
		s, err := app.GetService(path, in.Service)
		if err != nil {
			return nil, nil, fmt.Errorf("service %q not found: %w", in.Service, err)
		}
		return textResult(s)
	}
}

type validateOut struct {
	Status   string           `json:"status"`
	Findings []findingSummary `json:"findings,omitempty"`
	Count    int              `json:"count"`
}
type findingSummary struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
}

func toolValidate(path string) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
		result, err := app.ValidateCosmos(path)
		if err != nil {
			return nil, nil, err
		}
		out := validateOut{Status: result.Status}
		for _, f := range result.Findings {
			out.Findings = append(out.Findings, findingSummary{
				Code: f.Code, Severity: f.Severity, Message: f.Message, Path: f.Path,
			})
		}
		out.Count = len(out.Findings)
		return textResult(out)
	}
}

func toolListDecisions(path string) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
		dto, err := app.ListDecisions(path)
		if err != nil {
			return nil, nil, err
		}
		return textResult(dto)
	}
}

type evaluateDecisionIn struct {
	ID     string         `json:"id"      jsonschema:"decision ID"`
	Inputs map[string]any `json:"inputs"  jsonschema:"input values as key/value pairs"`
}

func toolEvaluateDecision(path string) func(context.Context, *mcp.CallToolRequest, evaluateDecisionIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in evaluateDecisionIn) (*mcp.CallToolResult, any, error) {
		result, err := app.EvaluateDecision(path, in.ID, app.EvaluateDecisionRequest{Inputs: in.Inputs})
		if err != nil {
			return nil, nil, fmt.Errorf("evaluate decision %q: %w", in.ID, err)
		}
		return textResult(result)
	}
}

type listBlueprintsOut struct {
	Blueprints []blueprintLine `json:"blueprints"`
	Count      int             `json:"count"`
}
type blueprintLine struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Owner  string `json:"owner"`
}

func toolListBlueprints(path string) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
		dto, err := app.ListBlueprints(path)
		if err != nil {
			return nil, nil, err
		}
		out := listBlueprintsOut{}
		for _, b := range dto.Blueprints {
			out.Blueprints = append(out.Blueprints, blueprintLine{
				ID: b.ID, Type: b.Type, Name: b.Name, Status: b.Status, Owner: b.Owner,
			})
		}
		out.Count = len(out.Blueprints)
		return textResult(out)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func textResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.TrimSpace(string(b))},
		},
	}, nil, nil
}
