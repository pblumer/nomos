package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func TestServiceUserInterfaceCarriesEngineSchemaBinding(t *testing.T) {
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/user-account"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: c\nname: C\nversion: 0.1.0\nstatus: draft\nowner: t\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/domain.yaml"), []byte("name: identity.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/user-account/service.yaml"), []byte(`name: user-account
owned_by: identity.blumer.cloud
user_interfaces:
  - id: UI-CAPTURE
    name: Benutzerkonto erfassen
    engine: form-js
    engine_version: "1"
    schema:
      type: default
      components:
        - type: textfield
          key: given_name
          label: Vorname
    binding:
      data_object_ref: identity.blumer.cloud/user-account/PersonData
      submit: identity.blumer.cloud/user-account#createAccount
`), 0o644))

	svc, err := GetService(p, "identity.blumer.cloud", "user-account")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.UserInterfaceDefs) != 1 {
		t.Fatalf("expected 1 user interface, got %d", len(svc.UserInterfaceDefs))
	}
	ui := svc.UserInterfaceDefs[0]
	if ui.Engine != "form-js" || ui.EngineVersion != "1" {
		t.Fatalf("engine not carried: %+v", ui)
	}
	if ui.Schema == nil || ui.Schema["type"] != "default" {
		t.Fatalf("engine-native schema not carried: %+v", ui.Schema)
	}
	if ui.Binding == nil || ui.Binding.Submit != "identity.blumer.cloud/user-account#createAccount" {
		t.Fatalf("binding not carried: %+v", ui.Binding)
	}
}
