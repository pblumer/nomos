// Package selfmodel implements the Nomos self-model bundle (ADR-0014).
// The bundle is embedded in the binary and can be imported into any cosmos workspace.
package selfmodel

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

//go:embed bundle
var bundleFS embed.FS

// BundleMetadata holds the parsed bundle.yaml metadata.
type BundleMetadata struct {
	ID                     string   `yaml:"id"`
	Version                string   `yaml:"version"`
	NomosVersion           string   `yaml:"nomos_version"`
	Generated              string   `yaml:"generated"`
	SourceLocale           string   `yaml:"source_locale"`
	SupportedLocales       []string `yaml:"supported_locales"`
	ConnectorSchemaVersion string   `yaml:"connector_schema_version"`
	Manifest               []string `yaml:"manifest"`
}

// StatusResult summarises the self-model state for a workspace.
type StatusResult struct {
	// BinaryVersion is the version embedded in the running binary.
	BinaryVersion string
	// BinaryChecksum is the SHA-256 of all bundle files in the binary.
	BinaryChecksum string
	// WorkspaceVersion is the version recorded in cosmos.yaml, empty if not imported.
	WorkspaceVersion string
	// WorkspaceChecksum is the checksum recorded in cosmos.yaml.
	WorkspaceChecksum string
	// ActualChecksum is the live SHA-256 of workspace self-model files.
	ActualChecksum string
	// Case is one of A (fresh), B (no-op), C (upgrade needed), D (local changes), E (drift+version).
	Case string
	// Message is a human-readable summary.
	Message string
}

// LoadBundleMeta returns the parsed bundle.yaml from the embedded FS.
func LoadBundleMeta() (BundleMetadata, error) {
	var meta BundleMetadata
	if err := fsx.ReadYAMLFromFS(bundleFS, "bundle/bundle.yaml", &meta); err != nil {
		return BundleMetadata{}, fmt.Errorf("selfmodel: read bundle.yaml: %w", err)
	}
	return meta, nil
}

// BinaryChecksum computes a deterministic SHA-256 over the manifest files embedded in the binary.
// Uses the same file set as checksumWorkspace so the two checksums are directly comparable.
func BinaryChecksum() (string, error) {
	meta, err := LoadBundleMeta()
	if err != nil {
		return "", err
	}
	type entry struct{ path, content string }
	var entries []entry
	for _, rel := range meta.Manifest {
		data, err := bundleFS.ReadFile("bundle/" + rel)
		if err != nil {
			data = []byte{}
		}
		entries = append(entries, entry{rel, string(data)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s\n%s\n", e.path, e.content)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Import writes the embedded bundle into the workspace at cosmosPath.
// If forceOverwrite is false and workspace files exist with local changes (Case D/E),
// it returns an error instead of overwriting.
func Import(cosmosPath string, forceOverwrite bool) error {
	meta, err := LoadBundleMeta()
	if err != nil {
		return err
	}
	binaryCS, err := BinaryChecksum()
	if err != nil {
		return err
	}

	st, err := Status(cosmosPath)
	if err == nil {
		switch st.Case {
		case "B":
			return nil // already up-to-date
		case "D", "E":
			if !forceOverwrite {
				return fmt.Errorf("selfmodel: workspace has local changes to self-model files (case %s). Use --force-overwrite-self to overwrite", st.Case)
			}
		}
	}

	// Walk the embedded bundle and write each file to workspace.
	if err := fs.WalkDir(bundleFS, "bundle", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(path) == "bundle.yaml" {
			return nil
		}
		// Relative path within the bundle (strip "bundle/" prefix).
		rel := strings.TrimPrefix(path, "bundle/")
		dest := filepath.Join(storage.NomosDir(cosmosPath), filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		data, err := bundleFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	}); err != nil {
		return fmt.Errorf("selfmodel: import bundle: %w", err)
	}

	// Update cosmos.yaml with self_model block.
	cosmosFile, _ := storage.CosmosFileForRead(cosmosPath)
	var cosmos model.Cosmos
	if err := fsx.ReadYAML(cosmosFile, &cosmos); err != nil {
		return fmt.Errorf("selfmodel: read cosmos.yaml: %w", err)
	}
	cosmos.SelfModel = &model.SelfModelRef{
		Version:        meta.Version,
		BundleChecksum: binaryCS,
		ImportedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	if err := fsx.WriteYAML(cosmosFile, cosmos); err != nil {
		return fmt.Errorf("selfmodel: write cosmos.yaml: %w", err)
	}
	return nil
}

// Status computes the current self-model state for a workspace.
func Status(cosmosPath string) (StatusResult, error) {
	meta, err := LoadBundleMeta()
	if err != nil {
		return StatusResult{}, err
	}
	binaryCS, err := BinaryChecksum()
	if err != nil {
		return StatusResult{}, err
	}

	res := StatusResult{
		BinaryVersion:  meta.Version,
		BinaryChecksum: binaryCS,
	}

	// Read workspace cosmos.yaml for recorded self_model block.
	cosmosFile, _ := storage.CosmosFileForRead(cosmosPath)
	var cosmos model.Cosmos
	if err := fsx.ReadYAML(cosmosFile, &cosmos); err != nil || cosmos.SelfModel == nil {
		// Case A — not yet imported.
		res.Case = "A"
		res.Message = "Self-model not yet imported. Run: nomos self import"
		return res, nil
	}
	res.WorkspaceVersion = cosmos.SelfModel.Version
	res.WorkspaceChecksum = cosmos.SelfModel.BundleChecksum

	// Compute actual checksum of workspace self-model files.
	actualCS, err := checksumWorkspace(cosmosPath, meta.Manifest)
	if err != nil {
		actualCS = ""
	}
	res.ActualChecksum = actualCS

	binaryEqRecorded := binaryCS == res.WorkspaceChecksum
	recordedEqActual := res.WorkspaceChecksum == actualCS

	switch {
	case binaryEqRecorded && recordedEqActual:
		res.Case = "B"
		res.Message = fmt.Sprintf("Up-to-date (v%s)", res.WorkspaceVersion)
	case !binaryEqRecorded && recordedEqActual:
		res.Case = "C"
		res.Message = fmt.Sprintf("Binary v%s, workspace v%s. Run: nomos self upgrade", meta.Version, res.WorkspaceVersion)
	case binaryEqRecorded && !recordedEqActual:
		res.Case = "D"
		res.Message = "Workspace has local changes to self-model files. Use --force-overwrite-self to overwrite."
	default:
		res.Case = "E"
		res.Message = fmt.Sprintf("Workspace has local changes AND binary version differs (binary v%s, workspace v%s).", meta.Version, res.WorkspaceVersion)
	}
	return res, nil
}

// checksumFS computes a deterministic SHA-256 over all files in an embed.FS subtree.
func checksumFS(fsys embed.FS, root string) (string, error) {
	type entry struct{ path, content string }
	var entries []entry
	if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fsys.ReadFile(path)
		if err != nil {
			return err
		}
		entries = append(entries, entry{path, string(data)})
		return nil
	}); err != nil {
		return "", err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s\n%s\n", e.path, e.content)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// checksumWorkspace computes SHA-256 over the manifest files in the workspace.
func checksumWorkspace(cosmosPath string, manifest []string) (string, error) {
	type entry struct{ path, content string }
	var entries []entry
	nomosDir := storage.NomosDir(cosmosPath)
	for _, rel := range manifest {
		dest := filepath.Join(nomosDir, filepath.FromSlash(rel))
		data, err := os.ReadFile(dest)
		if err != nil {
			// Missing file counts as empty for checksum purposes.
			data = []byte{}
		}
		entries = append(entries, entry{rel, string(data)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s\n%s\n", e.path, e.content)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
