package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
)

// registerCapabilityMethodTools adds MCP tools for capability and method CRUD.
// Methods are how a service exposes work; capabilities group methods into a
// user-facing contract. Use these together: define methods first, then group
// them under capabilities via method_refs.
func registerCapabilityMethodTools(srv *mcp.Server, cosmosPath string) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_capability_add",
		Description: "Add a new capability to a service. method_refs links it to existing methods so consumers see which methods implement the capability.",
	}, toolCapabilityAdd(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_capability_update",
		Description: "Update fields of an existing capability. Only provided fields are changed; omit method_refs/data_object_refs to leave them as-is, pass an empty array to clear them.",
	}, toolCapabilityUpdate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_capability_delete",
		Description: "Delete a capability from a service by ID. Linked methods themselves are not removed.",
	}, toolCapabilityDelete(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_method_add",
		Description: "Add a new method to a service. Optionally set summary, http_method and path in the same call.",
	}, toolMethodAdd(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_method_update",
		Description: "Update fields of an existing method on a service. Only non-empty fields are written.",
	}, toolMethodUpdate(cosmosPath))

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "nomos_method_delete",
		Description: "Delete a method from a service. References to it in capability.method_refs are cleaned up automatically.",
	}, toolMethodDelete(cosmosPath))
}

// ── capability_add ───────────────────────────────────────────────────────────

type capabilityAddIn struct {
	Service        string   `json:"service"                    jsonschema:"service name"`
	Name           string   `json:"name"                       jsonschema:"human-readable capability name"`
	ID             string   `json:"id,omitempty"               jsonschema:"optional capability ID (e.g. cap-blueprint-crud); derived from name if empty"`
	Summary        string   `json:"summary,omitempty"          jsonschema:"one-line description of what the capability does"`
	Stability      string   `json:"stability,omitempty"        jsonschema:"draft | experimental | stable | deprecated"`
	SideEffect     string   `json:"side_effect,omitempty"      jsonschema:"none | read_only | mutates_workspace | network_egress"`
	MethodRefs     []string `json:"method_refs,omitempty"      jsonschema:"names of methods on the same service that implement this capability"`
	DataObjectRefs []string `json:"data_object_refs,omitempty" jsonschema:"IDs of data objects on the same service referenced by this capability"`
}

func toolCapabilityAdd(path string) func(context.Context, *mcp.CallToolRequest, capabilityAddIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in capabilityAddIn) (*mcp.CallToolResult, any, error) {
		if in.Service == "" || in.Name == "" {
			return nil, nil, fmt.Errorf("service and name are required")
		}
		cap := model.ServiceCapability{
			ID:             strings.TrimSpace(in.ID),
			Name:           in.Name,
			Summary:        in.Summary,
			Stability:      in.Stability,
			SideEffect:     in.SideEffect,
			MethodRefs:     in.MethodRefs,
			DataObjectRefs: in.DataObjectRefs,
		}
		dto, err := app.AddServiceCapability(path, in.Service, cap)
		if err != nil {
			return nil, nil, fmt.Errorf("capability_add %q in %s: %w", in.Name, in.Service, err)
		}
		added := findCapability(dto.CapabilityDefs, cap.ID, cap.Name)
		return textResult(map[string]any{
			"id":               added.ID,
			"name":             added.Name,
			"service":          in.Service,
			"stability":        added.Stability,
			"side_effect":      added.SideEffect,
			"method_refs":      added.MethodRefs,
			"data_object_refs": added.DataObjectRefs,
		})
	}
}

// ── capability_update ────────────────────────────────────────────────────────

// Pointers distinguish "field omitted" (nil → keep existing value) from
// "field provided as empty string / empty array" (non-nil → write the value).
type capabilityUpdateIn struct {
	Service        string    `json:"service"                    jsonschema:"service name"`
	CapabilityID   string    `json:"capability_id"              jsonschema:"ID of the capability to update"`
	Name           *string   `json:"name,omitempty"             jsonschema:"new human-readable name"`
	Summary        *string   `json:"summary,omitempty"          jsonschema:"new one-line description"`
	Stability      *string   `json:"stability,omitempty"        jsonschema:"draft | experimental | stable | deprecated"`
	SideEffect     *string   `json:"side_effect,omitempty"      jsonschema:"none | read_only | mutates_workspace | network_egress"`
	MethodRefs     *[]string `json:"method_refs,omitempty"      jsonschema:"replace linked method names; pass [] to clear, omit to keep existing"`
	DataObjectRefs *[]string `json:"data_object_refs,omitempty" jsonschema:"replace linked data object IDs; pass [] to clear, omit to keep existing"`
}

func toolCapabilityUpdate(path string) func(context.Context, *mcp.CallToolRequest, capabilityUpdateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in capabilityUpdateIn) (*mcp.CallToolResult, any, error) {
		if in.Service == "" || in.CapabilityID == "" {
			return nil, nil, fmt.Errorf("service and capability_id are required")
		}
		// Fetch current so we can preserve slice fields the caller omitted.
		current, err := app.GetService(path, in.Service)
		if err != nil {
			return nil, nil, fmt.Errorf("capability_update: %w", err)
		}
		cur := findCapability(current.CapabilityDefs, in.CapabilityID, "")
		if cur.ID == "" {
			return nil, nil, fmt.Errorf("capability_update: capability not found: %s", in.CapabilityID)
		}

		patch := model.ServiceCapability{
			MethodRefs:     cur.MethodRefs,
			DataObjectRefs: cur.DataObjectRefs,
		}
		if in.Name != nil {
			patch.Name = *in.Name
		}
		if in.Summary != nil {
			patch.Summary = *in.Summary
		}
		if in.Stability != nil {
			patch.Stability = *in.Stability
		}
		if in.SideEffect != nil {
			patch.SideEffect = *in.SideEffect
		}
		if in.MethodRefs != nil {
			patch.MethodRefs = *in.MethodRefs
		}
		if in.DataObjectRefs != nil {
			patch.DataObjectRefs = *in.DataObjectRefs
		}

		dto, err := app.UpdateServiceCapability(path, in.Service, in.CapabilityID, patch)
		if err != nil {
			return nil, nil, fmt.Errorf("capability_update %s in %s: %w", in.CapabilityID, in.Service, err)
		}
		updated := findCapability(dto.CapabilityDefs, in.CapabilityID, "")
		return textResult(map[string]any{
			"id":               updated.ID,
			"name":             updated.Name,
			"service":          in.Service,
			"stability":        updated.Stability,
			"side_effect":      updated.SideEffect,
			"method_refs":      updated.MethodRefs,
			"data_object_refs": updated.DataObjectRefs,
		})
	}
}

// ── capability_delete ────────────────────────────────────────────────────────

type capabilityDeleteIn struct {
	Service      string `json:"service"       jsonschema:"service name"`
	CapabilityID string `json:"capability_id" jsonschema:"ID of the capability to delete"`
}

func toolCapabilityDelete(path string) func(context.Context, *mcp.CallToolRequest, capabilityDeleteIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in capabilityDeleteIn) (*mcp.CallToolResult, any, error) {
		if in.Service == "" || in.CapabilityID == "" {
			return nil, nil, fmt.Errorf("service and capability_id are required")
		}
		if _, err := app.RemoveServiceCapability(path, in.Service, in.CapabilityID); err != nil {
			return nil, nil, fmt.Errorf("capability_delete %s in %s: %w", in.CapabilityID, in.Service, err)
		}
		return textResult(map[string]any{
			"deleted": in.CapabilityID,
			"service": in.Service,
		})
	}
}

// ── method_add ───────────────────────────────────────────────────────────────

type methodAddIn struct {
	Service    string `json:"service"               jsonschema:"service name"`
	Name       string `json:"name"                  jsonschema:"method name (snake_case)"`
	Summary    string `json:"summary,omitempty"     jsonschema:"one-line description"`
	HTTPMethod string `json:"http_method,omitempty" jsonschema:"GET | POST | PUT | PATCH | DELETE"`
	Path       string `json:"path,omitempty"        jsonschema:"URI path template, e.g. /resource/{id}"`
}

func toolMethodAdd(path string) func(context.Context, *mcp.CallToolRequest, methodAddIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in methodAddIn) (*mcp.CallToolResult, any, error) {
		if in.Service == "" || in.Name == "" {
			return nil, nil, fmt.Errorf("service and name are required")
		}
		if _, err := app.AddServiceMethod(path, in.Service, in.Name); err != nil {
			return nil, nil, fmt.Errorf("method_add %q in %s: %w", in.Name, in.Service, err)
		}
		// If caller supplied extra fields, follow up with an update.
		if in.Summary != "" || in.HTTPMethod != "" || in.Path != "" {
			if _, err := app.UpdateMethod(path, in.Service, in.Name, model.MethodDefinition{
				Summary:    in.Summary,
				HTTPMethod: in.HTTPMethod,
				Path:       in.Path,
			}); err != nil {
				return nil, nil, fmt.Errorf("method_add %q: created but failed to set fields: %w", in.Name, err)
			}
		}
		return textResult(map[string]any{
			"name":        in.Name,
			"service":     in.Service,
			"summary":     in.Summary,
			"http_method": in.HTTPMethod,
			"path":        in.Path,
		})
	}
}

// ── method_update ────────────────────────────────────────────────────────────

type methodUpdateIn struct {
	Service    string `json:"service"               jsonschema:"service name"`
	Method     string `json:"method"                jsonschema:"name of the method to update"`
	Summary    string `json:"summary,omitempty"     jsonschema:"new one-line description"`
	HTTPMethod string `json:"http_method,omitempty" jsonschema:"GET | POST | PUT | PATCH | DELETE"`
	Path       string `json:"path,omitempty"        jsonschema:"new URI path template"`
}

func toolMethodUpdate(path string) func(context.Context, *mcp.CallToolRequest, methodUpdateIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in methodUpdateIn) (*mcp.CallToolResult, any, error) {
		if in.Service == "" || in.Method == "" {
			return nil, nil, fmt.Errorf("service and method are required")
		}
		dto, err := app.UpdateMethod(path, in.Service, in.Method, model.MethodDefinition{
			Summary:    in.Summary,
			HTTPMethod: in.HTTPMethod,
			Path:       in.Path,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("method_update %q in %s: %w", in.Method, in.Service, err)
		}
		return textResult(map[string]any{
			"name":        dto.Name,
			"service":     in.Service,
			"summary":     dto.Summary,
			"http_method": dto.HTTPMethod,
			"path":        dto.Path,
		})
	}
}

// ── method_delete ────────────────────────────────────────────────────────────

type methodDeleteIn struct {
	Service string `json:"service" jsonschema:"service name"`
	Method  string `json:"method"  jsonschema:"name of the method to delete"`
}

func toolMethodDelete(path string) func(context.Context, *mcp.CallToolRequest, methodDeleteIn) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in methodDeleteIn) (*mcp.CallToolResult, any, error) {
		if in.Service == "" || in.Method == "" {
			return nil, nil, fmt.Errorf("service and method are required")
		}
		if _, err := app.RemoveServiceMethod(path, in.Service, in.Method); err != nil {
			return nil, nil, fmt.Errorf("method_delete %q in %s: %w", in.Method, in.Service, err)
		}
		return textResult(map[string]any{
			"deleted": in.Method,
			"service": in.Service,
		})
	}
}

// findCapability returns the capability with matching ID (preferred) or name,
// or a zero value if not found.
func findCapability(caps []app.ServiceCapabilityDTO, id, name string) app.ServiceCapabilityDTO {
	for _, c := range caps {
		if id != "" && c.ID == id {
			return c
		}
		if name != "" && c.Name == name {
			return c
		}
	}
	return app.ServiceCapabilityDTO{}
}
