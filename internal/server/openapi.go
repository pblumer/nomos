package server

import (
	"encoding/json"
	"net/http"
)

var openAPISpec = map[string]any{
	"openapi": "3.1.0",
	"info": map[string]any{
		"title":       "Nomos API",
		"version":     "0.1.0",
		"description": "OpenAPI documentation for the Nomos local Cosmos HTTP API.",
	},
	"servers": []map[string]string{{"url": "/", "description": "Current Nomos server"}},
	"tags": []map[string]string{
		{"name": "System", "description": "Health and service metadata"},
		{"name": "Cosmos", "description": "Cosmos repository summary"},
		{"name": "Domains", "description": "Domain and service namespace operations"},
		{"name": "Validation", "description": "Validation and graph outputs"},
		{"name": "Catalog", "description": "Blueprint and instance catalog operations"},
		{"name": "Verification", "description": "Domain verification operations"},
	},
	"paths": map[string]any{
		"/health":        pathItem("System", "Health check", "Returns Nomos service status and version.", nil, schemaRef("HealthResponse")),
		"/api/v1/cosmos": pathItem("Cosmos", "Get Cosmos", "Returns the current Cosmos metadata and aggregate counts.", nil, schemaRef("Cosmos")),
		"/api/v1/domains": map[string]any{
			"get":  operation("Domains", "List domains", "Returns all known domains.", nil, schemaRef("DomainsResponse")),
			"post": operationWithRequest("Domains", "Create domain", "Creates a domain in the local Cosmos.", nil, formRequestBody(map[string]any{"dns": stringSchema("Canonical DNS name."), "owner": stringSchema("Domain owner."), "force": map[string]any{"type": "boolean"}}), map[string]any{"201": response("Created domain.", schemaRef("Domain")), "400": errorResponse(), "409": errorResponse()}),
		},
		"/api/v1/domains/{domain}": map[string]any{
			"get":    operation("Domains", "Get domain", "Returns one domain by canonical name.", []map[string]any{pathParam("domain", "Canonical domain name.")}, schemaRef("Domain")),
			"delete": operation("Domains", "Delete domain", "Deletes one domain by canonical name.", []map[string]any{pathParam("domain", "Canonical domain name.")}, schemaRef("DeletedResponse")),
		},
		"/api/v1/domains/{domain}/products": map[string]any{
			"get":  operation("Domains", "List domain products", "Returns product offerings whose offered_by matches the domain.", []map[string]any{pathParam("domain", "Canonical domain name.")}, map[string]any{"type": "object", "additionalProperties": true}),
			"post": operationWithRequest("Domains", "Create domain product", "Creates a product offering with offered_by defaulted from the domain path.", []map[string]any{pathParam("domain", "Canonical domain name.")}, jsonRequestBody(map[string]any{"type": "object", "additionalProperties": true}), map[string]any{"201": response("Created product.", schemaRef("Blueprint")), "400": errorResponse(), "404": errorResponse(), "409": errorResponse()}),
		},
		"/api/v1/products/{product}/move": map[string]any{
			"post": operationWithRequest("Catalog", "Move product offering", "Reassigns a product offering to another domain by updating offered_by and optionally owning_domain. The Catalog Index storage path is unchanged.", []map[string]any{pathParam("product", "Product id.")}, jsonRequestBody(schemaRef("MoveProductOfferingRequest")), map[string]any{"200": response("Moved product.", schemaRef("Blueprint")), "400": errorResponse(), "404": errorResponse(), "409": errorResponse()}),
		},
		"/api/v1/products/{product}/fulfillment-services": map[string]any{
			"post": operationWithRequest("Catalog", "Add fulfillment service", "Appends a required fulfillment service reference to a product offering.", []map[string]any{pathParam("product", "Product id.")}, jsonRequestBody(map[string]any{"type": "object", "additionalProperties": true}), map[string]any{"200": response("Updated product.", schemaRef("Blueprint")), "400": errorResponse(), "404": errorResponse()}),
		},
		"/api/v1/products/{product}/fulfillment-services/{index}": map[string]any{
			"put":    operationWithRequest("Catalog", "Update fulfillment service", "Updates an existing fulfillment service entry by index.", []map[string]any{pathParam("product", "Product id."), pathParam("index", "Fulfillment entry index.")}, jsonRequestBody(map[string]any{"type": "object", "additionalProperties": true}), map[string]any{"200": response("Updated product.", schemaRef("Blueprint")), "400": errorResponse(), "404": errorResponse(), "409": errorResponse()}),
			"delete": operation("Catalog", "Remove fulfillment service", "Removes an existing fulfillment service entry by index.", []map[string]any{pathParam("product", "Product id."), pathParam("index", "Fulfillment entry index.")}, schemaRef("Blueprint")),
		},
		"/api/v1/services/refs": pathItem("Domains", "List service refs", "Returns canonical service references grouped by domain fields.", nil, map[string]any{"type": "object", "additionalProperties": true}),
		"/api/v1/domains/{domain}/services": map[string]any{
			"get":  operation("Domains", "List domain services", "Returns services below a domain.", []map[string]any{pathParam("domain", "Canonical domain name.")}, schemaRef("ServicesResponse")),
			"post": operationWithRequest("Domains", "Create service", "Creates a service below a domain.", []map[string]any{pathParam("domain", "Canonical domain name.")}, formRequestBody(map[string]any{"name": stringSchema("Service name."), "owner": stringSchema("Service owner."), "force": map[string]any{"type": "boolean"}}), map[string]any{"201": response("Created service.", schemaRef("Service")), "400": errorResponse(), "409": errorResponse()}),
		},
		"/api/v1/domains/{domain}/services/{service}": map[string]any{
			"get":    operation("Domains", "Get service", "Returns one service below a domain.", []map[string]any{pathParam("domain", "Canonical domain name."), pathParam("service", "Service name.")}, schemaRef("Service")),
			"delete": operation("Domains", "Delete service", "Deletes one service below a domain.", []map[string]any{pathParam("domain", "Canonical domain name."), pathParam("service", "Service name.")}, schemaRef("DeletedResponse")),
		},
		"/api/v1/services/{domain}/{service}": pathItem("Domains", "Get service (legacy)", "Legacy route for reading one service.", []map[string]any{pathParam("domain", "Canonical domain name."), pathParam("service", "Service name.")}, schemaRef("Service")),
		"/api/v1/namespaces":                  pathItem("Domains", "Get namespace tree", "Returns the domain/service namespace tree.", nil, schemaRef("NamespaceTree")),
		"/api/v1/graph":                       map[string]any{"get": map[string]any{"tags": []string{"Validation"}, "summary": "Get graph", "description": "Returns Mermaid graph text by default, or JSON when format=json is used.", "parameters": []map[string]any{{"name": "format", "in": "query", "required": false, "schema": map[string]any{"type": "string", "enum": []string{"json"}}}}, "responses": map[string]any{"200": map[string]any{"description": "Mermaid text or graph JSON.", "content": map[string]any{"text/plain": map[string]any{"schema": map[string]any{"type": "string"}}, "application/json": map[string]any{"schema": schemaRef("Graph")}}}, "500": errorResponse()}}},
		"/api/v1/validate":                    pathItem("Validation", "Validate Cosmos", "Runs deterministic validation for the current Cosmos.", nil, schemaRef("ValidationResult")),
		"/api/v1/blueprints": map[string]any{
			"get":  operation("Catalog", "List blueprints", "Returns product and service blueprints.", nil, schemaRef("BlueprintsResponse")),
			"post": operationWithRequest("Catalog", "Create blueprint", "Creates a blueprint artifact.", nil, jsonRequestBody(schemaRef("Blueprint")), map[string]any{"201": response("Created blueprint.", schemaRef("Blueprint")), "400": errorResponse(), "409": errorResponse()}),
		},
		"/api/v1/blueprints/{blueprint}": map[string]any{
			"get":    operation("Catalog", "Get blueprint", "Returns one blueprint by id.", []map[string]any{pathParam("blueprint", "Blueprint id.")}, schemaRef("Blueprint")),
			"delete": operation("Catalog", "Delete blueprint", "Deletes one blueprint by id.", []map[string]any{pathParam("blueprint", "Blueprint id.")}, schemaRef("DeletedResponse")),
		},
		"/api/v1/blueprints/{blueprint}/requirements": map[string]any{
			"post": operationWithRequest("Catalog", "Add requirement", "Adds a blueprint requirement and optionally links it to local blueprint attributes.", []map[string]any{pathParam("blueprint", "Blueprint id.")}, jsonRequestBody(object(map[string]any{"label": stringSchema("Requirement text."), "attribute_refs": arrayOf(map[string]any{"type": "string", "description": "Local blueprint attribute id."})})), map[string]any{"200": response("Updated blueprint.", schemaRef("Blueprint")), "400": errorResponse()}),
		},
		"/api/v1/blueprints/{blueprint}/requirements/{requirement}": map[string]any{
			"patch":  operationWithRequest("Catalog", "Set requirement status", "Sets a requirement status to open or fulfilled.", []map[string]any{pathParam("blueprint", "Blueprint id."), pathParam("requirement", "Requirement id.")}, jsonRequestBody(object(map[string]any{"status": stringSchema("open or fulfilled.")})), map[string]any{"200": response("Updated blueprint.", schemaRef("Blueprint")), "400": errorResponse()}),
			"delete": operation("Catalog", "Delete requirement", "Removes a blueprint requirement.", []map[string]any{pathParam("blueprint", "Blueprint id."), pathParam("requirement", "Requirement id.")}, schemaRef("Blueprint")),
		},
		"/api/v1/blueprints/{blueprint}/attributes": map[string]any{
			"post": operationWithRequest("Catalog", "Add attribute", "Adds a typed attribute to a blueprint.", []map[string]any{pathParam("blueprint", "Blueprint id.")}, jsonRequestBody(object(map[string]any{"label": stringSchema("Display label."), "type": stringSchema("Attribute type: text, number, boolean, date, enum, service_ref."), "required": map[string]any{"type": "boolean"}, "service_ref": stringSchema("Optional referenced namespace service in the form <domain>/<service> for service_ref attributes.")})), map[string]any{"201": response("Updated blueprint.", schemaRef("Blueprint")), "400": errorResponse()}),
		},
		"/api/v1/blueprints/{blueprint}/attributes/{attribute}": map[string]any{
			"delete": operation("Catalog", "Delete attribute", "Removes an attribute from a blueprint.", []map[string]any{pathParam("blueprint", "Blueprint id."), pathParam("attribute", "Attribute id.")}, schemaRef("Blueprint")),
		},
		"/api/v1/blueprints/{blueprint}/attributes/{attribute}/rules": map[string]any{
			"post": operationWithRequest("Catalog", "Add attribute rule", "Adds a validation rule to a blueprint attribute.", []map[string]any{pathParam("blueprint", "Blueprint id."), pathParam("attribute", "Attribute id.")}, jsonRequestBody(object(map[string]any{"label": stringSchema("Rule description."), "type": stringSchema("Rule type: regex, max_length, min_length, starts_with, ends_with, one_of, manual, reference."), "value": stringSchema("Constraint value. For reference: blueprint id to match against.")})), map[string]any{"201": response("Updated blueprint.", schemaRef("Blueprint")), "400": errorResponse()}),
		},
		"/api/v1/blueprints/{blueprint}/attributes/{attribute}/rules/{rule}": map[string]any{
			"delete": operation("Catalog", "Delete attribute rule", "Removes a validation rule from a blueprint attribute.", []map[string]any{pathParam("blueprint", "Blueprint id."), pathParam("attribute", "Attribute id."), pathParam("rule", "Rule id.")}, schemaRef("Blueprint")),
		},
		"/api/v1/instances":                                 pathItem("Catalog", "List instances", "Returns product and service instances.", nil, schemaRef("InstancesResponse")),
		"/api/v1/instances/{instance}":                      pathItem("Catalog", "Get instance", "Returns one instance by id.", []map[string]any{pathParam("instance", "Instance id.")}, schemaRef("Instance")),
		"/api/v1/instances/{instance}/compliance":           pathItem("Catalog", "Get instance compliance", "Returns compliance status, evidence, and findings for one instance.", []map[string]any{pathParam("instance", "Instance id.")}, schemaRef("Compliance")),
		"/api/v1/instances/{instance}/attribute-validation": pathItem("Catalog", "Validate instance attributes", "Validates an instance's attribute values against its blueprint's attribute rules.", []map[string]any{pathParam("instance", "Instance id.")}, schemaRef("AttributeValidation")),
		"/api/v1/instances/{instance}/attribute-values":     map[string]any{"patch": operationWithRequest("Catalog", "Set attribute values", "Sets or updates attribute values on an instance.", []map[string]any{pathParam("instance", "Instance id.")}, jsonRequestBody(map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}}), map[string]any{"200": response("Updated instance.", schemaRef("Instance")), "400": errorResponse()})},
		"/api/v1/verify/domain/{domain}":                    map[string]any{"post": operation("Verification", "Verify domain", "Runs domain verification for one domain.", []map[string]any{pathParam("domain", "Canonical domain name.")}, map[string]any{"type": "object", "additionalProperties": true})},
	},
	"components": map[string]any{"schemas": schemas()},
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

func pathItem(tag, summary, description string, params []map[string]any, responseSchema map[string]any) map[string]any {
	return map[string]any{"get": operation(tag, summary, description, params, responseSchema)}
}

func operation(tag, summary, description string, params []map[string]any, responseSchema map[string]any) map[string]any {
	op := map[string]any{"tags": []string{tag}, "summary": summary, "description": description, "responses": map[string]any{"200": response("Successful response.", responseSchema), "404": errorResponse(), "500": errorResponse()}}
	if len(params) > 0 {
		op["parameters"] = params
	}
	return op
}

func operationWithRequest(tag, summary, description string, params []map[string]any, requestBody map[string]any, responses map[string]any) map[string]any {
	op := map[string]any{"tags": []string{tag}, "summary": summary, "description": description, "requestBody": requestBody, "responses": responses}
	if len(params) > 0 {
		op["parameters"] = params
	}
	return op
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
		"Cosmos":                     object(map[string]any{"path": stringSchema("Filesystem path."), "id": stringSchema("Cosmos id."), "name": stringSchema("Cosmos name."), "version": stringSchema("Cosmos version."), "status": stringSchema("Cosmos status."), "owner": stringSchema("Owner."), "domainCount": map[string]any{"type": "integer"}, "serviceCount": map[string]any{"type": "integer"}}),
		"Namespace":                  object(map[string]any{"canonical": stringSchema("Canonical name."), "parts": arrayOf(map[string]any{"type": "string"}), "treeParts": arrayOf(map[string]any{"type": "string"}), "treePath": stringSchema("Tree path."), "displayPath": stringSchema("Display path."), "leaf": stringSchema("Leaf name.")}),
		"Domain":                     object(map[string]any{"name": stringSchema("Domain name."), "canonical": stringSchema("Canonical domain."), "namespace": schemaRef("Namespace"), "displayName": stringSchema("Display name."), "owner": stringSchema("Owner."), "status": stringSchema("Status."), "path": stringSchema("Filesystem path."), "serviceCount": map[string]any{"type": "integer"}, "services": arrayOf(schemaRef("Service"))}),
		"MoveProductOfferingRequest": object(map[string]any{"target_domain": stringSchema("Canonical target domain."), "update_owning_domain": map[string]any{"type": "boolean", "description": "Also set owning_domain to the target domain."}}),
		"Service":                    object(map[string]any{"name": stringSchema("Service name."), "domain": stringSchema("Domain."), "owner": stringSchema("Owner."), "owned_by": stringSchema("Owning domain."), "operated_by": arrayOf(map[string]any{"type": "string"}), "capabilities": arrayOf(map[string]any{"type": "string"}), "supported_products": arrayOf(map[string]any{"type": "string"}), "status": stringSchema("Status."), "path": stringSchema("Filesystem path.")}),
		"DomainsResponse":            object(map[string]any{"domains": arrayOf(schemaRef("Domain"))}),
		"ServicesResponse":           object(map[string]any{"domain": stringSchema("Domain."), "services": arrayOf(schemaRef("Service"))}),
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
