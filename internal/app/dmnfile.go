package app

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/nomos/nomos/internal/dmn"
	"github.com/nomos/nomos/internal/fsx"
)

// DMNFileSyncResult reports the outcome of writing a .dmn file and (optionally)
// syncing the decision I/O contract into a sibling decision.yaml.
type DMNFileSyncResult struct {
	OK      bool            `json:"ok"`
	Synced  bool            `json:"synced"`           // a sibling decision.yaml was updated
	Target  string          `json:"target,omitempty"` // sibling file that was synced
	Inputs  []DecisionIODTO `json:"inputs,omitempty"`
	Outputs []DecisionIODTO `json:"outputs,omitempty"`
}

// ReadDMNFile returns the raw XML of a .dmn file at an absolute, pre-validated
// workspace path.
func ReadDMNFile(absPath string) (string, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", Error(CodeInvalidInput, "DMN file not found", http.StatusNotFound, err)
		}
		return "", Error(CodeInternalError, "failed to read DMN file: "+err.Error(), http.StatusInternalServerError, err)
	}
	return string(data), nil
}

// SaveDMNFile normalizes and writes a .dmn file, then mirrors the DMN's I/O
// contract (inputData → inputs, top-level decision outputs → outputs) into a
// sibling decision.yaml when one exists. The DMN is the source of truth for the
// I/O contract; the sibling YAML's other keys are preserved untouched.
func SaveDMNFile(absPath, xml string) (DMNFileSyncResult, error) {
	newBytes := []byte(xml)
	if normalized, nerr := dmn.NormalizeInputIdentifiers(newBytes); nerr == nil {
		newBytes = normalized
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(absPath); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(absPath, newBytes, mode); err != nil {
		return DMNFileSyncResult{}, Error(CodeInternalError, "failed to write DMN file: "+err.Error(), http.StatusInternalServerError, err)
	}
	res := DMNFileSyncResult{OK: true}

	sibling := filepath.Join(filepath.Dir(absPath), "decision.yaml")
	if _, err := os.Stat(sibling); err != nil {
		return res, nil
	}
	defs, perr := dmn.ParseDefinitions(newBytes)
	if perr != nil {
		return res, nil // unparseable DMN: file is saved, but no sync
	}
	meta := map[string]any{}
	_ = fsx.ReadYAML(sibling, &meta)
	changed := false
	if len(defs.InputData) > 0 {
		ins := decisionIODTOs(inputsFromDMN(defs))
		meta["inputs"] = ioMaps(ins)
		res.Inputs = ins
		changed = true
	}
	if len(defs.Decisions) > 0 {
		outs := decisionIODTOs(outputsFromDMN(defs))
		meta["outputs"] = ioMaps(outs)
		res.Outputs = outs
		changed = true
	}
	if changed {
		if err := fsx.WriteYAML(sibling, meta); err == nil {
			res.Synced = true
			res.Target = "decision.yaml"
		}
	}
	return res, nil
}

// ioMaps renders decision I/O entries as ordered YAML maps, omitting empty
// optional keys so the synced decision.yaml stays clean.
func ioMaps(ios []DecisionIODTO) []map[string]any {
	out := make([]map[string]any, 0, len(ios))
	for _, v := range ios {
		m := map[string]any{"name": v.Name}
		if v.Type != "" {
			m["type"] = v.Type
		}
		if v.Description != "" {
			m["description"] = v.Description
		}
		out = append(out, m)
	}
	return out
}
