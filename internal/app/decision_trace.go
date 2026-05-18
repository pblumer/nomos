package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/dmn"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	versionpkg "github.com/nomos/nomos/internal/version"
)

// tracesDir returns the traces directory for a decision.
func tracesDir(decisionDir string) string {
	return filepath.Join(decisionDir, "traces")
}

// traceFileName builds a filesystem-sortable file name for a trace:
// <RFC3339-with-dashes>__<short-hash>.yaml
func traceFileName(t model.DecisionTrace) string {
	ts := strings.ReplaceAll(t.Timestamp, ":", "-")
	short := t.TraceID
	if idx := strings.Index(short, ":"); idx >= 0 && idx+13 <= len(short) {
		short = short[idx+1 : idx+13]
	}
	return ts + "__" + short + ".yaml"
}

// findDecisionNode locates a decision node by ID within a domain.
func findDecisionNode(path, domainCanonical, id string) (cosmosfs.DecisionNode, error) {
	d, err := findDomainNode(path, domainCanonical)
	if err != nil {
		return cosmosfs.DecisionNode{}, err
	}
	nodes, err := cosmosfs.ScanDecisions(d.Path)
	if err != nil {
		return cosmosfs.DecisionNode{}, err
	}
	for _, n := range nodes {
		if n.Metadata.ID == id {
			return n, nil
		}
	}
	return cosmosfs.DecisionNode{}, Error(CodeInvalidInput, "Decision not found: "+id, http.StatusNotFound, nil)
}

// loadTraces reads all trace YAMLs from a decision's traces/ directory, sorted by timestamp ascending.
func loadTraces(decisionDir string) ([]model.DecisionTrace, error) {
	root := tracesDir(decisionDir)
	ents, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []model.DecisionTrace
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".yml") {
			continue
		}
		var tr model.DecisionTrace
		if err := fsx.ReadYAML(filepath.Join(root, e.Name()), &tr); err != nil {
			return nil, fmt.Errorf("read trace %s: %w", e.Name(), err)
		}
		out = append(out, tr)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Timestamp != out[j].Timestamp {
			return out[i].Timestamp < out[j].Timestamp
		}
		return out[i].TraceID < out[j].TraceID
	})
	return out, nil
}

// writeTrace persists a trace YAML and returns its full path.
func writeTrace(decisionDir string, t model.DecisionTrace) (string, error) {
	dir := tracesDir(decisionDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, traceFileName(t))
	if err := fsx.WriteYAML(p, t); err != nil {
		return "", err
	}
	return p, nil
}

// EvaluateDecisionWithTrace evaluates a decision, writes a trace artifact, and returns both.
// Behaviour: if writing the trace fails the evaluation result is still returned alongside the error
// so callers can decide whether to surface it.
func EvaluateDecisionWithTrace(path, domainCanonical, id string, req EvaluateDecisionRequest, evaluator model.Evaluator) (*dmn.Result, *model.DecisionTrace, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return nil, nil, err
	}
	if n.DMNPath == "" {
		return nil, nil, Error(CodeInvalidInput, "No DMN file for decision: "+id, http.StatusNotFound, nil)
	}
	ruleBytes, err := os.ReadFile(n.DMNPath)
	if err != nil {
		return nil, nil, err
	}
	table, err := dmn.ParseDMN(ruleBytes)
	if err != nil {
		return nil, nil, Error(CodeInvalidInput, "DMN parse error: "+err.Error(), http.StatusUnprocessableEntity, err)
	}
	result, err := dmn.Evaluate(table, req.Inputs)
	if err != nil {
		return nil, nil, Error(CodeInvalidInput, "DMN evaluation error: "+err.Error(), http.StatusUnprocessableEntity, err)
	}

	// Determine parent trace from the chain tip (latest existing trace).
	existing, _ := loadTraces(n.Path)
	parent := ""
	if len(existing) > 0 {
		parent = existing[len(existing)-1].TraceID
	}

	info := versionpkg.Get()
	trace, err := dmn.BuildTrace(dmn.TraceBuildInput{
		Domain:          domainCanonical,
		DecisionID:      n.Metadata.ID,
		DecisionName:    n.Metadata.Name,
		DecisionVersion: n.Metadata.Version,
		RuleBytes:       ruleBytes,
		Inputs:          req.Inputs,
		Result:          result,
		Evaluator:       evaluator,
		Engine:          model.EngineInfo{Name: info.Name, Version: info.Version, Commit: info.Commit},
		Timestamp:       time.Now().UTC().Format(time.RFC3339Nano),
		ParentTraceID:   parent,
	})
	if err != nil {
		return result, nil, fmt.Errorf("build trace: %w", err)
	}
	if _, err := writeTrace(n.Path, trace); err != nil {
		return result, &trace, fmt.Errorf("write trace: %w", err)
	}
	return result, &trace, nil
}

// ListDecisionTraces returns all traces for a decision, oldest first.
func ListDecisionTraces(path, domainCanonical, id string) (DecisionTracesDTO, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return DecisionTracesDTO{}, err
	}
	traces, err := loadTraces(n.Path)
	if err != nil {
		return DecisionTracesDTO{}, err
	}
	items := make([]model.DecisionTrace, 0, len(traces))
	items = append(items, traces...)
	return DecisionTracesDTO{
		Domain:     domainCanonical,
		DecisionID: id,
		Items:      items,
		Count:      len(items),
	}, nil
}

// GetDecisionTrace returns a single trace by its trace_id.
func GetDecisionTrace(path, domainCanonical, id, traceID string) (model.DecisionTrace, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return model.DecisionTrace{}, err
	}
	traces, err := loadTraces(n.Path)
	if err != nil {
		return model.DecisionTrace{}, err
	}
	for _, t := range traces {
		if t.TraceID == traceID {
			return t, nil
		}
	}
	return model.DecisionTrace{}, Error(CodeInvalidInput, "Trace not found: "+traceID, http.StatusNotFound, nil)
}

// VerifyDecisionTraces recomputes each trace's hash and validates the parent chain.
// Returns a per-trace status list and an overall ok flag.
func VerifyDecisionTraces(path, domainCanonical, id string) (DecisionTraceVerifyDTO, error) {
	n, err := findDecisionNode(path, domainCanonical, id)
	if err != nil {
		return DecisionTraceVerifyDTO{}, err
	}
	traces, err := loadTraces(n.Path)
	if err != nil {
		return DecisionTraceVerifyDTO{}, err
	}
	results := make([]TraceVerifyEntryDTO, 0, len(traces))
	overallOK := true
	var prev string
	for _, t := range traces {
		entry := TraceVerifyEntryDTO{TraceID: t.TraceID, Timestamp: t.Timestamp}
		if err := dmn.VerifyTrace(t); err != nil {
			entry.Issue = err.Error()
			overallOK = false
		} else if t.ParentTraceID != prev {
			entry.Issue = fmt.Sprintf("parent chain broken: expected parent=%q got %q", prev, t.ParentTraceID)
			overallOK = false
		} else {
			entry.OK = true
		}
		results = append(results, entry)
		prev = t.TraceID
	}
	return DecisionTraceVerifyDTO{
		Domain:     domainCanonical,
		DecisionID: id,
		Count:      len(results),
		OK:         overallOK,
		Items:      results,
	}, nil
}
