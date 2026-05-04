package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestCosmos(t *testing.T) string {
	p := t.TempDir()
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	must(os.MkdirAll(filepath.Join(p, "domains/identity.blumer.cloud/services/user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "domains/platform.blumer.cloud/services/rule-validation-api"), 0o755))
	must(os.WriteFile(filepath.Join(p, "cosmos.yaml"), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: unknown\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/identity.blumer.cloud/domain.yaml"), []byte("name: identity.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/platform.blumer.cloud/domain.yaml"), []byte("name: platform.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/identity.blumer.cloud/services/user-account/service.yaml"), []byte("name: user-account\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/platform.blumer.cloud/services/rule-validation-api/service.yaml"), []byte("name: rule-validation-api\n"), 0o644))
	return p
}
func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	return rr
}
func TestAll(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	if get(t, h, "/health").Code != 200 {
		t.Fatal()
	}
	if !strings.Contains(get(t, h, "/").Body.String(), "Local Cosmos") {
		t.Fatal()
	}
	if !strings.Contains(get(t, h, "/domains").Body.String(), "identity.blumer.cloud") {
		t.Fatal()
	}
	if get(t, h, "/domains/identity.blumer.cloud").Code != 200 {
		t.Fatal()
	}
	if get(t, h, "/services/identity.blumer.cloud/user-account").Code != 200 {
		t.Fatal()
	}
	if !strings.Contains(get(t, h, "/graph").Body.String(), "graph TD") {
		t.Fatal()
	}
	if get(t, h, "/validate").Code != 200 {
		t.Fatal()
	}
	var m map[string]any
	_ = json.Unmarshal(get(t, h, "/api/v1/cosmos").Body.Bytes(), &m)
	if m["name"] != "Local Cosmos" {
		t.Fatal(m)
	}
	if !strings.HasPrefix(get(t, h, "/api/v1/graph").Body.String(), "graph TD") {
		t.Fatal()
	}
	if get(t, h, "/does-not-exist").Code != 404 {
		t.Fatal()
	}
}
func TestMissingCosmos(t *testing.T) {
	h := NewHandler(t.TempDir())
	if get(t, h, "/").Code != 500 {
		t.Fatal()
	}
	if !strings.Contains(get(t, h, "/api/v1/cosmos").Body.String(), "cosmos.yaml") {
		t.Fatal()
	}
}
