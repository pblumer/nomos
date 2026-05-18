package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
)

// registerWriteTools adds all mutating MCP tools to srv.
func registerWriteTools(srv *mcp.Server, cosmosPath string) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_domain_add",
		Description: "Create a new top-level domain in the cosmos. The dns parameter must be a valid DNS-style namespace (e.g. 'payments.acme.com').",
	}, toolDomainAdd(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_service_add",
		Description: "Create a new service inside an existing domain.",
	}, toolServiceAdd(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_product_create",
		Description: "Create a new product offering (product blueprint) in a domain.",
	}, toolProductCreate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_process_create",
		Description: "Create a BPMN process for an existing product. Returns the new process ID and its initial state.",
	}, toolProcessCreate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_process_step_add",
		Description: "Add an ordered step to a process. A step maps to a service method call (task) and optionally references a service capability (= BPMN collaboration boundary).",
	}, toolProcessStepAdd(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_decision_create",
		Description: "Create a new decision table (DMN) inside a domain.",
	}, toolDecisionCreate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_instance_create",
		Description: "Create a product or service instance from a blueprint reference.",
	}, toolInstanceCreate(cosmosPath))
}

// ── domain_add ───────────────────────────────────────────────────────────────

type domainAddIn struct {
	DNS   string `json:"dns"             jsonschema:"DNS-style namespace, e.g. payments.acme.com"`
	Owner string `json:"owner,omitempty" jsonschema:"owning team or person"`
}

func toolDomainAdd(path string) func(context.Context, *mcp.CallToolRequest, domainAddIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in domainAddIn) (*mcp.CallToolResult, any, error) {
		if in.DNS == "" {
			return nil, nil, fmt.Errorf("dns is required")
		}
		dto, err := app.AddDomain(path, in.DNS, in.Owner, false)
		if err != nil {
			return nil, nil, fmt.Errorf("domain_add %q: %w", in.DNS, err)
		}
		return textResult(map[string]any{
			"canonical": dto.Canonical,
			"name":      dto.DisplayName,
			"owner":     dto.Owner,
			"status":    dto.Status,
			"git_path":  dto.GitPath,
		})
	}
}

// ── service_add ──────────────────────────────────────────────────────────────

type serviceAddIn struct {
	Domain string `json:"domain"          jsonschema:"canonical domain name"`
	Name   string `json:"name"            jsonschema:"service name (kebab-case)"`
	Owner  string `json:"owner,omitempty" jsonschema:"owning team or person"`
}

func toolServiceAdd(path string) func(context.Context, *mcp.CallToolRequest, serviceAddIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in serviceAddIn) (*mcp.CallToolResult, any, error) {
		if in.Domain == "" || in.Name == "" {
			return nil, nil, fmt.Errorf("domain and name are required")
		}
		dto, err := app.AddService(path, in.Domain, in.Name, in.Owner, false)
		if err != nil {
			return nil, nil, fmt.Errorf("service_add %q in %q: %w", in.Name, in.Domain, err)
		}
		return textResult(map[string]any{
			"name":   dto.Name,
			"domain": dto.Domain,
			"owner":  dto.Owner,
			"status": dto.Status,
			"path":   dto.Path,
		})
	}
}

// ── product_create ───────────────────────────────────────────────────────────

type productCreateIn struct {
	Domain  string `json:"domain"            jsonschema:"canonical domain name"`
	ID      string `json:"id"                jsonschema:"unique product ID (kebab-case)"`
	Name    string `json:"name"              jsonschema:"human-readable product name"`
	Owner   string `json:"owner,omitempty"   jsonschema:"owning team or person"`
	Summary string `json:"summary,omitempty" jsonschema:"one-line description of the product"`
}

func toolProductCreate(path string) func(context.Context, *mcp.CallToolRequest, productCreateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in productCreateIn) (*mcp.CallToolResult, any, error) {
		if in.Domain == "" || in.ID == "" || in.Name == "" {
			return nil, nil, fmt.Errorf("domain, id and name are required")
		}
		dto, err := app.CreateProductOffering(path, in.Domain, app.CreateProductOfferingRequest{
			ID:           in.ID,
			Name:         in.Name,
			Owner:        in.Owner,
			Summary:      in.Summary,
			OwningDomain: in.Domain,
			Version:      "0.1.0",
			Status:       "draft",
		})
		if err != nil {
			return nil, nil, fmt.Errorf("product_create %q: %w", in.ID, err)
		}
		return textResult(map[string]any{
			"id":      dto.ID,
			"name":    dto.Name,
			"domain":  dto.OwningDomain,
			"status":  dto.Status,
			"version": dto.Version,
		})
	}
}

// ── process_create ───────────────────────────────────────────────────────────

type processCreateIn struct {
	ProductID string `json:"product_id"        jsonschema:"product ID the process belongs to"`
	Name      string `json:"name"              jsonschema:"process name"`
	Summary   string `json:"summary,omitempty" jsonschema:"short description of what the process does"`
}

func toolProcessCreate(path string) func(context.Context, *mcp.CallToolRequest, processCreateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in processCreateIn) (*mcp.CallToolResult, any, error) {
		if in.ProductID == "" || in.Name == "" {
			return nil, nil, fmt.Errorf("product_id and name are required")
		}
		dto, err := app.CreateProductProcess(path, in.ProductID, app.CreateProcessRequest{
			Name:    in.Name,
			Summary: in.Summary,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("process_create for product %q: %w", in.ProductID, err)
		}
		return textResult(map[string]any{
			"id":         dto.ID,
			"name":       dto.Name,
			"status":     dto.Status,
			"product_id": dto.RelatedProduct,
			"bpmn_file":  dto.BPMN.File,
		})
	}
}

// ── process_step_add ─────────────────────────────────────────────────────────

type processStepAddIn struct {
	ProcessID     string `json:"process_id"               jsonschema:"process ID to add the step to"`
	Name          string `json:"name"                     jsonschema:"step name"`
	ServiceRef    string `json:"service_ref,omitempty"    jsonschema:"service reference, e.g. core.nomos/catalog-manager"`
	CapabilityRef string `json:"capability_ref,omitempty" jsonschema:"capability ID on the service (= collaboration boundary)"`
	Method        string `json:"method,omitempty"         jsonschema:"service method name to invoke"`
	TaskType      string `json:"task_type,omitempty"      jsonschema:"task type: service_task | business_rule_task | user_task | manual_task | script_task"`
	DecisionRef   string `json:"decision_ref,omitempty"   jsonschema:"decision table ID for business_rule_task steps"`
	Role          string `json:"role,omitempty"           jsonschema:"executing role or team"`
	Required      bool   `json:"required,omitempty"       jsonschema:"whether this step is mandatory"`
	Notes         string `json:"notes,omitempty"          jsonschema:"additional notes"`
}

func toolProcessStepAdd(path string) func(context.Context, *mcp.CallToolRequest, processStepAddIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in processStepAddIn) (*mcp.CallToolResult, any, error) {
		if in.ProcessID == "" || in.Name == "" {
			return nil, nil, fmt.Errorf("process_id and name are required")
		}
		dto, err := app.AddProcessStep(path, in.ProcessID, app.UpsertProcessStepRequest{
			Name:          in.Name,
			ServiceRef:    in.ServiceRef,
			CapabilityRef: in.CapabilityRef,
			Method:        in.Method,
			TaskType:      in.TaskType,
			DecisionRef:   in.DecisionRef,
			Role:          in.Role,
			Required:      in.Required,
			Notes:         in.Notes,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("process_step_add to %q: %w", in.ProcessID, err)
		}
		added := dto.Steps[len(dto.Steps)-1]
		return textResult(map[string]any{
			"step_id":        added.ID,
			"name":           added.Name,
			"task_type":      added.TaskType,
			"service_ref":    added.ServiceRef,
			"capability_ref": added.CapabilityRef,
			"method":         added.Method,
			"decision_ref":   added.DecisionRef,
			"total_steps":    len(dto.Steps),
		})
	}
}

// ── decision_create ──────────────────────────────────────────────────────────

type decisionCreateIn struct {
	Domain  string `json:"domain"            jsonschema:"canonical domain name"`
	ID      string `json:"id"                jsonschema:"unique decision ID (kebab-case)"`
	Name    string `json:"name"              jsonschema:"human-readable decision name"`
	Owner   string `json:"owner,omitempty"   jsonschema:"owning team or person"`
	Summary string `json:"summary,omitempty" jsonschema:"short description of what this decision evaluates"`
}

func toolDecisionCreate(path string) func(context.Context, *mcp.CallToolRequest, decisionCreateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in decisionCreateIn) (*mcp.CallToolResult, any, error) {
		if in.Domain == "" || in.ID == "" || in.Name == "" {
			return nil, nil, fmt.Errorf("domain, id and name are required")
		}
		dto, err := app.CreateDecision(path, in.Domain, app.CreateDecisionRequest{
			ID:      in.ID,
			Name:    in.Name,
			Owner:   in.Owner,
			Summary: in.Summary,
			Version: "0.1.0",
			Status:  "draft",
		})
		if err != nil {
			return nil, nil, fmt.Errorf("decision_create %q in %q: %w", in.ID, in.Domain, err)
		}
		return textResult(map[string]any{
			"id":      dto.ID,
			"name":    dto.Name,
			"domain":  in.Domain,
			"status":  dto.Status,
			"has_dmn": dto.HasDMN,
		})
	}
}

// ── instance_create ──────────────────────────────────────────────────────────

type instanceCreateIn struct {
	ID           string `json:"id"                    jsonschema:"unique instance ID (kebab-case)"`
	BlueprintRef string `json:"blueprint_ref"         jsonschema:"blueprint ID this instance is based on"`
	Type         string `json:"type,omitempty"        jsonschema:"product_instance or service_instance (default: product_instance)"`
	Name         string `json:"name,omitempty"        jsonschema:"human-readable name for this instance"`
	Owner        string `json:"owner,omitempty"       jsonschema:"owning team or person"`
}

func toolInstanceCreate(path string) func(context.Context, *mcp.CallToolRequest, instanceCreateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in instanceCreateIn) (*mcp.CallToolResult, any, error) {
		if in.ID == "" || in.BlueprintRef == "" {
			return nil, nil, fmt.Errorf("id and blueprint_ref are required")
		}
		instType := in.Type
		if instType == "" {
			instType = "product_instance"
		}
		if instType != "product_instance" && instType != "service_instance" {
			return nil, nil, fmt.Errorf("type must be product_instance or service_instance")
		}
		name := in.Name
		if name == "" {
			name = in.ID
		}
		err := app.CreateInstance(path, model.Instance{
			ID:           in.ID,
			Type:         instType,
			Name:         name,
			Owner:        in.Owner,
			BlueprintRef: in.BlueprintRef,
			Status:       "active",
		})
		if err != nil {
			return nil, nil, fmt.Errorf("instance_create %q: %w", in.ID, err)
		}
		return textResult(map[string]any{
			"id":            in.ID,
			"type":          instType,
			"name":          name,
			"blueprint_ref": in.BlueprintRef,
			"status":        "active",
		})
	}
}
