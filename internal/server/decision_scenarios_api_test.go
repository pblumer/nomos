package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

const scenarioTestDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="DEC-001" name="Eligibility">
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="i1" label="Category">
        <inputExpression typeRef="string"><text>category</text></inputExpression>
      </input>
      <output id="o1" name="eligible" typeRef="boolean"/>
      <rule id="r1">
        <inputEntry><text>"premium"</text></inputEntry>
        <outputEntry><text>true</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>false</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

// scenarioTestCosmos sets up a minimal cosmos with one decision that has a
// DMN file, so we can exercise the full /scenarios route surface.
func scenarioTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	dom := filepath.Join(storage.DomainsDir(p), "governance.blumer.com")
	must(os.MkdirAll(filepath.Join(dom, "decisions", "DEC-001"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p),
		[]byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Team\n"), 0o644))
	must(os.WriteFile(filepath.Join(dom, "domain.yaml"),
		[]byte("name: governance.blumer.com\nowner: Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(dom, "decisions", "DEC-001", "decision.yaml"),
		[]byte("id: DEC-001\ntype: decision\nname: Eligibility\nversion: 0.1.0\nstatus: draft\nowner: Team\ndmn_file: decision.dmn\n"), 0o644))
	must(os.WriteFile(filepath.Join(dom, "decisions", "DEC-001", "decision.dmn"),
		[]byte(scenarioTestDMN), 0o644))
	return p
}

func TestDecisionScenarios_FreeFormCreateAndList(t *testing.T) {
	p := scenarioTestCosmos(t)
	h := NewHandler(p)
	base := "/api/v1/domains/governance.blumer.com/decisions/DEC-001/scenarios"

	// Empty list initially.
	if rr := get(h, base); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"count":0`) {
		t.Fatalf("initial list: code=%d body=%s", rr.Code, rr.Body.String())
	}

	// Free-form create.
	create := postJSONMethod(h, http.MethodPost, base,
		`{"name":"Premium happy path","inputs":{"category":"premium"},"expected_outputs":{"eligible":true}}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create: code=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	sid, _ := created["id"].(string)
	if !strings.HasPrefix(sid, "VSC_") {
		t.Errorf("expected VSC_ prefix, got %q", sid)
	}

	// List now has one.
	rr := get(h, base)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"count":1`) {
		t.Fatalf("list after create: code=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Premium happy path") {
		t.Errorf("scenario name missing from list: %s", rr.Body.String())
	}

	// Get by ID.
	rr = get(h, base+"/"+sid)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), sid) {
		t.Fatalf("get by id: code=%d body=%s", rr.Code, rr.Body.String())
	}

	// Update name and clear expected_outputs.
	upd := postJSONMethod(h, http.MethodPut, base+"/"+sid, `{"name":"renamed","clear_expected":true}`)
	if upd.Code != http.StatusOK || !strings.Contains(upd.Body.String(), "renamed") {
		t.Fatalf("update: code=%d body=%s", upd.Code, upd.Body.String())
	}
	if strings.Contains(upd.Body.String(), "expected_outputs") {
		t.Errorf("expected_outputs should be cleared, body=%s", upd.Body.String())
	}

	// Delete.
	del := postJSONMethod(h, http.MethodDelete, base+"/"+sid, "")
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete: code=%d body=%s", del.Code, del.Body.String())
	}
	// 404 on second delete.
	if del2 := postJSONMethod(h, http.MethodDelete, base+"/"+sid, ""); del2.Code != http.StatusNotFound {
		t.Fatalf("delete again: code=%d body=%s", del2.Code, del2.Body.String())
	}
}

func TestDecisionScenarios_FromTraceShortcut(t *testing.T) {
	p := scenarioTestCosmos(t)
	h := NewHandler(p)

	// First produce a trace via /evaluate so the from-trace shortcut has
	// something to consume.
	ev := postJSONMethod(h, http.MethodPost,
		"/api/v1/domains/governance.blumer.com/decisions/DEC-001/evaluate",
		`{"inputs":{"category":"premium"}}`)
	if ev.Code != http.StatusOK {
		t.Fatalf("evaluate: code=%d body=%s", ev.Code, ev.Body.String())
	}
	var evResp struct {
		Trace struct {
			TraceID string `json:"trace_id"`
		} `json:"trace"`
	}
	if err := json.Unmarshal(ev.Body.Bytes(), &evResp); err != nil || evResp.Trace.TraceID == "" {
		t.Fatalf("decode evaluate response: %v body=%s", err, ev.Body.String())
	}

	// Promote it.
	base := "/api/v1/domains/governance.blumer.com/decisions/DEC-001/scenarios"
	rr := postJSONMethod(h, http.MethodPost,
		base+"/from-trace/"+evResp.Trace.TraceID,
		`{"name":"Premium baseline","description":"first golden case"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("from-trace: code=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	// Should have inherited inputs + outputs from the trace, plus reference it.
	if !strings.Contains(body, `"category":"premium"`) {
		t.Errorf("inputs not seeded from trace: %s", body)
	}
	if !strings.Contains(body, `"eligible":true`) {
		t.Errorf("expected_outputs not seeded from trace outputs: %s", body)
	}
	if !strings.Contains(body, evResp.Trace.TraceID) {
		t.Errorf("source_trace_id not set: %s", body)
	}

	// Missing name is a 400.
	if rr := postJSONMethod(h, http.MethodPost,
		base+"/from-trace/"+evResp.Trace.TraceID, `{}`); rr.Code == http.StatusCreated {
		t.Errorf("missing name should not succeed: code=%d body=%s", rr.Code, rr.Body.String())
	}
	// Unknown trace ID is a 404.
	if rr := postJSONMethod(h, http.MethodPost,
		base+"/from-trace/sha256:doesnotexist", `{"name":"x"}`); rr.Code != http.StatusNotFound {
		t.Errorf("unknown trace should 404: code=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDecisionScenarios_DispatchFromTraceViaBody(t *testing.T) {
	// The POST .../scenarios endpoint also accepts from_trace_id in the body
	// so the UI can use a single create endpoint for both paths.
	p := scenarioTestCosmos(t)
	h := NewHandler(p)
	ev := postJSONMethod(h, http.MethodPost,
		"/api/v1/domains/governance.blumer.com/decisions/DEC-001/evaluate",
		`{"inputs":{"category":"premium"}}`)
	if ev.Code != http.StatusOK {
		t.Fatalf("evaluate: code=%d body=%s", ev.Code, ev.Body.String())
	}
	var evResp struct {
		Trace struct {
			TraceID string `json:"trace_id"`
		} `json:"trace"`
	}
	_ = json.Unmarshal(ev.Body.Bytes(), &evResp)

	base := "/api/v1/domains/governance.blumer.com/decisions/DEC-001/scenarios"
	rr := postJSONMethod(h, http.MethodPost, base,
		`{"name":"via-body","from_trace_id":"`+evResp.Trace.TraceID+`"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create-via-from_trace_id: code=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), evResp.Trace.TraceID) {
		t.Errorf("source_trace_id not echoed: %s", rr.Body.String())
	}
}

func TestDecisionScenarios_UnknownDecision(t *testing.T) {
	p := scenarioTestCosmos(t)
	h := NewHandler(p)
	rr := get(h, "/api/v1/domains/governance.blumer.com/decisions/DEC-NOPE/scenarios")
	if rr.Code != http.StatusNotFound {
		t.Errorf("unknown decision should 404 the scenarios listing: code=%d body=%s", rr.Code, rr.Body.String())
	}
}
