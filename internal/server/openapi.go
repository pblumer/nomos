package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

// apiEndpoint is one documented HTTP operation. The registry below is the
// single source of truth for the OpenAPI document and the human-readable /api
// page: both are generated from it. apiendpoint_coverage_test.go asserts every
// entry is actually served by the router, so the docs cannot silently drift
// away from the implementation in server.go.
type apiEndpoint struct {
	Method      string
	Path        string
	Tag         string
	Summary     string
	Description string
	Request     map[string]any // JSON request body schema, or nil
	Form        map[string]any // form-urlencoded body properties, or nil
	Response    map[string]any // success response schema, or nil for no/non-JSON body
	Code        string         // success status code; defaults to "200"
}

// apiEndpoints lists every documented endpoint of the Nomos HTTP API.
// Keep this in sync with the routes registered in NewHandler — the coverage
// test will fail if an entry here is not reachable.
var apiEndpoints = []apiEndpoint{
	// System
	{Method: "GET", Path: "/health", Tag: "System", Summary: "Health check", Description: "Returns Nomos service status and version.", Response: schemaRef("HealthResponse")},

	// Cosmos
	{Method: "GET", Path: "/api/v1/cosmos", Tag: "Cosmos", Summary: "Get Cosmos", Description: "Returns the current Cosmos metadata and aggregate counts.", Response: schemaRef("Cosmos")},
	{Method: "GET", Path: "/api/v1/index", Tag: "Cosmos", Summary: "ID index", Description: "Returns the ID→address index of artifacts (ADR-0028).", Response: schemaRef("IndexResponse")},
	{Method: "GET", Path: "/api/v1/index/{id}", Tag: "Cosmos", Summary: "Resolve ID", Description: "Resolves a stable artifact ID to its current address (ADR-0028).", Response: schemaRef("IndexEntry")},

	// Repositories
	{Method: "GET", Path: "/api/v1/repositories", Tag: "Repositories", Summary: "List repositories", Description: "Returns the git-first repositories managed by this server (ADR-0022). The local server returns its single default repository.", Response: schemaRef("RepositoriesResponse")},
	{Method: "GET", Path: "/api/v1/repositories/{repo}", Tag: "Repositories", Summary: "Get repository", Description: "Returns metadata for one repository.", Response: schemaRef("Repository")},
	{Method: "GET", Path: "/api/v1/repositories/{repo}/cosmos", Tag: "Repositories", Summary: "Get repository Cosmos", Description: "Repository-scoped Cosmos summary; equivalent to /api/v1/cosmos for the default repository.", Response: schemaRef("Cosmos")},
	{Method: "GET", Path: "/api/v1/repositories/{repo}/namespaces", Tag: "Repositories", Summary: "Get repository namespace tree", Description: "Repository-scoped namespace tree.", Response: schemaRef("NamespaceTree")},

	// Mounts
	{Method: "GET", Path: "/api/v1/mounts", Tag: "Mounts", Summary: "List mounts", Description: "Returns the implicit local server mount followed by configured remote mounts (ADR-0022).", Response: schemaRef("MountsResponse")},
	{Method: "POST", Path: "/api/v1/mounts", Tag: "Mounts", Summary: "Add mount", Description: "Registers a remote server mount by endpoint (host:7373). An optional token is the remote server's API key used for proxied writes (ADR-0023).", Request: object(map[string]any{"endpoint": stringSchema("Remote server endpoint, e.g. nomos.blumer.cloud:7373."), "label": stringSchema("Optional display label."), "token": stringSchema("Optional remote API key for write access.")}), Response: schemaRef("Mount"), Code: "201"},
	{Method: "DELETE", Path: "/api/v1/mounts/{id}", Tag: "Mounts", Summary: "Remove mount", Description: "Unmounts a remote server by id. The local mount cannot be removed.", Response: schemaRef("DeletedResponse"), Code: "204"},
	{Method: "GET", Path: "/api/v1/mounts/{id}/r/{path}", Tag: "Mounts", Summary: "Proxy to mounted server", Description: "Forwards the request to the mounted server, attaching the mount token as X-API-Key (ADR-0023). Reads are open; mutating methods require a token (403 MOUNT_NOT_AUTHENTICATED otherwise).", Response: genObj()},
	{Method: "GET", Path: "/api/v1/ping", Tag: "Mounts", Summary: "Ping", Description: "Returns this server's identity and the endpoints of its configured mounts (peers), without tokens (ADR-0025).", Response: schemaRef("Ping")},
	{Method: "GET", Path: "/api/v1/discover", Tag: "Mounts", Summary: "Discover servers", Description: "1-hop peer-gossip discovery (ADR-0025): pings the configured mounts and returns their advertised peers that are not yet mounted as candidates.", Response: schemaRef("Discovery")},

	// Decisions
	{Method: "GET", Path: "/api/v1/decisions", Tag: "Decisions", Summary: "List decisions", Description: "Returns all decisions.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/decisions", Tag: "Decisions", Summary: "Create decision", Description: "Creates a decision.", Request: genObj(), Response: genObj(), Code: "201"},
	{Method: "GET", Path: "/api/v1/decisions/{decision}", Tag: "Decisions", Summary: "Get decision", Description: "Returns one decision by id.", Response: genObj()},
	{Method: "PUT", Path: "/api/v1/decisions/{decision}", Tag: "Decisions", Summary: "Update decision", Description: "Updates a decision.", Request: genObj(), Response: genObj()},
	{Method: "DELETE", Path: "/api/v1/decisions/{decision}", Tag: "Decisions", Summary: "Delete decision", Description: "Deletes a decision.", Code: "204"},
	{Method: "POST", Path: "/api/v1/decisions/{decision}/evaluate", Tag: "Decisions", Summary: "Evaluate decision", Description: "Evaluates a decision against supplied inputs and records a trace.", Request: object(map[string]any{"inputs": genObj()}), Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/definitions", Tag: "Decisions", Summary: "Get decision definitions", Description: "Returns the parsed decision requirements graph (DRG).", Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/dmn", Tag: "Decisions", Summary: "Get decision DMN", Description: "Returns the raw DMN XML for a decision.", Response: nil},
	{Method: "PUT", Path: "/api/v1/decisions/{decision}/dmn", Tag: "Decisions", Summary: "Update decision DMN", Description: "Replaces the DMN XML of a decision.", Request: genObj(), Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/traces", Tag: "Decisions", Summary: "List decision traces", Description: "Returns recorded evaluation traces for a decision.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/decisions/{decision}/traces/verify", Tag: "Decisions", Summary: "Verify decision traces", Description: "Re-verifies recorded traces against the current decision logic.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/traces/{trace}", Tag: "Decisions", Summary: "Get decision trace", Description: "Returns one recorded evaluation trace.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/versions", Tag: "Decisions", Summary: "List decision versions", Description: "Returns immutable version snapshots of a decision.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/versions/{version}", Tag: "Decisions", Summary: "Get decision version", Description: "Returns metadata for one decision version.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/versions/{version}/dmn", Tag: "Decisions", Summary: "Get decision version DMN", Description: "Returns the DMN XML at a specific version.", Response: nil},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/versions/{version}/definitions", Tag: "Decisions", Summary: "Get decision version definitions", Description: "Returns the parsed DRG at a specific version.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/scenarios", Tag: "Decisions", Summary: "List decision scenarios", Description: "Returns saved test scenarios for a decision.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/decisions/{decision}/scenarios", Tag: "Decisions", Summary: "Create decision scenario", Description: "Creates a test scenario, optionally from an existing trace.", Request: genObj(), Response: genObj(), Code: "201"},
	{Method: "POST", Path: "/api/v1/decisions/{decision}/scenarios/from-trace/{trace}", Tag: "Decisions", Summary: "Create scenario from trace", Description: "Promotes a recorded trace into a saved scenario.", Response: genObj(), Code: "201"},
	{Method: "GET", Path: "/api/v1/decisions/{decision}/scenarios/{scenario}", Tag: "Decisions", Summary: "Get decision scenario", Description: "Returns one saved scenario.", Response: genObj()},
	{Method: "PUT", Path: "/api/v1/decisions/{decision}/scenarios/{scenario}", Tag: "Decisions", Summary: "Update decision scenario", Description: "Updates a saved scenario.", Request: genObj(), Response: genObj()},
	{Method: "DELETE", Path: "/api/v1/decisions/{decision}/scenarios/{scenario}", Tag: "Decisions", Summary: "Delete decision scenario", Description: "Removes a saved scenario.", Code: "204"},

	// Services
	{Method: "GET", Path: "/api/v1/services", Tag: "Services", Summary: "List services", Description: "Returns all services.", Response: schemaRef("ServicesResponse")},
	{Method: "POST", Path: "/api/v1/services", Tag: "Services", Summary: "Create service", Description: "Creates a service.", Form: map[string]any{"name": stringSchema("Service name."), "owner": stringSchema("Service owner."), "force": map[string]any{"type": "boolean"}}, Response: schemaRef("Service"), Code: "201"},
	{Method: "GET", Path: "/api/v1/services/refs", Tag: "Services", Summary: "List service refs", Description: "Returns canonical service references.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/services/{service}", Tag: "Services", Summary: "Get service", Description: "Returns one service by name.", Response: schemaRef("Service")},
	{Method: "PUT", Path: "/api/v1/services/{service}", Tag: "Services", Summary: "Rename service", Description: "Renames a service.", Request: object(map[string]any{"name": stringSchema("New service name.")}), Response: schemaRef("Service")},
	{Method: "DELETE", Path: "/api/v1/services/{service}", Tag: "Services", Summary: "Delete service", Description: "Deletes one service by name.", Code: "204"},
	{Method: "POST", Path: "/api/v1/services/{service}/capabilities", Tag: "Services", Summary: "Add capability", Description: "Adds a capability to a service.", Request: genObj(), Response: schemaRef("Service"), Code: "201"},
	{Method: "PUT", Path: "/api/v1/services/{service}/capabilities/{capability}", Tag: "Services", Summary: "Update capability", Description: "Updates a service capability.", Request: genObj(), Response: schemaRef("Service")},
	{Method: "DELETE", Path: "/api/v1/services/{service}/capabilities/{capability}", Tag: "Services", Summary: "Delete capability", Description: "Removes a service capability.", Response: schemaRef("Service")},
	{Method: "POST", Path: "/api/v1/services/{service}/data-objects", Tag: "Services", Summary: "Add data object", Description: "Adds a data object to a service.", Request: genObj(), Response: schemaRef("Service"), Code: "201"},
	{Method: "PUT", Path: "/api/v1/services/{service}/data-objects/{dataObject}", Tag: "Services", Summary: "Update data object", Description: "Updates a service data object.", Request: genObj(), Response: schemaRef("Service")},
	{Method: "DELETE", Path: "/api/v1/services/{service}/data-objects/{dataObject}", Tag: "Services", Summary: "Delete data object", Description: "Removes a service data object.", Response: schemaRef("Service")},
	{Method: "POST", Path: "/api/v1/services/{service}/user-interfaces", Tag: "Services", Summary: "Add user interface", Description: "Adds a user interface to a service.", Request: genObj(), Response: schemaRef("Service"), Code: "201"},
	{Method: "PUT", Path: "/api/v1/services/{service}/user-interfaces/{userInterface}", Tag: "Services", Summary: "Update user interface", Description: "Updates a service user interface.", Request: genObj(), Response: schemaRef("Service")},
	{Method: "DELETE", Path: "/api/v1/services/{service}/user-interfaces/{userInterface}", Tag: "Services", Summary: "Delete user interface", Description: "Removes a service user interface.", Response: schemaRef("Service")},
	{Method: "GET", Path: "/api/v1/services/{service}/methods", Tag: "Services", Summary: "List methods", Description: "Returns the methods defined on a service.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/services/{service}/methods", Tag: "Services", Summary: "Add method", Description: "Adds a method to a service.", Request: object(map[string]any{"method": stringSchema("Method name.")}), Response: schemaRef("Service"), Code: "201"},
	{Method: "GET", Path: "/api/v1/services/{service}/methods/{method}", Tag: "Services", Summary: "Get method", Description: "Returns one service method definition.", Response: genObj()},
	{Method: "PUT", Path: "/api/v1/services/{service}/methods/{method}", Tag: "Services", Summary: "Update method", Description: "Updates a service method definition.", Request: genObj(), Response: schemaRef("Service")},
	{Method: "DELETE", Path: "/api/v1/services/{service}/methods/{method}", Tag: "Services", Summary: "Delete method", Description: "Removes a service method.", Response: schemaRef("Service")},

	// Namespaces
	{Method: "GET", Path: "/api/v1/namespaces", Tag: "Namespaces", Summary: "Get namespace tree", Description: "Returns the cosmos namespace tree.", Response: schemaRef("NamespaceTree")},

	// Validation
	{Method: "GET", Path: "/api/v1/graph", Tag: "Validation", Summary: "Get graph", Description: "Returns Mermaid graph text by default, or JSON when format=json is used.", Response: schemaRef("Graph")},
	{Method: "GET", Path: "/api/v1/validate", Tag: "Validation", Summary: "Validate Cosmos", Description: "Runs deterministic validation for the current Cosmos.", Response: schemaRef("ValidationResult")},

	// Catalog — Blueprints
	{Method: "GET", Path: "/api/v1/blueprints", Tag: "Catalog", Summary: "List blueprints", Description: "Returns product and service blueprints.", Response: schemaRef("BlueprintsResponse")},
	{Method: "POST", Path: "/api/v1/blueprints", Tag: "Catalog", Summary: "Create blueprint", Description: "Creates a blueprint artifact.", Request: schemaRef("Blueprint"), Response: schemaRef("Blueprint"), Code: "201"},
	{Method: "GET", Path: "/api/v1/blueprints/{blueprint}", Tag: "Catalog", Summary: "Get blueprint", Description: "Returns one blueprint by id.", Response: schemaRef("Blueprint")},
	{Method: "PATCH", Path: "/api/v1/blueprints/{blueprint}", Tag: "Catalog", Summary: "Patch blueprint", Description: "Updates top-level blueprint fields.", Request: genObj(), Response: schemaRef("Blueprint")},
	{Method: "DELETE", Path: "/api/v1/blueprints/{blueprint}", Tag: "Catalog", Summary: "Delete blueprint", Description: "Deletes one blueprint by id.", Response: schemaRef("DeletedResponse")},
	{Method: "POST", Path: "/api/v1/blueprints/{blueprint}/publish", Tag: "Catalog", Summary: "Publish blueprint", Description: "Publishes a blueprint.", Response: schemaRef("Blueprint")},
	{Method: "GET", Path: "/api/v1/blueprints/{blueprint}/validate", Tag: "Catalog", Summary: "Validate blueprint", Description: "Validates one blueprint.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/blueprints/{blueprint}/requirements", Tag: "Catalog", Summary: "Add requirement", Description: "Adds a blueprint requirement and optionally links it to local blueprint attributes.", Request: object(map[string]any{"label": stringSchema("Requirement text."), "attribute_refs": arrayOf(map[string]any{"type": "string"})}), Response: schemaRef("Blueprint")},
	{Method: "PATCH", Path: "/api/v1/blueprints/{blueprint}/requirements/{requirement}", Tag: "Catalog", Summary: "Set requirement status", Description: "Sets a requirement status to open or fulfilled.", Request: object(map[string]any{"status": stringSchema("open or fulfilled.")}), Response: schemaRef("Blueprint")},
	{Method: "DELETE", Path: "/api/v1/blueprints/{blueprint}/requirements/{requirement}", Tag: "Catalog", Summary: "Delete requirement", Description: "Removes a blueprint requirement.", Response: schemaRef("Blueprint")},
	{Method: "POST", Path: "/api/v1/blueprints/{blueprint}/service-blueprints", Tag: "Catalog", Summary: "Add service blueprint", Description: "Links a service blueprint to a product blueprint.", Request: object(map[string]any{"service_id": stringSchema("Service blueprint id.")}), Response: schemaRef("Blueprint")},
	{Method: "DELETE", Path: "/api/v1/blueprints/{blueprint}/service-blueprints/{serviceBlueprint}", Tag: "Catalog", Summary: "Remove service blueprint", Description: "Unlinks a service blueprint from a product blueprint.", Response: schemaRef("Blueprint")},
	{Method: "POST", Path: "/api/v1/blueprints/{blueprint}/attributes", Tag: "Catalog", Summary: "Add attribute", Description: "Adds a typed attribute to a blueprint.", Request: object(map[string]any{"label": stringSchema("Display label."), "type": stringSchema("Attribute type."), "required": map[string]any{"type": "boolean"}, "service_ref": stringSchema("Optional service reference.")}), Response: schemaRef("Blueprint"), Code: "201"},
	{Method: "DELETE", Path: "/api/v1/blueprints/{blueprint}/attributes/{attribute}", Tag: "Catalog", Summary: "Delete attribute", Description: "Removes an attribute from a blueprint.", Response: schemaRef("Blueprint")},
	{Method: "POST", Path: "/api/v1/blueprints/{blueprint}/attributes/{attribute}/rules", Tag: "Catalog", Summary: "Add attribute rule", Description: "Adds a validation rule to a blueprint attribute.", Request: object(map[string]any{"label": stringSchema("Rule description."), "type": stringSchema("Rule type."), "value": stringSchema("Constraint value.")}), Response: schemaRef("Blueprint"), Code: "201"},
	{Method: "DELETE", Path: "/api/v1/blueprints/{blueprint}/attributes/{attribute}/rules/{rule}", Tag: "Catalog", Summary: "Delete attribute rule", Description: "Removes a validation rule from a blueprint attribute.", Response: schemaRef("Blueprint")},

	// Catalog — Products
	{Method: "GET", Path: "/api/v1/products/{product}", Tag: "Catalog", Summary: "Get product", Description: "Returns one product offering by id.", Response: schemaRef("Blueprint")},
	{Method: "GET", Path: "/api/v1/products/{product}/collaboration", Tag: "Catalog", Summary: "Get product collaboration", Description: "Returns the collaboration view for a product.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/products/{product}/processes", Tag: "Catalog", Summary: "List product processes", Description: "Returns processes attached to a product.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/products/{product}/processes", Tag: "Catalog", Summary: "Create product process", Description: "Creates a process attached to a product.", Request: genObj(), Response: genObj(), Code: "201"},
	{Method: "POST", Path: "/api/v1/products/{product}/fulfillment-services", Tag: "Catalog", Summary: "Add fulfillment service", Description: "Appends a required fulfillment service reference to a product offering.", Request: genObj(), Response: schemaRef("Blueprint")},
	{Method: "PUT", Path: "/api/v1/products/{product}/fulfillment-services/{index}", Tag: "Catalog", Summary: "Update fulfillment service", Description: "Updates an existing fulfillment service entry by index.", Request: genObj(), Response: schemaRef("Blueprint")},
	{Method: "DELETE", Path: "/api/v1/products/{product}/fulfillment-services/{index}", Tag: "Catalog", Summary: "Remove fulfillment service", Description: "Removes an existing fulfillment service entry by index.", Response: schemaRef("Blueprint")},

	// Processes
	{Method: "GET", Path: "/api/v1/processes/{process}", Tag: "Processes", Summary: "Get process", Description: "Returns one process by id.", Response: genObj()},
	{Method: "DELETE", Path: "/api/v1/processes/{process}", Tag: "Processes", Summary: "Delete process", Description: "Deletes a process.", Code: "204"},
	{Method: "GET", Path: "/api/v1/processes/{process}/bpmn", Tag: "Processes", Summary: "Get process BPMN", Description: "Returns the raw BPMN XML for a process.", Response: nil},
	{Method: "PUT", Path: "/api/v1/processes/{process}/bpmn", Tag: "Processes", Summary: "Update process BPMN", Description: "Replaces the BPMN XML of a process.", Request: genObj(), Response: genObj()},
	{Method: "GET", Path: "/api/v1/processes/{process}/tasks", Tag: "Processes", Summary: "List process tasks", Description: "Returns the tasks parsed from a process.", Response: genObj()},
	{Method: "PUT", Path: "/api/v1/processes/{process}/participant", Tag: "Processes", Summary: "Update process participant", Description: "Updates the participant of a process.", Request: genObj(), Response: genObj()},
	{Method: "GET", Path: "/api/v1/processes/{process}/triggers", Tag: "Processes", Summary: "List process triggers", Description: "Returns the triggers of a process.", Response: genObj()},
	{Method: "PUT", Path: "/api/v1/processes/{process}/triggers", Tag: "Processes", Summary: "Update process triggers", Description: "Replaces the triggers of a process.", Request: genObj(), Response: genObj()},
	{Method: "PUT", Path: "/api/v1/processes/{process}/task-mappings", Tag: "Processes", Summary: "Update task mappings", Description: "Updates the task-to-service mappings of a process.", Request: genObj(), Response: genObj()},
	{Method: "GET", Path: "/api/v1/processes/{process}/steps", Tag: "Processes", Summary: "List process steps", Description: "Returns the steps of a process.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/processes/{process}/steps", Tag: "Processes", Summary: "Add process step", Description: "Adds a step to a process.", Request: genObj(), Response: genObj(), Code: "201"},
	{Method: "PUT", Path: "/api/v1/processes/{process}/steps/{step}", Tag: "Processes", Summary: "Update process step", Description: "Updates a process step.", Request: genObj(), Response: genObj()},
	{Method: "DELETE", Path: "/api/v1/processes/{process}/steps/{step}", Tag: "Processes", Summary: "Delete process step", Description: "Removes a process step.", Response: genObj()},

	// Catalog — Instances
	{Method: "GET", Path: "/api/v1/instances", Tag: "Catalog", Summary: "List instances", Description: "Returns product and service instances.", Response: schemaRef("InstancesResponse")},
	{Method: "POST", Path: "/api/v1/instances", Tag: "Catalog", Summary: "Create instance", Description: "Creates an instance artifact.", Request: schemaRef("Instance"), Response: schemaRef("Instance"), Code: "201"},
	{Method: "GET", Path: "/api/v1/instances/{instance}", Tag: "Catalog", Summary: "Get instance", Description: "Returns one instance by id.", Response: schemaRef("Instance")},
	{Method: "PATCH", Path: "/api/v1/instances/{instance}", Tag: "Catalog", Summary: "Patch instance", Description: "Updates top-level instance fields.", Request: genObj(), Response: schemaRef("Instance")},
	{Method: "DELETE", Path: "/api/v1/instances/{instance}", Tag: "Catalog", Summary: "Delete instance", Description: "Deletes one instance by id.", Response: schemaRef("DeletedResponse")},
	{Method: "GET", Path: "/api/v1/instances/{instance}/compliance", Tag: "Catalog", Summary: "Get instance compliance", Description: "Returns compliance status, evidence, and findings for one instance.", Response: schemaRef("Compliance")},
	{Method: "POST", Path: "/api/v1/instances/{instance}/verify", Tag: "Catalog", Summary: "Verify instance", Description: "Runs verification for one instance.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/instances/{instance}/attribute-validation", Tag: "Catalog", Summary: "Validate instance attributes", Description: "Validates an instance's attribute values against its blueprint's attribute rules.", Response: schemaRef("AttributeValidation")},
	{Method: "PATCH", Path: "/api/v1/instances/{instance}/attribute-values", Tag: "Catalog", Summary: "Set attribute values", Description: "Sets or updates attribute values on an instance.", Request: map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}}, Response: schemaRef("Instance")},
	{Method: "POST", Path: "/api/v1/product-instances/{product}/service-instances", Tag: "Catalog", Summary: "Provision service instance", Description: "Provisions a service instance under a product instance.", Request: schemaRef("Instance"), Response: schemaRef("Instance"), Code: "201"},

	// Servicegraphs
	{Method: "GET", Path: "/api/v1/servicegraphs", Tag: "Servicegraphs", Summary: "List servicegraphs", Description: "Returns all servicegraphs.", Response: genObj()},
	{Method: "POST", Path: "/api/v1/servicegraphs", Tag: "Servicegraphs", Summary: "Create servicegraph", Description: "Creates a servicegraph.", Request: genObj(), Response: genObj(), Code: "201"},
	{Method: "GET", Path: "/api/v1/servicegraphs/{servicegraph}", Tag: "Servicegraphs", Summary: "Get servicegraph", Description: "Returns one servicegraph by id.", Response: genObj()},
	{Method: "DELETE", Path: "/api/v1/servicegraphs/{servicegraph}", Tag: "Servicegraphs", Summary: "Delete servicegraph", Description: "Deletes a servicegraph.", Response: schemaRef("DeletedResponse")},
	{Method: "GET", Path: "/api/v1/servicegraphs/{servicegraph}/mermaid", Tag: "Servicegraphs", Summary: "Get servicegraph Mermaid", Description: "Returns the Mermaid rendering of a servicegraph.", Response: genObj()},
	{Method: "GET", Path: "/api/v1/servicegraphs/{servicegraph}/execution", Tag: "Servicegraphs", Summary: "Get servicegraph execution order", Description: "Returns the execution order of a servicegraph.", Response: genObj()},
}

var openAPISpec = buildOpenAPISpec()

func buildOpenAPISpec() map[string]any {
	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":       "Nomos API",
			"version":     "0.1.0",
			"description": "OpenAPI documentation for the Nomos local Cosmos HTTP API. Generated from the in-code endpoint registry in openapi.go.",
		},
		"servers": []map[string]string{{"url": "/", "description": "Current Nomos server"}},
		"tags": []map[string]string{
			{"name": "System", "description": "Health and service metadata"},
			{"name": "Cosmos", "description": "Cosmos repository summary"},
			{"name": "Repositories", "description": "Server-managed git-first repositories"},
			{"name": "Mounts", "description": "Server mounts and peer discovery shown in the Cosmos Explorer"},
			{"name": "Decisions", "description": "DMN decision authoring, evaluation, traces, versions, and scenarios"},
			{"name": "Services", "description": "Service listing, creation, and detail sub-resources (capabilities, data objects, UIs, methods)"},
			{"name": "Namespaces", "description": "Cosmos namespace tree"},
			{"name": "Validation", "description": "Validation and graph outputs"},
			{"name": "Catalog", "description": "Blueprint, product, and instance catalog operations"},
			{"name": "Processes", "description": "BPMN process authoring operations"},
			{"name": "Servicegraphs", "description": "Servicegraph composition operations"},
		},
		"paths":      buildPaths(),
		"components": map[string]any{"schemas": schemas()},
	}
}

func buildPaths() map[string]any {
	paths := map[string]any{}
	for _, e := range apiEndpoints {
		item, ok := paths[e.Path].(map[string]any)
		if !ok {
			item = map[string]any{}
			paths[e.Path] = item
		}
		item[strings.ToLower(e.Method)] = buildOperation(e)
	}
	return paths
}

func buildOperation(e apiEndpoint) map[string]any {
	op := map[string]any{
		"tags":        []string{e.Tag},
		"summary":     e.Summary,
		"description": e.Description,
	}
	if params := extractPathParams(e.Path); len(params) > 0 {
		op["parameters"] = params
	}
	switch {
	case e.Request != nil:
		op["requestBody"] = jsonRequestBody(e.Request)
	case e.Form != nil:
		op["requestBody"] = formRequestBody(e.Form)
	}
	code := e.Code
	if code == "" {
		code = "200"
	}
	responses := map[string]any{}
	if e.Response != nil {
		responses[code] = response("Successful response.", e.Response)
	} else {
		responses[code] = map[string]any{"description": "Successful response."}
	}
	responses["400"] = errorResponse()
	responses["404"] = errorResponse()
	op["responses"] = responses
	return op
}

func extractPathParams(path string) []map[string]any {
	var params []map[string]any
	for _, seg := range strings.Split(path, "/") {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			name := seg[1 : len(seg)-1]
			params = append(params, pathParam(name, "Path parameter: "+name+"."))
		}
	}
	return params
}

// endpointSummary is the compact view rendered on the human-readable /api page.
type endpointSummary struct {
	Method  string
	Path    string
	Summary string
	Tag     string
	IsGet   bool
}

// endpointSummaries returns the registry sorted by path then method for display.
func endpointSummaries() []endpointSummary {
	out := make([]endpointSummary, 0, len(apiEndpoints))
	for _, e := range apiEndpoints {
		out = append(out, endpointSummary{Method: e.Method, Path: e.Path, Summary: e.Summary, Tag: e.Tag, IsGet: e.Method == http.MethodGet})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})
	return out
}

func (h *handler) openAPIJSON(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/openapi.json" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(openAPISpec)
}

func (h *handler) swaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/swagger" && r.URL.Path != "/swagger/" && r.URL.Path != "/api/docs" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Nomos API · Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({ url: '/openapi.json', dom_id: '#swagger-ui' });
    };
  </script>
</body>
</html>`))
}

func response(description string, schema map[string]any) map[string]any {
	return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
}

func errorResponse() map[string]any { return response("Error response.", schemaRef("Error")) }
func schemaRef(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}
func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func genObj() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true}
}

func pathParam(name, description string) map[string]any {
	return map[string]any{"name": name, "in": "path", "required": true, "description": description, "schema": map[string]any{"type": "string"}}
}

func jsonRequestBody(schema map[string]any) map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
}

func formRequestBody(properties map[string]any) map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/x-www-form-urlencoded": map[string]any{"schema": map[string]any{"type": "object", "properties": properties}}}}
}

func schemas() map[string]any {
	return map[string]any{
		"HealthResponse":             object(map[string]any{"service": stringSchema("Service name."), "status": stringSchema("Health status."), "version": stringSchema("Nomos version.")}),
		"Cosmos":                     object(map[string]any{"path": stringSchema("Filesystem path."), "id": stringSchema("Cosmos id."), "name": stringSchema("Cosmos name."), "version": stringSchema("Cosmos version."), "status": stringSchema("Cosmos status."), "owner": stringSchema("Owner."), "serviceCount": map[string]any{"type": "integer"}, "decisionCount": map[string]any{"type": "integer"}}),
		"Repository":                 object(map[string]any{"id": stringSchema("Repository id, unique per server."), "name": stringSchema("Display name."), "kind": stringSchema("filesystem, github, or gitbucket."), "location": stringSchema("Filesystem path or remote URL."), "default_branch": stringSchema("Default git branch."), "status": stringSchema("clean, dirty, unreachable, or unknown."), "head": stringSchema("Short HEAD commit hash.")}),
		"RepositoriesResponse":       object(map[string]any{"repositories": arrayOf(schemaRef("Repository"))}),
		"Mount":                      object(map[string]any{"id": stringSchema("Mount id."), "endpoint": stringSchema("Server endpoint host:port."), "label": stringSchema("Display label."), "local": map[string]any{"type": "boolean", "description": "True for the implicit local server mount."}, "authenticated": map[string]any{"type": "boolean", "description": "True when a write token is configured for this mount."}}),
		"MountsResponse":             object(map[string]any{"mounts": arrayOf(schemaRef("Mount"))}),
		"PingPeer":                   object(map[string]any{"endpoint": stringSchema("Peer endpoint host:port."), "label": stringSchema("Peer label.")}),
		"PingServer":                 object(map[string]any{"name": stringSchema("Server name."), "version": stringSchema("Server version."), "repositoryCount": map[string]any{"type": "integer"}}),
		"Ping":                       object(map[string]any{"server": schemaRef("PingServer"), "peers": arrayOf(schemaRef("PingPeer"))}),
		"DiscoveredServer":           object(map[string]any{"endpoint": stringSchema("Candidate endpoint host:port."), "label": stringSchema("Label."), "name": stringSchema("Server name if reachable."), "reachable": map[string]any{"type": "boolean"}, "mounted": map[string]any{"type": "boolean"}, "via": stringSchema("Endpoint of the mount that advertised this candidate.")}),
		"Discovery":                  object(map[string]any{"servers": arrayOf(schemaRef("DiscoveredServer"))}),
		"IndexEntry":                 object(map[string]any{"id": stringSchema("Stable artifact id."), "kind": stringSchema("Artifact kind."), "name": stringSchema("Name."), "address": stringSchema("Current derived address."), "domain": stringSchema("Owning domain address."), "path": stringSchema("Filesystem path.")}),
		"IndexResponse":              object(map[string]any{"entries": arrayOf(schemaRef("IndexEntry"))}),
		"Namespace":                  object(map[string]any{"canonical": stringSchema("Canonical name."), "parts": arrayOf(map[string]any{"type": "string"}), "treeParts": arrayOf(map[string]any{"type": "string"}), "treePath": stringSchema("Tree path."), "displayPath": stringSchema("Display path."), "leaf": stringSchema("Leaf name.")}),
		"Domain":                     object(map[string]any{"name": stringSchema("Domain name."), "canonical": stringSchema("Canonical domain."), "namespace": schemaRef("Namespace"), "displayName": stringSchema("Display name."), "owner": stringSchema("Owner."), "status": stringSchema("Status."), "path": stringSchema("Filesystem path."), "serviceCount": map[string]any{"type": "integer"}, "services": arrayOf(schemaRef("Service"))}),
		"MoveProductOfferingRequest": object(map[string]any{"target_domain": stringSchema("Canonical target domain."), "update_owning_domain": map[string]any{"type": "boolean", "description": "Also set owning_domain to the target domain."}}),
		"Service":                    object(map[string]any{"id": stringSchema("Service id."), "name": stringSchema("Service name."), "owner": stringSchema("Owner."), "capabilities": arrayOf(map[string]any{"type": "string"}), "supported_products": arrayOf(map[string]any{"type": "string"}), "status": stringSchema("Status."), "path": stringSchema("Filesystem path.")}),
		"ServicesResponse":           object(map[string]any{"services": arrayOf(schemaRef("Service"))}),
		"Finding":                    object(map[string]any{"code": stringSchema("Finding code."), "severity": stringSchema("Severity."), "message": stringSchema("Message."), "path": stringSchema("Optional path.")}),
		"ValidationResult":           object(map[string]any{"status": stringSchema("Validation status."), "findings": arrayOf(schemaRef("Finding"))}),
		"Graph":                      object(map[string]any{"format": stringSchema("Graph format."), "content": stringSchema("Graph content.")}),
		"NamespaceTree":              object(map[string]any{"root": map[string]any{"type": "object", "additionalProperties": true}}),
		"Blueprint":                  object(map[string]any{"id": stringSchema("Blueprint id."), "type": stringSchema("Blueprint type."), "name": stringSchema("Name."), "version": stringSchema("Version."), "status": stringSchema("Status."), "owner": stringSchema("Owner."), "offered_by": stringSchema("Offering domain."), "owning_domain": stringSchema("Owning domain."), "fulfillment": schemaRef("ProductFulfillment"), "summary": stringSchema("Summary."), "path": stringSchema("Filesystem path."), "primary_home": stringSchema("Primary domain home."), "primary_home_domain": stringSchema("Primary domain home canonical."), "required_inputs": arrayOf(map[string]any{"type": "string"}), "required_service_blueprints": arrayOf(map[string]any{"type": "string"}), "required_services": arrayOf(map[string]any{"type": "object", "additionalProperties": true}), "requirements": arrayOf(schemaRef("BlueprintRequirement")), "attributes": arrayOf(schemaRef("BlueprintAttribute"))}),
		"ProductFulfillment":         object(map[string]any{"required_services": arrayOf(schemaRef("ProductRequiredService"))}),
		"ProductRequiredService":     object(map[string]any{"service_ref": stringSchema("Canonical service reference."), "role": stringSchema("Fulfillment role."), "required": map[string]any{"type": "boolean"}, "description": stringSchema("Description."), "resolution_status": stringSchema("resolved, missing, unresolved_domain, or unresolved_service."), "fulfillment_type": stringSchema("local, cross-domain, or unresolved."), "cross_domain": map[string]any{"type": "boolean"}, "sla_ref": stringSchema("Optional SLA reference."), "ola_ref": stringSchema("Optional OLA reference."), "tree_target": stringSchema("Resolved Cosmos Explorer selection target."), "sla": schemaRef("ServiceLevel"), "ola": schemaRef("ServiceLevel"), "service_level_label": stringSchema("Compact SLA/OLA label for tree display.")}),
		"ServiceLevel":               object(map[string]any{"name": stringSchema("Service-level name."), "target": stringSchema("Target such as 4h or 99.9%."), "availability": stringSchema("Availability window."), "description": stringSchema("Description.")}),
		"BlueprintsResponse":         object(map[string]any{"blueprints": arrayOf(schemaRef("Blueprint")), "count": map[string]any{"type": "integer"}}),
		"Instance":                   object(map[string]any{"id": stringSchema("Instance id."), "type": stringSchema("Instance type."), "name": stringSchema("Name."), "blueprint_ref": stringSchema("Blueprint reference."), "blueprint_version": stringSchema("Blueprint version."), "status": stringSchema("Status."), "owner": stringSchema("Owner."), "path": stringSchema("Filesystem path."), "compliance_status": stringSchema("Compliance status."), "findings": arrayOf(map[string]any{"type": "object", "additionalProperties": true})}),
		"InstancesResponse":          object(map[string]any{"instances": arrayOf(schemaRef("Instance")), "count": map[string]any{"type": "integer"}}),
		"Compliance":                 object(map[string]any{"instance_id": stringSchema("Instance id."), "status": stringSchema("Compliance status."), "evidence": arrayOf(map[string]any{"type": "object", "additionalProperties": true}), "findings": arrayOf(map[string]any{"type": "object", "additionalProperties": true})}),
		"DeletedResponse":            object(map[string]any{"deleted": stringSchema("Deleted resource id.")}),
		"Error":                      object(map[string]any{"error": stringSchema("Stable error code."), "message": stringSchema("Human-readable error message.")}),
		"AttributeRule":              object(map[string]any{"id": stringSchema("Rule id."), "label": stringSchema("Rule label."), "type": stringSchema("Rule type: regex, max_length, min_length, starts_with, ends_with, one_of, manual, reference."), "value": stringSchema("Constraint value. For reference: blueprint id to match against.")}),
		"BlueprintRequirement":       object(map[string]any{"id": stringSchema("Requirement id."), "label": stringSchema("Requirement text."), "status": stringSchema("Requirement status: open or fulfilled."), "attribute_refs": arrayOf(map[string]any{"type": "string", "description": "Local blueprint attribute id."})}),
		"BlueprintAttribute":         object(map[string]any{"id": stringSchema("Attribute id."), "label": stringSchema("Display label."), "type": stringSchema("Attribute type: text, number, boolean, date, enum, service_ref."), "required": map[string]any{"type": "boolean"}, "service_ref": stringSchema("Optional referenced namespace service in the form <domain>/<service> for service_ref attributes."), "rules": arrayOf(schemaRef("AttributeRule"))}),
		"RuleResult":                 object(map[string]any{"rule_id": stringSchema("Rule id."), "label": stringSchema("Rule label."), "type": stringSchema("Rule type."), "status": stringSchema("Result status: pass, fail, manual."), "message": stringSchema("Status message.")}),
		"AttrValidationResult":       object(map[string]any{"attribute_id": stringSchema("Attribute id."), "label": stringSchema("Attribute label."), "value": stringSchema("Current value."), "status": stringSchema("Status: valid, invalid, missing."), "rules": arrayOf(schemaRef("RuleResult"))}),
		"AttributeValidation":        object(map[string]any{"status": stringSchema("Overall status: valid or invalid."), "attributes": arrayOf(schemaRef("AttrValidationResult"))}),
	}
}

func object(properties map[string]any) map[string]any {
	return map[string]any{"type": "object", "properties": properties}
}
func arrayOf(items map[string]any) map[string]any {
	return map[string]any{"type": "array", "items": items}
}
