package idmigrate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/storage"
)

// Candidate beschreibt ein Artefakt mit Legacy-ID, das migriert werden kann.
type Candidate struct {
	Path         string // absoluter Pfad zur YAML-Datei
	ArtefactType string // Wert des type:-Feldes (z. B. "product_blueprint")
	OldID        string // aktuelle ID im Artefakt
}

// Scan durchläuft die Artefakt-Verzeichnisse eines Cosmos und liefert
// alle Artefakte mit Legacy-IDs zurück. Kandidaten sind Dateien, deren
// top-level `id:` Feld nicht dem ADR-0020-Format entspricht.
//
// Cosmos selbst wird nicht als Kandidat zurückgegeben (Sonderfall COS_root).
func Scan(workspace string) ([]Candidate, error) {
	roots := []string{
		storage.CatalogDirForRead(workspace),
		storage.DomainsDirForRead(workspace),
	}
	var candidates []Candidate
	seen := map[string]bool{}
	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if !isYAML(p) {
				return nil
			}
			if seen[p] {
				return nil
			}
			seen[p] = true
			c, ok, err := inspectFile(p)
			if err != nil {
				return err
			}
			if ok {
				candidates = append(candidates, c)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return candidates, nil
}

func isYAML(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	return ext == ".yaml" || ext == ".yml"
}

// inspectFile liest die obersten id/type-Felder aus einer YAML-Datei.
func inspectFile(path string) (Candidate, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Candidate{}, false, err
	}
	var header struct {
		ID   string `yaml:"id"`
		Type string `yaml:"type"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil {
		// Ungueltige YAML-Datei oder Datei ohne Top-Level-Map (z. B. DMN-XML
		// dass faelschlicherweise als .yaml liegt). Ueberspringen.
		return Candidate{}, false, nil
	}
	if strings.TrimSpace(header.ID) == "" {
		return Candidate{}, false, nil
	}
	if !idgen.IsLegacy(header.ID) {
		return Candidate{}, false, nil
	}
	artefactType := strings.TrimSpace(header.Type)
	if artefactType == "" {
		artefactType = inferTypeFromPath(path)
	}
	return Candidate{
		Path:         path,
		ArtefactType: artefactType,
		OldID:        strings.TrimSpace(header.ID),
	}, true, nil
}

// inferTypeFromPath rät den Artefakttyp aus dem Verzeichnis, wenn das
// type:-Feld in der YAML fehlt. Deckt Legacy-Artefakte ab, die noch ohne
// explizites type-Feld geschrieben wurden.
func inferTypeFromPath(p string) string {
	parts := strings.Split(filepath.ToSlash(p), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		switch parts[i] {
		case "requirements":
			return "requirement"
		case "rules":
			return "rule"
		case "products":
			// Kontext: "blueprints/products" oder "instances/products".
			if i > 0 && parts[i-1] == "instances" {
				return "product_instance"
			}
			if i > 0 && parts[i-1] == "blueprints" {
				return "product_blueprint"
			}
			return "product"
		case "services":
			if i > 0 && parts[i-1] == "instances" {
				return "service_instance"
			}
			if i > 0 && parts[i-1] == "blueprints" {
				return "service_blueprint"
			}
			return "service"
		case "servicegraphs":
			return "servicegraph"
		case "decisions":
			return "decision"
		}
	}
	return ""
}
