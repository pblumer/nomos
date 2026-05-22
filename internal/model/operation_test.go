package model

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOperationRoundTrip(t *testing.T) {
	in := Service{
		ID:   "SRV_1",
		Type: "service",
		Name: "billing",
		Operations: []Operation{
			{
				Name:         "createInvoice",
				Protocol:     "rest",
				InputObject:  "invoice",
				OutputObject: "invoice_result",
				REST:         &RESTOperation{HTTPMethod: "POST", Path: "/invoices", BaseURL: "https://api.example.com"},
			},
			{Name: "lookupTool", Protocol: "mcp", MCP: &MCPOperation{Transport: "http", ServerURL: "https://mcp.example.com", Tool: "search"}},
			{Name: "calc", Protocol: "grpc", GRPC: &GRPCOperation{Target: "host:50051", Service: "Calc", Method: "Add"}},
		},
	}
	data, err := yaml.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Service
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Operations) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(out.Operations))
	}
	if out.Operations[0].REST == nil || out.Operations[0].REST.HTTPMethod != "POST" {
		t.Errorf("rest variant lost: %+v", out.Operations[0])
	}
	if out.Operations[0].InputObject != "invoice" {
		t.Errorf("input_object lost: %q", out.Operations[0].InputObject)
	}
	if out.Operations[1].MCP == nil || out.Operations[1].MCP.Tool != "search" {
		t.Errorf("mcp variant lost: %+v", out.Operations[1])
	}
	if out.Operations[2].GRPC == nil || out.Operations[2].GRPC.Method != "Add" {
		t.Errorf("grpc variant lost: %+v", out.Operations[2])
	}
}

func TestMethodToOperation(t *testing.T) {
	m := MethodDefinition{Name: "getThing", Summary: "Reads a thing", HTTPMethod: "GET", Path: "/things/{id}"}
	op := m.ToOperation()
	if op.Protocol != "rest" || op.Name != "getThing" || op.Summary != "Reads a thing" {
		t.Fatalf("unexpected operation: %+v", op)
	}
	if op.REST == nil || op.REST.HTTPMethod != "GET" || op.REST.Path != "/things/{id}" {
		t.Errorf("rest fields not lifted: %+v", op.REST)
	}
}
