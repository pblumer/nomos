package storage

import (
	"os"
	"path/filepath"
)

const DirName = ".nomos"

func NomosDir(workspace string) string   { return filepath.Join(workspace, DirName) }
func CosmosFile(workspace string) string { return filepath.Join(NomosDir(workspace), "cosmos.yaml") }
func DomainsDir(workspace string) string { return filepath.Join(NomosDir(workspace), "domains") }
func CatalogDir(workspace string) string { return filepath.Join(NomosDir(workspace), "catalog") }
func ServicegraphsDir(workspace string) string {
	return filepath.Join(NomosDir(workspace), "servicegraphs")
}
func CatalogServicegraphsDir(workspace string) string {
	return filepath.Join(CatalogDir(workspace), "servicegraphs")
}
func EvidenceDir(workspace string) string { return filepath.Join(NomosDir(workspace), "evidence") }
func IndexDir(workspace string) string    { return filepath.Join(NomosDir(workspace), "index") }
func CacheDir(workspace string) string    { return filepath.Join(NomosDir(workspace), "cache") }
func UIDir(workspace string) string       { return filepath.Join(NomosDir(workspace), "ui") }
func KeysFile(workspace string) string    { return filepath.Join(NomosDir(workspace), "keys.yaml") }

func SelfModelDir(workspace string) string {
	return filepath.Join(DomainsDir(workspace), "nomos", "core")
}

func LegacyCosmosFile(workspace string) string { return filepath.Join(workspace, "cosmos.yaml") }
func LegacyDomainsDir(workspace string) string { return filepath.Join(workspace, "domains") }
func LegacyCatalogDir(workspace string) string { return filepath.Join(workspace, "catalog") }
func LegacyServicegraphsDir(workspace string) string {
	return filepath.Join(workspace, "servicegraphs")
}

func CosmosFileForRead(workspace string) (string, bool) {
	canonical := CosmosFile(workspace)
	if _, err := os.Stat(canonical); err == nil {
		return canonical, false
	}
	legacy := LegacyCosmosFile(workspace)
	if _, err := os.Stat(legacy); err == nil {
		return legacy, true
	}
	return canonical, false
}

func DomainsDirForRead(workspace string) string {
	if _, err := os.Stat(DomainsDir(workspace)); err == nil {
		return DomainsDir(workspace)
	}
	if _, err := os.Stat(LegacyDomainsDir(workspace)); err == nil {
		return LegacyDomainsDir(workspace)
	}
	return DomainsDir(workspace)
}

func CatalogDirForRead(workspace string) string {
	if _, err := os.Stat(CatalogDir(workspace)); err == nil {
		return CatalogDir(workspace)
	}
	if _, err := os.Stat(LegacyCatalogDir(workspace)); err == nil {
		return LegacyCatalogDir(workspace)
	}
	return CatalogDir(workspace)
}

func CatalogServicegraphsDirForRead(workspace string) string {
	return filepath.Join(CatalogDirForRead(workspace), "servicegraphs")
}

func LegacyLayoutDetected(workspace string) bool {
	if _, err := os.Stat(CosmosFile(workspace)); err == nil {
		return false
	}
	_, err := os.Stat(LegacyCosmosFile(workspace))
	return err == nil
}
