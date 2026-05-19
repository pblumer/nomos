package model_test

import (
	"testing"

	"github.com/nomos/nomos/internal/model"
	"gopkg.in/yaml.v3"
)

func TestMethodDefinition_UnmarshalYAML_String(t *testing.T) {
	var m model.MethodDefinition
	if err := yaml.Unmarshal([]byte(`"createUser"`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "createUser" {
		t.Errorf("expected Name=createUser, got %q", m.Name)
	}
}

func TestMethodDefinition_UnmarshalYAML_Object(t *testing.T) {
	src := `
name: createUser
http_method: POST
path: /users
summary: Creates a user
`
	var m model.MethodDefinition
	if err := yaml.Unmarshal([]byte(src), &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "createUser" || m.HTTPMethod != "POST" || m.Path != "/users" {
		t.Errorf("unexpected method: %+v", m)
	}
}

func TestServiceCapability_UnmarshalYAML_String(t *testing.T) {
	var c model.ServiceCapability
	if err := yaml.Unmarshal([]byte(`"cap-create-user"`), &c); err != nil {
		t.Fatal(err)
	}
	if c.ID != "cap-create-user" || c.Name != "cap-create-user" {
		t.Errorf("unexpected capability: %+v", c)
	}
}

func TestServiceCapability_UnmarshalYAML_Object(t *testing.T) {
	src := `
id: cap-001
name: Create User
summary: Creates a new user account
stability: stable
`
	var c model.ServiceCapability
	if err := yaml.Unmarshal([]byte(src), &c); err != nil {
		t.Fatal(err)
	}
	if c.ID != "cap-001" || c.Name != "Create User" || c.Summary != "Creates a new user account" {
		t.Errorf("unexpected capability: %+v", c)
	}
}

func TestServiceDataObject_UnmarshalYAML_String(t *testing.T) {
	var d model.ServiceDataObject
	if err := yaml.Unmarshal([]byte(`"UserRecord"`), &d); err != nil {
		t.Fatal(err)
	}
	if d.ID != "UserRecord" || d.Name != "UserRecord" {
		t.Errorf("unexpected data object: %+v", d)
	}
}

func TestServiceDataObject_UnmarshalYAML_Object(t *testing.T) {
	src := `
id: do-user
name: User
summary: Represents a user entity
schema: UserSchema
`
	var d model.ServiceDataObject
	if err := yaml.Unmarshal([]byte(src), &d); err != nil {
		t.Fatal(err)
	}
	if d.ID != "do-user" || d.Schema != "UserSchema" {
		t.Errorf("unexpected data object: %+v", d)
	}
}

func TestServiceUserInterface_UnmarshalYAML_String(t *testing.T) {
	var u model.ServiceUserInterface
	if err := yaml.Unmarshal([]byte(`"admin-portal"`), &u); err != nil {
		t.Fatal(err)
	}
	if u.ID != "admin-portal" || u.Name != "admin-portal" {
		t.Errorf("unexpected UI: %+v", u)
	}
}

func TestServiceUserInterface_UnmarshalYAML_Object(t *testing.T) {
	src := `
id: ui-admin
name: Admin Portal
channel: web
url: https://admin.example.com
`
	var u model.ServiceUserInterface
	if err := yaml.Unmarshal([]byte(src), &u); err != nil {
		t.Fatal(err)
	}
	if u.ID != "ui-admin" || u.Channel != "web" || u.URL != "https://admin.example.com" {
		t.Errorf("unexpected UI: %+v", u)
	}
}

func TestMethodDefinition_InSlice_StringsAndObjects(t *testing.T) {
	src := `
- createUser
- name: deleteUser
  http_method: DELETE
  path: /users/{id}
`
	var methods []model.MethodDefinition
	if err := yaml.Unmarshal([]byte(src), &methods); err != nil {
		t.Fatal(err)
	}
	if len(methods) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(methods))
	}
	if methods[0].Name != "createUser" {
		t.Errorf("methods[0].Name = %q", methods[0].Name)
	}
	if methods[1].Name != "deleteUser" || methods[1].HTTPMethod != "DELETE" {
		t.Errorf("methods[1] = %+v", methods[1])
	}
}
