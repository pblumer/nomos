package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/model"
)

// registerWriteTools adds all mutating MCP tools to srv.
func registerWriteTools(srv *mcp.Server, cosmosPath string) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_service_add",
		Description: "Create a new service in the cosmos.",
	}, toolServiceAdd(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_product_create",
		Description: "Create a new product offering (product blueprint).",
	}, toolProductCreate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_process_create",
		Description: "Create a BPMN process for an existing product. Returns the new process ID and its initial state.",
	}, toolProcessCreate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name: "nomos_process_step_add",
		Description: "Add an ordered step to a process. " +
			"For business_rule_task steps with a decision_ref an exclusive_gateway step is automatically appended " +
			"(use gateway_name to set its label; defaults to '<step name>?'). " +
			"The gateway branches (Ja/Nein with condition expressions) are resolved from the decision outputs at BPMN render time — " +
			"no separate gateway tool call is needed.",
	}, toolProcessStepAdd(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_decision_create",
		Description: "Create a new decision table (DMN).",
	}, toolDecisionCreate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_instance_create",
		Description: "Create a product or service instance from a blueprint reference.",
	}, toolInstanceCreate(cosmosPath))
}

// ── service_add ──────────────────────────────────────────────────────────────

type serviceAddIn struct {
	Name  string `json:"name"            jsonschema:"service name (kebab-case)"`
	Owner string `json:"owner,omitempty" jsonschema:"owning team or person"`
}

func toolServiceAdd(path string) func(context.Context, *mcp.CallToolRequest, serviceAddIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in serviceAddIn) (*mcp.CallToolResult, any, error) {
		if in.Name == "" {
			return nil, nil, fmt.Errorf("name is required")
		}
		dto, err := app.AddService(path, in.Name, in.Owner, false)
		if err != nil {
			return nil, nil, fmt.Errorf("service_add %q: %w", in.Name, err)
		}
		return textResult(map[string]any{
			"id":     dto.ID,
			"name":   dto.Name,
			"owner":  dto.Owner,
			"status": dto.Status,
			"path":   dto.Path,
		})
	}
}

// ── product_create ───────────────────────────────────────────────────────────

type productCreateIn struct {
	ID      string `json:"id,omitempty"      jsonschema:"optional product ID; leave empty to auto-generate per ADR-0020 (e.g. PRD_A7K3M2)"`
	Name    string `json:"name"              jsonschema:"human-readable product name"`
	Owner   string `json:"owner,omitempty"   jsonschema:"owning team or person"`
	Summary string `json:"summary,omitempty" jsonschema:"one-line description of the product"`
}

func toolProductCreate(path string) func(context.Context, *mcp.CallToolRequest, productCreateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in productCreateIn) (*mcp.CallToolResult, any, error) {
		if in.Name == "" {
			return nil, nil, fmt.Errorf("name is required")
		}
		dto, err := app.CreateProductOffering(path, app.CreateProductOfferingRequest{
			ID:      in.ID,
			Name:    in.Name,
			Owner:   in.Owner,
			Summary: in.Summary,
			Version: "0.1.0",
			Status:  "draft",
		})
		if err != nil {
			return nil, nil, fmt.Errorf("product_create %q: %w", in.ID, err)
		}
		return textResult(map[string]any{
			"id":      dto.ID,
			"name":    dto.Name,
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
	TaskType      string `json:"task_type,omitempty"      jsonschema:"task type: service_task | business_rule_task | user_task | manual_task | script_task | exclusive_gateway. For business_rule_task + decision_ref a gateway is auto-appended; you rarely need to add exclusive_gateway manually."`
	DecisionRef   string `json:"decision_ref,omitempty"   jsonschema:"decision ID referenced by a business_rule_task (e.g. dec-dns-format-check). Triggers automatic exclusive_gateway creation after this step."`
	GatewayName   string `json:"gateway_name,omitempty"   jsonschema:"label for the auto-created exclusive gateway that follows a business_rule_task; leave empty to use '<step name>?'"`
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

		result := map[string]any{
			"step_id":        added.ID,
			"name":           added.Name,
			"task_type":      added.TaskType,
			"service_ref":    added.ServiceRef,
			"capability_ref": added.CapabilityRef,
			"method":         added.Method,
			"decision_ref":   added.DecisionRef,
			"total_steps":    len(dto.Steps),
		}

		// Auto-append exclusive gateway for business_rule_task + decision_ref.
		// Branch conditions (Ja/Nein) are resolved from the decision's boolean
		// output at BPMN render time — no manual gateway call needed.
		if isBusinessRuleTaskType(in.TaskType) && strings.TrimSpace(in.DecisionRef) != "" {
			gwName := strings.TrimSpace(in.GatewayName)
			if gwName == "" {
				gwName = strings.TrimSpace(in.Name) + "?"
			}
			gwDTO, gwErr := app.AddProcessStep(path, in.ProcessID, app.UpsertProcessStepRequest{
				Name:     gwName,
				TaskType: "exclusive_gateway",
				Required: true,
			})
			if gwErr == nil && len(gwDTO.Steps) > 0 {
				gw := gwDTO.Steps[len(gwDTO.Steps)-1]
				result["gateway_step_id"] = gw.ID
				result["gateway_name"] = gw.Name
				result["total_steps"] = len(gwDTO.Steps)
				result["gateway_note"] = "exclusive_gateway auto-created; Ja/Nein branches with conditionExpression resolved from decision outputs at BPMN render time"
			}
		}

		return textResult(result)
	}
}

func isBusinessRuleTaskType(t string) bool {
	t = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(t, "_", ""), "bpmn:", ""))
	return t == "businessruletask"
}

// ── decision_create ──────────────────────────────────────────────────────────

type decisionCreateIn struct {
	ID      string `json:"id,omitempty"      jsonschema:"optional decision ID; leave empty to auto-generate per ADR-0020 (e.g. DEC_T4R7W3)"`
	Name    string `json:"name"              jsonschema:"human-readable decision name"`
	Owner   string `json:"owner,omitempty"   jsonschema:"owning team or person"`
	Summary string `json:"summary,omitempty" jsonschema:"short description of what this decision evaluates"`
}

func toolDecisionCreate(path string) func(context.Context, *mcp.CallToolRequest, decisionCreateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in decisionCreateIn) (*mcp.CallToolResult, any, error) {
		if in.Name == "" {
			return nil, nil, fmt.Errorf("name is required")
		}
		dto, err := app.CreateDecision(path, app.CreateDecisionRequest{
			ID:      in.ID,
			Name:    in.Name,
			Owner:   in.Owner,
			Summary: in.Summary,
			Version: "0.1.0",
			Status:  "draft",
		})
		if err != nil {
			return nil, nil, fmt.Errorf("decision_create %q: %w", in.ID, err)
		}
		return textResult(map[string]any{
			"id":      dto.ID,
			"name":    dto.Name,
			"status":  dto.Status,
			"has_dmn": dto.HasDMN,
		})
	}
}

// ── instance_create ──────────────────────────────────────────────────────────

type instanceCreateIn struct {
	ID           string `json:"id,omitempty"          jsonschema:"optional instance ID; leave empty to auto-generate per ADR-0020 (e.g. PRI_M5Q8N4)"`
	BlueprintRef string `json:"blueprint_ref"         jsonschema:"blueprint ID this instance is based on"`
	Type         string `json:"type,omitempty"        jsonschema:"product_instance or service_instance (default: product_instance)"`
	Name         string `json:"name,omitempty"        jsonschema:"human-readable name for this instance"`
	Owner        string `json:"owner,omitempty"       jsonschema:"owning team or person"`
}

func toolInstanceCreate(path string) func(context.Context, *mcp.CallToolRequest, instanceCreateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in instanceCreateIn) (*mcp.CallToolResult, any, error) {
		if in.BlueprintRef == "" {
			return nil, nil, fmt.Errorf("blueprint_ref is required")
		}
		instType := in.Type
		if instType == "" {
			instType = "product_instance"
		}
		if instType != "product_instance" && instType != "service_instance" {
			return nil, nil, fmt.Errorf("type must be product_instance or service_instance")
		}
		id := strings.TrimSpace(in.ID)
		if id == "" {
			generated, err := idgen.NewForType(instType)
			if err != nil {
				return nil, nil, fmt.Errorf("instance_create: id generation: %w", err)
			}
			id = generated
		}
		name := in.Name
		if name == "" {
			name = id
		}
		if err := app.CreateInstance(path, model.Instance{
			ID:           id,
			Type:         instType,
			Name:         name,
			Owner:        in.Owner,
			BlueprintRef: in.BlueprintRef,
			Status:       "active",
		}); err != nil {
			return nil, nil, fmt.Errorf("instance_create %q: %w", id, err)
		}
		return textResult(map[string]any{
			"id":            id,
			"type":          instType,
			"name":          name,
			"blueprint_ref": in.BlueprintRef,
			"status":        "active",
		})
	}
}
