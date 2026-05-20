package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fillPathParams replaces every {param} segment with a sentinel value so the
// request reaches the routing layer. We only care that the route is registered
// and the method is handled — not that the sentinel resource exists.
func fillPathParams(path string) string {
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
			segs[i] = "x"
		}
	}
	return strings.Join(segs, "/")
}

// TestOpenAPICoverage asserts that every endpoint documented in the registry is
// actually served by NewHandler. If a route is renamed, removed, or its method
// changes without the registry being updated, this test fails — which is what
// keeps the OpenAPI document from silently drifting away from server.go.
func TestOpenAPICoverage(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	for _, e := range apiEndpoints {
		e := e
		t.Run(e.Method+" "+e.Path, func(t *testing.T) {
			reqPath := fillPathParams(e.Path)
			var body *strings.Reader
			if e.Method != http.MethodGet {
				body = strings.NewReader("{}")
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(e.Method, reqPath, body)
			if e.Method != http.MethodGet {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code == http.StatusMethodNotAllowed {
				t.Fatalf("%s %s: method not handled (405) — registry is out of sync with server routing", e.Method, e.Path)
			}
			if ct := rr.Header().Get("Content-Type"); strings.Contains(ct, "text/html") {
				t.Fatalf("%s %s: request fell through to the web/HTML handler (Content-Type %q) — route is not registered as an API endpoint", e.Method, e.Path, ct)
			}
			if strings.TrimSpace(rr.Body.String()) == "404 page not found" {
				t.Fatalf("%s %s: route not registered (default 404) — documented endpoint does not exist in server.go", e.Method, e.Path)
			}
		})
	}
}

// TestOpenAPISpecServed verifies the generated document is exposed and well-formed.
func TestOpenAPISpecServed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /openapi.json: %d", rr.Code)
	}
	var doc struct {
		OpenAPI string                    `json:"openapi"`
		Paths   map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &doc); err != nil {
		t.Fatalf("openapi.json is not valid JSON: %v", err)
	}
	if doc.OpenAPI == "" {
		t.Error("missing openapi version field")
	}
	// Every registry entry must appear in the generated paths under its method.
	for _, e := range apiEndpoints {
		ops, ok := doc.Paths[e.Path]
		if !ok {
			t.Errorf("path %s missing from generated spec", e.Path)
			continue
		}
		if _, ok := ops[strings.ToLower(e.Method)]; !ok {
			t.Errorf("operation %s %s missing from generated spec", e.Method, e.Path)
		}
	}
}
