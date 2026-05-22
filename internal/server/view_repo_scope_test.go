package server

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestApiViewResolvesWithinAttachedRepo covers the repo-aware path resolution:
// a freshly created filesystem repository seeds decision_*.frm under its own
// working directory, so the view endpoint must find them only when told which
// repository owns the (repo-relative) path.
func TestApiViewResolvesWithinAttachedRepo(t *testing.T) {
	h := NewHandler(t.TempDir())

	rr := postJSON(h, "/api/v1/repositories", `{"name":"myFirst"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create repository: code %d, body %s", rr.Code, rr.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil || created.ID == "" {
		t.Fatalf("decode created repo: %v (%s)", err, rr.Body.String())
	}

	// With the owning repo, the seeded form resolves.
	scoped := get(h, "/api/v1/view?path=.nomos/views/decision_new.frm&repo="+created.ID)
	if scoped.Code != http.StatusOK {
		t.Fatalf("scoped view: code %d, body %s", scoped.Code, scoped.Body.String())
	}

	// Without the repo, the same repo-relative path is looked up in the default
	// workspace, where it does not exist.
	unscoped := get(h, "/api/v1/view?path=.nomos/views/decision_new.frm")
	if unscoped.Code != http.StatusNotFound {
		t.Fatalf("unscoped view: code %d, want 404", unscoped.Code)
	}

	// All four seeded variants resolve within the repo.
	for _, v := range []string{"new", "edit", "list", "short"} {
		rr := get(h, "/api/v1/view?path=.nomos/views/decision_"+v+".frm&repo="+created.ID)
		if rr.Code != http.StatusOK {
			t.Errorf("variant %s: code %d, body %s", v, rr.Code, rr.Body.String())
		}
	}
}
