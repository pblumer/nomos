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
		Instructions: "Nomos Cosmos-Workspace assistant. Use these tools to explore domains, services, decisions and validate the cosmos.",
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_cosmos_info",
		Description: "Get top-level cosmos metadata (name, ID, version, domain count, service count).",
	}, toolCosmosInfo(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_list_domains",
		Description: "List all domains in the cosmos with their canonical name, owner, status and service count.",
	}, toolListDomains(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_get_domain",
		Description: "Get detailed information about a domain including its services and decisions.",
	}, toolGetDomain(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_list_services",
		Description: "List services for a specific domain.",
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
		Description: "List decision tables for a specific domain.",
	}, toolListDecisions(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_evaluate_decision",
		Description: "Evaluate a decision table with given input values and return the output.",
	}, toolEvaluateDecision(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_list_blueprints",
		Description: "List all product and service blueprints in the cosmos catalog.",
	}, toolListBlueprints(cosmosPath))

	return srv
}

// ── tool input/output types ──────────────────────────────────────────────────

type emptyIn struct{}

type cosmosInfoOut struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Owner        string `json:"owner"`
	DomainCount  int    `json:"domain_count"`
	ServiceCount int    `json:"service_count"`
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
			DomainCount: c.DomainCount, ServiceCount: c.ServiceCount,
		}
		return textResult(out)
	}
}

type listDomainsOut struct {
	Domains []domainSummary `json:"domains"`
	Count   int             `json:"count"`
}
type domainSummary struct {
	Canonical    string `json:"canonical"`
	DisplayName  string `json:"display_name"`
	Owner        string `json:"owner"`
	Status       string `json:"status"`
	ServiceCount int    `json:"service_count"`
	ProductCount int    `json:"product_count"`
}

func toolListDomains(path string) func(context.Context, *mcp.CallToolRequest, emptyIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
		dto, err := app.ListDomains(path)
		if err != nil {
			return nil, nil, err
		}
		summaries := make([]domainSummary, 0, len(dto.Domains))
		for _, d := range dto.Domains {
			summaries = append(summaries, domainSummary{
				Canonical: d.Canonical, DisplayName: d.DisplayName,
				Owner: d.Owner, Status: d.Status,
				ServiceCount: d.ServiceCount, ProductCount: d.ProductCount,
			})
		}
		return textResult(listDomainsOut{Domains: summaries, Count: len(summaries)})
	}
}

type getDomainIn struct {
	Domain string `json:"domain" jsonschema:"canonical domain name, e.g. identity.blumer.cloud"`
}
type getDomainOut struct {
	Canonical string           `json:"canonical"`
	Owner     string           `json:"owner"`
	Status    string           `json:"status"`
	Services  []serviceSummary `json:"services,omitempty"`
	Decisions []decisionLine   `json:"decisions,omitempty"`
}
type serviceSummary struct {
	Name         string   `json:"name"`
	Owner        string   `json:"owner"`
	Status       string   `json:"status"`
	Capabilities []string `json:"capabilities,omitempty"`
}
type decisionLine struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	HasDMN bool   `json:"has_dmn"`
}

func toolGetDomain(path string) func(context.Context, *mcp.CallToolRequest, getDomainIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getDomainIn) (*mcp.CallToolResult, any, error) {
		d, err := app.GetDomain(path, in.Domain)
		if err != nil {
			return nil, nil, fmt.Errorf("domain %q not found: %w", in.Domain, err)
		}
		out := getDomainOut{Canonical: d.Canonical, Owner: d.Owner, Status: d.Status}
		for _, s := range d.Services {
			out.Services = append(out.Services, serviceSummary{
				Name: s.Name, Owner: s.Owner, Status: s.Status, Capabilities: s.Capabilities,
			})
		}
		for _, dec := range d.Decisions {
			out.Decisions = append(out.Decisions, decisionLine{
				ID: dec.ID, Name: dec.Name, Status: dec.Status, HasDMN: dec.HasDMN,
			})
		}
		return textResult(out)
	}
}

type listServicesIn struct {
	Domain string `json:"domain" jsonschema:"canonical domain name"`
}

func toolListServices(path string) func(context.Context, *mcp.CallToolRequest, listServicesIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listServicesIn) (*mcp.CallToolResult, any, error) {
		dto, err := app.ListServices(path, in.Domain)
		if err != nil {
			return nil, nil, err
		}
		return textResult(dto)
	}
}

type getServiceIn struct {
	Domain  string `json:"domain"  jsonschema:"canonical domain name"`
	Service string `json:"service" jsonschema:"service name"`
}

func toolGetService(path string) func(context.Context, *mcp.CallToolRequest, getServiceIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getServiceIn) (*mcp.CallToolResult, any, error) {
		s, err := app.GetService(path, in.Domain, in.Service)
		if err != nil {
			return nil, nil, fmt.Errorf("service %q in domain %q not found: %w", in.Service, in.Domain, err)
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

type listDecisionsIn struct {
	Domain string `json:"domain" jsonschema:"canonical domain name"`
}

func toolListDecisions(path string) func(context.Context, *mcp.CallToolRequest, listDecisionsIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listDecisionsIn) (*mcp.CallToolResult, any, error) {
		dto, err := app.ListDecisions(path, in.Domain)
		if err != nil {
			return nil, nil, err
		}
		return textResult(dto)
	}
}

type evaluateDecisionIn struct {
	Domain string         `json:"domain"  jsonschema:"canonical domain name"`
	ID     string         `json:"id"      jsonschema:"decision ID"`
	Inputs map[string]any `json:"inputs"  jsonschema:"input values as key/value pairs"`
}

func toolEvaluateDecision(path string) func(context.Context, *mcp.CallToolRequest, evaluateDecisionIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in evaluateDecisionIn) (*mcp.CallToolResult, any, error) {
		result, err := app.EvaluateDecision(path, in.Domain, in.ID, app.EvaluateDecisionRequest{Inputs: in.Inputs})
		if err != nil {
			return nil, nil, fmt.Errorf("evaluate decision %q in %q: %w", in.ID, in.Domain, err)
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
