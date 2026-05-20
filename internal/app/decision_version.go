package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/dmn"
	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

// versionsDir returns the directory where per-version snapshots live for a decision.
//
// Layout:
//
//	decisions/<id>/
//	  decision.yaml       ← HEAD pointer (always == latest version)
//	  decision.dmn        ← HEAD pointer
//	  versions/
//	    v0.1.0/{decision.yaml, decision.dmn}
//	    v0.1.1/{decision.yaml, decision.dmn}
//
// Each snapshot directory is self-contained so a trace's decision_version
// resolves to exactly the rule set that produced its outputs.
func versionsDir(decisionDir string) string {
	return filepath.Join(decisionDir, "versions")
}

// versionKey converts a semver string "0.1.2" into the directory name "v0.1.2".
// The leading "v" guards against awkward names like ".." and keeps tab-completion friendly.
func versionKey(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "v") || strings.HasPrefix(v, "V") {
		return "v" + v[1:]
	}
	return "v" + v
}

func versionDir(decisionDir, version string) string {
	return filepath.Join(versionsDir(decisionDir), versionKey(version))
}

// parseSemver parses "MAJOR.MINOR.PATCH" with optional leading "v".
// Pre-release / build suffixes are not supported and trigger fallback in bumpPatch.
func parseSemver(v string) (int, int, int, error) {
	s := strings.TrimSpace(v)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("not semver: %q", v)
	}
	nums := [3]int{}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return 0, 0, 0, fmt.Errorf("not semver: %q", v)
		}
		nums[i] = n
	}
	return nums[0], nums[1], nums[2], nil
}

// bumpPatch returns the next patch-level semver. Non-semver inputs are wrapped
// with a .1 suffix so the chain keeps moving rather than failing the save.
func bumpPatch(v string) string {
	major, minor, patch, err := parseSemver(v)
	if err != nil {
		base := strings.TrimSpace(v)
		if base == "" {
			return "0.1.0"
		}
		return base + ".1"
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
}

// semverLess sorts version strings by their numeric components. Non-semver
// values fall back to lexical compare so the list never panics.
func semverLess(a, b string) bool {
	aM, am, ap, aerr := parseSemver(a)
	bM, bm, bp, berr := parseSemver(b)
	if aerr != nil || berr != nil {
		return a < b
	}
	if aM != bM {
		return aM < bM
	}
	if am != bm {
		return am < bm
	}
	return ap < bp
}

// writeVersionSnapshot persists a self-contained copy of the decision for the
// given version. dmnBytes may be nil for decisions that don't have a DMN file
// yet (freshly created metadata-only artifact).
func writeVersionSnapshot(decisionDir string, dec model.Decision, dmnBytes []byte) error {
	dir := versionDir(decisionDir, dec.Version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := fsx.WriteYAML(filepath.Join(dir, "decision.yaml"), dec); err != nil {
		return err
	}
	if dmnBytes != nil {
		if err := os.WriteFile(filepath.Join(dir, "decision.dmn"), dmnBytes, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// metadataChanged reports whether two Decision values differ in any field a
// caller-controlled UpdateDecisionRequest can touch. The Version field is
// ignored because we drive bumps automatically.
func metadataChanged(a, b model.Decision) bool {
	a.Version = ""
	b.Version = ""
	return !reflect.DeepEqual(a, b)
}

// ListDecisionVersionsDTO returns the metadata of every snapshot, oldest first.
// The HEAD version is always included even when the versions/ directory does
// not exist yet — legacy decisions created before snapshotting still get a
// usable list with a single entry.
func ListDecisionVersions(path, id string) (DecisionVersionsDTO, error) {
	n, err := findDecisionNode(path, id)
	if err != nil {
		return DecisionVersionsDTO{}, err
	}
	items, err := readVersionSnapshots(n)
	if err != nil {
		return DecisionVersionsDTO{}, err
	}
	if len(items) == 0 {
		items = []DecisionVersionDTO{headAsVersion(n)}
	}
	sort.Slice(items, func(i, j int) bool { return semverLess(items[i].Version, items[j].Version) })
	return DecisionVersionsDTO{
		DecisionID: id,
		Current:    n.Metadata.Version,
		Items:      items,
		Count:      len(items),
	}, nil
}

func headAsVersion(n cosmosfs.DecisionNode) DecisionVersionDTO {
	v := DecisionVersionDTO{
		Version: n.Metadata.Version,
		Status:  n.Metadata.Status,
		HasDMN:  n.DMNPath != "",
	}
	if n.DMNPath != "" {
		if data, err := os.ReadFile(n.DMNPath); err == nil {
			v.RuleHash = dmn.HashSHA256(data)
		}
	}
	return v
}

func readVersionSnapshots(n cosmosfs.DecisionNode) ([]DecisionVersionDTO, error) {
	root := versionsDir(n.Path)
	ents, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]DecisionVersionDTO, 0, len(ents))
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		yamlPath := filepath.Join(dir, "decision.yaml")
		var dec model.Decision
		if err := fsx.ReadYAML(yamlPath, &dec); err != nil {
			continue
		}
		entry := DecisionVersionDTO{
			Version: dec.Version,
			Status:  dec.Status,
		}
		dmnPath := filepath.Join(dir, "decision.dmn")
		if data, err := os.ReadFile(dmnPath); err == nil {
			entry.HasDMN = true
			entry.RuleHash = dmn.HashSHA256(data)
		}
		out = append(out, entry)
	}
	return out, nil
}

// GetDecisionVersion returns the decision metadata as stored at the given version.
// Falls back to the HEAD record when the request targets the current version
// of a decision created before snapshotting existed.
func GetDecisionVersion(path, id, version string) (DecisionDTO, error) {
	n, err := findDecisionNode(path, id)
	if err != nil {
		return DecisionDTO{}, err
	}
	dir := versionDir(n.Path, version)
	yamlPath := filepath.Join(dir, "decision.yaml")
	var dec model.Decision
	if err := fsx.ReadYAML(yamlPath, &dec); err == nil {
		dn := cosmosfs.DecisionNode{Path: dir, Metadata: dec}
		if _, err := os.Stat(filepath.Join(dir, "decision.dmn")); err == nil {
			dn.DMNPath = filepath.Join(dir, "decision.dmn")
		}
		return decisionDTO(dn), nil
	}
	if n.Metadata.Version == version {
		return decisionDTO(n), nil
	}
	return DecisionDTO{}, Error(CodeInvalidInput, "Decision version not found: "+version, http.StatusNotFound, nil)
}

// GetDecisionVersionDMN serves the DMN XML for a specific historical version.
func GetDecisionVersionDMN(path, id, version string) (string, error) {
	n, err := findDecisionNode(path, id)
	if err != nil {
		return "", err
	}
	if data, err := os.ReadFile(filepath.Join(versionDir(n.Path, version), "decision.dmn")); err == nil {
		return string(data), nil
	}
	if n.Metadata.Version == version && n.DMNPath != "" {
		data, err := os.ReadFile(n.DMNPath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", Error(CodeInvalidInput, "Decision version not found: "+version, http.StatusNotFound, nil)
}

// GetDecisionVersionDefinitions parses the DMN file backing the given version
// and returns the full DMN 1.5 DRG. Used by the Cosmos Explorer to render a
// trace's decision table against the rules that were active at evaluation time.
func GetDecisionVersionDefinitions(path, id, version string) (*model.DMNDefinitions, error) {
	xml, err := GetDecisionVersionDMN(path, id, version)
	if err != nil {
		return nil, err
	}
	defs, err := dmn.ParseDefinitions([]byte(xml))
	if err != nil {
		return nil, Error(CodeInvalidInput, "DMN parse error: "+err.Error(), http.StatusUnprocessableEntity, err)
	}
	return defs, nil
}
