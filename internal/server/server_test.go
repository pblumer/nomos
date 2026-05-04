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
func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	return rr
}
func hasAll(t *testing.T, b string, m ...string) {
	t.Helper()
	for _, s := range m {
		if !strings.Contains(b, s) {
			t.Fatalf("missing %q", s)
		}
	}
}

func TestWebShellPagesAndAPI(t *testing.T) { /* same as before */
	h := NewHandler(createTestCosmos(t))
	if rr := get(h, "/"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Nomos", "Governance Platform", "Governance Dashboard", "app-sidebar", "app-topbar", "metric-card", "Cosmos Overview")
	}
	if rr := get(h, "/domains"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Domains Explorer", "Browse domains and services like a repository tree", "identity.blumer.cloud", "user-account", "explorer-layout", "explorer-tree", "details-panel", "Domain count", "Service count")
	}
	if rr := get(h, "/domains?selected=domain:identity.blumer.cloud"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Domain Details", "identity.blumer.cloud", "Service count", "user-account")
	}
	if rr := get(h, "/domains?selected=service:identity.blumer.cloud/user-account"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Service Details", "user-account", "identity.blumer.cloud")
	}
	if rr := get(h, "/domains?selected=domain:does-not-exist.example"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Cosmos Summary", "Local Cosmos")
	}
	if rr := get(h, "/domains/identity.blumer.cloud"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "identity.blumer.cloud")
	}
	if rr := get(h, "/services/identity.blumer.cloud/user-account"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "user-account", "identity.blumer.cloud", "Service Metadata")
	}
	if rr := get(h, "/graph"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Cosmos Graph", "graph TD", "code-block")
	}
	if rr := get(h, "/validate"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else if !(strings.Contains(rr.Body.String(), "No validation findings") || strings.Contains(rr.Body.String(), "Findings")) {
		t.Fatal("missing validation state")
	}
	if rr := get(h, "/validate"); !strings.Contains(rr.Body.String(), "badge") && !strings.Contains(rr.Body.String(), "alert") {
		t.Fatal("missing badge/alert")
	}
	if rr := get(h, "/static/app.css"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "--color-primary", ".app-sidebar", ".metric-card", ".explorer-layout", ".explorer-tree", ".details-panel")
		if !strings.Contains(rr.Header().Get("Content-Type"), "text/css") {
			t.Fatal(rr.Header().Get("Content-Type"))
		}
	}
	if rr := get(h, "/health"); rr.Code != 200 {
		t.Fatal(rr.Code)
	}
	var m map[string]any
	_ = json.Unmarshal(get(h, "/api/v1/cosmos").Body.Bytes(), &m)
	if m["name"] != "Local Cosmos" {
		t.Fatal(m)
	}
	if get(h, "/api/v1/domains").Code != 200 || get(h, "/api/v1/graph").Code != 200 || get(h, "/api/v1/validate").Code != 200 {
		t.Fatal("api failed")
	}
	if get(h, "/does-not-exist").Code != 404 {
		t.Fatal()
	}
}

func TestMissingCosmosStyledError(t *testing.T) {
	h := NewHandler(t.TempDir())
	rr := get(h, "/")
	if rr.Code == 200 {
		t.Fatal("expected non-200")
	}
	hasAll(t, rr.Body.String(), "Nomos", "Error", "Back to dashboard")
	if !strings.Contains(get(h, "/api/v1/cosmos").Body.String(), "cosmos.yaml") {
		t.Fatal()
	}
}
