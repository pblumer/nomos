package app

import (
	"testing"

	"github.com/nomos/nomos/internal/model"
)

// ---------------------------------------------------------------------------
// slugify
// ---------------------------------------------------------------------------

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Create User", "create-user"},
		{"  hello  ", "hello"},
		{"Hello_World", "hello-world"},
		{"multi  space", "multi-space"},
		{"--leading", "leading"},
		{"trailing--", "trailing"},
		{"CamelCase123", "camelcase123"},
		{"already-slug", "already-slug"},
		{"", ""},
	}
	for _, c := range cases {
		got := slugify(c.in)
		if got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// removeString
// ---------------------------------------------------------------------------

func TestRemoveString(t *testing.T) {
	got := removeString([]string{"a", "b", "c", "b"}, "b")
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("unexpected result: %v", got)
	}
	// remove nonexistent
	got2 := removeString([]string{"x", "y"}, "z")
	if len(got2) != 2 {
		t.Errorf("unexpected result: %v", got2)
	}
	// empty
	got3 := removeString(nil, "x")
	if len(got3) != 0 {
		t.Errorf("unexpected result: %v", got3)
	}
}

// ---------------------------------------------------------------------------
// Capability CRUD
// ---------------------------------------------------------------------------

func TestAddServiceCapability_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	cap := model.ServiceCapability{Name: "Create User Account", Summary: "Creates a user account"}
	svc, err := AddServiceCapability(p, "user-account", cap)
	if err != nil {
		t.Fatalf("AddServiceCapability: %v", err)
	}
	if len(svc.CapabilityDefs) != 1 {
		t.Fatalf("expected 1 capability def, got %d", len(svc.CapabilityDefs))
	}
	if svc.CapabilityDefs[0].ID != "cap-create-user-account" {
		t.Errorf("unexpected ID: %q", svc.CapabilityDefs[0].ID)
	}
}

func TestAddServiceCapability_CustomID(t *testing.T) {
	p := createAppTestCosmos(t)
	cap := model.ServiceCapability{ID: "my-cap", Name: "My Cap", Summary: "My capability"}
	svc, err := AddServiceCapability(p, "user-account", cap)
	if err != nil {
		t.Fatalf("AddServiceCapability: %v", err)
	}
	if len(svc.CapabilityDefs) == 0 || svc.CapabilityDefs[0].ID != "my-cap" {
		t.Errorf("expected my-cap, got %+v", svc.CapabilityDefs)
	}
}

// TestAddServiceCapability_NameOnly verifies that a minimal capability (name only, no Summary/Stability)
// is still persisted in the Capabilities string list even if CapabilityDefs is empty.
func TestAddServiceCapability_NameOnly(t *testing.T) {
	p := createAppTestCosmos(t)
	cap := model.ServiceCapability{Name: "Simple Cap"}
	svc, err := AddServiceCapability(p, "user-account", cap)
	if err != nil {
		t.Fatalf("AddServiceCapability: %v", err)
	}
	found := false
	for _, c := range svc.Capabilities {
		if c == "Simple Cap" {
			found = true
		}
	}
	if !found {
		t.Errorf("capability not in Capabilities list: %v", svc.Capabilities)
	}
}

func TestAddServiceCapability_MissingName(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddServiceCapability(p, "user-account", model.ServiceCapability{})
	if err == nil {
		t.Error("expected error for missing capability name")
	}
}

func TestAddServiceCapability_Duplicate(t *testing.T) {
	p := createAppTestCosmos(t)
	cap := model.ServiceCapability{Name: "Dup Cap"}
	if _, err := AddServiceCapability(p, "user-account", cap); err != nil {
		t.Fatal(err)
	}
	_, err := AddServiceCapability(p, "user-account", cap)
	if err == nil {
		t.Error("expected error for duplicate capability")
	}
}

func TestUpdateServiceCapability(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceCapability(p, "user-account", model.ServiceCapability{ID: "cap-x", Name: "Cap X", Summary: "initial"}); err != nil {
		t.Fatal(err)
	}
	svc, err := UpdateServiceCapability(p, "user-account", "cap-x", model.ServiceCapability{
		Summary:       "Updated summary",
		Stability:     "stable",
		OperationRefs: []string{"createUser"},
	})
	if err != nil {
		t.Fatalf("UpdateServiceCapability: %v", err)
	}
	if len(svc.CapabilityDefs) == 0 {
		t.Fatal("expected at least 1 capability def")
	}
	if svc.CapabilityDefs[0].Summary != "Updated summary" {
		t.Errorf("unexpected Summary: %q", svc.CapabilityDefs[0].Summary)
	}
	if len(svc.CapabilityDefs[0].OperationRefs) != 1 {
		t.Errorf("expected 1 method ref, got %v", svc.CapabilityDefs[0].OperationRefs)
	}
}

func TestUpdateServiceCapability_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := UpdateServiceCapability(p, "user-account", "ghost", model.ServiceCapability{Name: "x"})
	if err == nil {
		t.Error("expected error for missing capability")
	}
}

func TestRemoveServiceCapability(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceCapability(p, "user-account", model.ServiceCapability{ID: "cap-del", Name: "Del", Summary: "to delete"}); err != nil {
		t.Fatal(err)
	}
	svc, err := RemoveServiceCapability(p, "user-account", "cap-del")
	if err != nil {
		t.Fatalf("RemoveServiceCapability: %v", err)
	}
	if len(svc.CapabilityDefs) != 0 {
		t.Errorf("expected 0 capabilities, got %d", len(svc.CapabilityDefs))
	}
}

func TestRemoveServiceCapability_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := RemoveServiceCapability(p, "user-account", "nope")
	if err == nil {
		t.Error("expected error for missing capability")
	}
}

// ---------------------------------------------------------------------------
// DataObject CRUD
// ---------------------------------------------------------------------------

func TestAddServiceDataObject_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	obj := model.ServiceDataObject{Name: "User Record"}
	svc, err := AddServiceDataObject(p, "user-account", obj)
	if err != nil {
		t.Fatalf("AddServiceDataObject: %v", err)
	}
	if len(svc.DataObjectDefs) != 1 || svc.DataObjectDefs[0].ID != "do-user-record" {
		t.Errorf("unexpected data object: %+v", svc.DataObjectDefs)
	}
}

func TestAddServiceDataObject_MissingName(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddServiceDataObject(p, "user-account", model.ServiceDataObject{})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestAddServiceDataObject_Duplicate(t *testing.T) {
	p := createAppTestCosmos(t)
	obj := model.ServiceDataObject{Name: "Dup DO"}
	if _, err := AddServiceDataObject(p, "user-account", obj); err != nil {
		t.Fatal(err)
	}
	if _, err := AddServiceDataObject(p, "user-account", obj); err == nil {
		t.Error("expected duplicate error")
	}
}

func TestUpdateServiceDataObject(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceDataObject(p, "user-account", model.ServiceDataObject{ID: "do-test", Name: "Test DO"}); err != nil {
		t.Fatal(err)
	}
	svc, err := UpdateServiceDataObject(p, "user-account", "do-test", model.ServiceDataObject{
		Summary:   "A test data object",
		Stability: "stable",
	})
	if err != nil {
		t.Fatalf("UpdateServiceDataObject: %v", err)
	}
	if len(svc.DataObjectDefs) == 0 || svc.DataObjectDefs[0].Summary != "A test data object" {
		t.Errorf("unexpected Summary: %+v", svc.DataObjectDefs)
	}
}

func TestUpdateServiceDataObject_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := UpdateServiceDataObject(p, "user-account", "ghost", model.ServiceDataObject{})
	if err == nil {
		t.Error("expected error for missing data object")
	}
}

func TestRemoveServiceDataObject(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceDataObject(p, "user-account", model.ServiceDataObject{ID: "do-rm", Name: "Remove"}); err != nil {
		t.Fatal(err)
	}
	svc, err := RemoveServiceDataObject(p, "user-account", "do-rm")
	if err != nil {
		t.Fatalf("RemoveServiceDataObject: %v", err)
	}
	if len(svc.DataObjectDefs) != 0 {
		t.Errorf("expected 0 data objects, got %d", len(svc.DataObjectDefs))
	}
}

func TestRemoveServiceDataObject_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := RemoveServiceDataObject(p, "user-account", "ghost")
	if err == nil {
		t.Error("expected error for missing data object")
	}
}

// ---------------------------------------------------------------------------
// UserInterface CRUD
// ---------------------------------------------------------------------------

func TestAddServiceUserInterface_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	ui := model.ServiceUserInterface{Name: "Admin Portal"}
	svc, err := AddServiceUserInterface(p, "user-account", ui)
	if err != nil {
		t.Fatalf("AddServiceUserInterface: %v", err)
	}
	if len(svc.UserInterfaceDefs) != 1 || svc.UserInterfaceDefs[0].ID != "ui-admin-portal" {
		t.Errorf("unexpected UI: %+v", svc.UserInterfaceDefs)
	}
}

func TestAddServiceUserInterface_MissingName(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddServiceUserInterface(p, "user-account", model.ServiceUserInterface{})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestAddServiceUserInterface_Duplicate(t *testing.T) {
	p := createAppTestCosmos(t)
	ui := model.ServiceUserInterface{Name: "Portal"}
	if _, err := AddServiceUserInterface(p, "user-account", ui); err != nil {
		t.Fatal(err)
	}
	if _, err := AddServiceUserInterface(p, "user-account", ui); err == nil {
		t.Error("expected duplicate error")
	}
}

func TestUpdateServiceUserInterface(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceUserInterface(p, "user-account", model.ServiceUserInterface{ID: "ui-test", Name: "Test UI"}); err != nil {
		t.Fatal(err)
	}
	svc, err := UpdateServiceUserInterface(p, "user-account", "ui-test", model.ServiceUserInterface{
		Channel: "web",
		URL:     "https://example.com",
	})
	if err != nil {
		t.Fatalf("UpdateServiceUserInterface: %v", err)
	}
	if len(svc.UserInterfaceDefs) == 0 || svc.UserInterfaceDefs[0].Channel != "web" {
		t.Errorf("unexpected Channel: %+v", svc.UserInterfaceDefs)
	}
}

func TestUpdateServiceUserInterface_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := UpdateServiceUserInterface(p, "user-account", "ghost", model.ServiceUserInterface{})
	if err == nil {
		t.Error("expected error for missing UI")
	}
}

func TestRemoveServiceUserInterface(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceUserInterface(p, "user-account", model.ServiceUserInterface{ID: "ui-rm", Name: "Remove UI"}); err != nil {
		t.Fatal(err)
	}
	svc, err := RemoveServiceUserInterface(p, "user-account", "ui-rm")
	if err != nil {
		t.Fatalf("RemoveServiceUserInterface: %v", err)
	}
	if len(svc.UserInterfaceDefs) != 0 {
		t.Errorf("expected 0 UIs, got %d", len(svc.UserInterfaceDefs))
	}
}

func TestRemoveServiceUserInterface_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := RemoveServiceUserInterface(p, "user-account", "ghost")
	if err == nil {
		t.Error("expected error for missing UI")
	}
}

// ---------------------------------------------------------------------------
// UpdateMethod / GetServiceOperation
// ---------------------------------------------------------------------------

func TestUpdateMethod_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceOperation(p, "user-account", "createUser"); err != nil {
		t.Fatal(err)
	}
	dto, err := UpdateOperation(p, "user-account", "createUser", model.Operation{
		Protocol: "rest",
		Summary:  "Creates a user",
		REST:     &model.RESTOperation{HTTPMethod: "POST", Path: "/users"},
	})
	if err != nil {
		t.Fatalf("UpdateOperation: %v", err)
	}
	if dto.HTTPMethod != "POST" || dto.Path != "/users" || dto.Summary != "Creates a user" {
		t.Errorf("unexpected DTO: %+v", dto)
	}
}

func TestUpdateMethod_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := UpdateOperation(p, "user-account", "ghost", model.Operation{Protocol: "rest", REST: &model.RESTOperation{HTTPMethod: "GET"}})
	if err == nil {
		t.Error("expected error for missing method")
	}
}

func TestUpdateMethod_WithHeadersAndSecurity(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceOperation(p, "user-account", "secureCreate"); err != nil {
		t.Fatal(err)
	}
	dto, err := UpdateOperation(p, "user-account", "secureCreate", model.Operation{
		Protocol: "rest",
		REST: &model.RESTOperation{
			Headers:  []model.MethodHeader{{Name: "X-Request-ID", Value: "{{requestId}}", Required: true}},
			Security: &model.MethodSecurity{Scheme: "bearer", In: "header", Name: "Authorization"},
			Payload: &model.MethodPayload{
				ContentType: "application/json",
				Fields: []model.MethodPayloadField{
					{Name: "username", Type: "string", Required: true},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateOperation: %v", err)
	}
	if len(dto.Headers) != 1 || dto.Headers[0].Name != "X-Request-ID" {
		t.Errorf("unexpected headers: %+v", dto.Headers)
	}
	if dto.Security == nil || dto.Security.Scheme != "bearer" {
		t.Errorf("unexpected security: %+v", dto.Security)
	}
	if dto.Payload == nil || len(dto.Payload.Fields) != 1 {
		t.Errorf("unexpected payload: %+v", dto.Payload)
	}
}

func TestUpdateOperationParameters(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceOperation(p, "user-account", "getUser"); err != nil {
		t.Fatal(err)
	}
	dto, err := UpdateOperationParameters(p, "user-account", "getUser", []model.MethodParameter{
		{Name: "userId", Type: "string", In: "path", Required: true},
	})
	if err != nil {
		t.Fatalf("UpdateOperationParameters: %v", err)
	}
	if len(dto.Parameters) != 1 || dto.Parameters[0].Name != "userId" {
		t.Errorf("unexpected parameters: %+v", dto.Parameters)
	}
}

func TestGetServiceOperation_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddServiceOperation(p, "user-account", "findUser"); err != nil {
		t.Fatal(err)
	}
	dto, err := GetServiceOperation(p, "user-account", "findUser")
	if err != nil {
		t.Fatalf("GetServiceOperation: %v", err)
	}
	if dto.Name != "findUser" {
		t.Errorf("unexpected name: %q", dto.Name)
	}
}

func TestGetServiceOperation_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := GetServiceOperation(p, "user-account", "ghost")
	if err == nil {
		t.Error("expected error for missing method")
	}
}
