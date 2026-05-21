package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/app"
)

func putJSON(h http.Handler, path, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}

func TestTypeDefAndInstanceCRUD(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	const base = "/api/v1/repositories/default/types"

	// Create a type definition with a detection file.
	if rr := postJSON(h, base, `{"id":"requirement","label":"Requirement","file":"requirement.yaml"}`); rr.Code != http.StatusCreated {
		t.Fatalf("create type status=%d body=%s", rr.Code, rr.Body.String())
	}

	// Get one + list.
	if rr := get(h, base+"/requirement"); rr.Code != 200 {
		t.Fatalf("get type status=%d body=%s", rr.Code, rr.Body.String())
	} else {
		hasAll(t, rr.Body.String(), "requirement", "Requirement", "requirement.yaml")
	}
	if rr := get(h, base); rr.Code != 200 || !strings.Contains(rr.Body.String(), "requirement") {
		t.Fatalf("list types status=%d body=%s", rr.Code, rr.Body.String())
	}

	// Update the definition.
	if rr := putJSON(h, base+"/requirement", `{"label":"Anforderung","file":"requirement.yaml"}`); rr.Code != 200 {
		t.Fatalf("put type status=%d body=%s", rr.Code, rr.Body.String())
	} else {
		hasAll(t, rr.Body.String(), "Anforderung")
	}

	// Create an instance.
	const inst = base + "/requirement/instances"
	if rr := postJSON(h, inst, `{"path":"requirements/req-001","data":{"title":"Login works","priority":"high"}}`); rr.Code != http.StatusCreated {
		t.Fatalf("create instance status=%d body=%s", rr.Code, rr.Body.String())
	}

	// Creating it again conflicts.
	if rr := postJSON(h, inst, `{"path":"requirements/req-001","data":{}}`); rr.Code != http.StatusConflict {
		t.Fatalf("duplicate instance want 409 got=%d body=%s", rr.Code, rr.Body.String())
	}

	// List instances.
	if rr := get(h, inst); rr.Code != 200 {
		t.Fatalf("list instances status=%d body=%s", rr.Code, rr.Body.String())
	} else {
		hasAll(t, rr.Body.String(), "requirements/req-001", "Login works")
	}

	// Read one + verify data round-trips.
	if rr := get(h, inst+"?path=requirements/req-001"); rr.Code != 200 {
		t.Fatalf("get instance status=%d body=%s", rr.Code, rr.Body.String())
	} else {
		var got app.TypeInstanceDTO
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode instance: %v", err)
		}
		if got.Data["title"] != "Login works" || got.Name != "req-001" {
			t.Fatalf("unexpected instance: %+v", got)
		}
	}

	// Update instance data.
	if rr := putJSON(h, inst+"?path=requirements/req-001", `{"data":{"title":"Login works","priority":"low"}}`); rr.Code != 200 {
		t.Fatalf("put instance status=%d body=%s", rr.Code, rr.Body.String())
	} else if !strings.Contains(rr.Body.String(), "low") {
		t.Fatalf("update not reflected: %s", rr.Body.String())
	}

	// Delete instance.
	if rr := deleteReq(h, inst+"?path=requirements/req-001"); rr.Code != http.StatusNoContent {
		t.Fatalf("delete instance status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := get(h, inst+"?path=requirements/req-001"); rr.Code != http.StatusNotFound {
		t.Fatalf("instance still present, status=%d", rr.Code)
	}

	// Delete the type definition.
	if rr := deleteReq(h, base+"/requirement"); rr.Code != http.StatusNoContent {
		t.Fatalf("delete type status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := get(h, base+"/requirement"); rr.Code != http.StatusNotFound {
		t.Fatalf("type still present, status=%d", rr.Code)
	}
}

func TestTypeInstancesRequireDetectionFile(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	const base = "/api/v1/repositories/default/types"

	// A type without a detection file cannot manage instances.
	if rr := postJSON(h, base, `{"id":"note","label":"Note"}`); rr.Code != http.StatusCreated {
		t.Fatalf("create type status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := get(h, base+"/note/instances"); rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for fileless type, got=%d body=%s", rr.Code, rr.Body.String())
	}
}
