package storage

import (
	"os"
	"path/filepath"
)

const DirName = ".nomos"

// FolderMetaName is the per-folder metadata file that declares which artifact
// types may be created in a folder (and optional label/description).
const FolderMetaName = ".nomos.folder.yaml"

// FolderMetaFile returns the metadata file path for a directory.
func FolderMetaFile(dir string) string { return filepath.Join(dir, FolderMetaName) }

func NomosDir(workspace string) string     { return filepath.Join(workspace, DirName) }
func CosmosFile(workspace string) string   { return filepath.Join(NomosDir(workspace), "cosmos.yaml") }
func ServicesDir(workspace string) string  { return filepath.Join(NomosDir(workspace), "services") }
func DecisionsDir(workspace string) string { return filepath.Join(NomosDir(workspace), "decisions") }
func CatalogDir(workspace string) string   { return filepath.Join(NomosDir(workspace), "catalog") }

// TypesDir holds user-defined artifact type definitions (.nomos/types/<id>.yaml).
func TypesDir(workspace string) string { return filepath.Join(NomosDir(workspace), "types") }

// ViewsDir holds standalone form/view definitions (.nomos/views/<name>.frm).
func ViewsDir(workspace string) string { return filepath.Join(NomosDir(workspace), "views") }

// DataDir holds reusable data-object (table-schema) definitions
// (.nomos/data/<id>.yaml) that types reference as their main/helper objects.
func DataDir(workspace string) string { return filepath.Join(NomosDir(workspace), "data") }
func ServicegraphsDir(workspace string) string {
	return filepath.Join(NomosDir(workspace), "servicegraphs")
}
func CatalogServicegraphsDir(workspace string) string {
	return filepath.Join(CatalogDir(workspace), "servicegraphs")
}
func EvidenceDir(workspace string) string { return filepath.Join(NomosDir(workspace), "evidence") }
func IndexDir(workspace string) string    { return filepath.Join(NomosDir(workspace), "index") }
func CacheDir(workspace string) string    { return filepath.Join(NomosDir(workspace), "cache") }
func UCIDir(workspace string) string      { return filepath.Join(NomosDir(workspace), "uci") }
func KeysFile(workspace string) string    { return filepath.Join(NomosDir(workspace), "keys.yaml") }
func MountsFile(workspace string) string  { return filepath.Join(NomosDir(workspace), "mounts.yaml") }

// ServerConfigFile is the optional, file-based identity of a Nomos server
// (.nomos/server.yml): its public domain and a display label. Non-authoritative
// operator config; absent by default. The NOMOS_DOMAIN environment variable
// still overrides the file's domain.
func ServerConfigFile(workspace string) string {
	return filepath.Join(NomosDir(workspace), "server.yml")
}

// RepositoriesFile lists the additional repositories a server manages beyond
// its default workspace (ADR-0022 §2). Non-authoritative server config.
func RepositoriesFile(workspace string) string {
	return filepath.Join(NomosDir(workspace), "repositories.yaml")
}

// ReposDirEnv lets an operator point new local filesystem repositories at a
// predefined, writable base directory instead of the default ~/.nomos-repos.
// Only the operator sets this; API clients never supply a path, and repository
// ids are validated, so this cannot be abused for path traversal.
const ReposDirEnv = "NOMOS_REPOS_DIR"

// DefaultReposDirName is the home-relative base directory for new local
// filesystem repositories.
const DefaultReposDirName = ".nomos-repos"

// ReposDir is the base directory under which new local filesystem repositories
// are created. It lives outside the workspace (~/.nomos-repos) so that managed
// repositories are not nested inside the workspace's own git tree. The
// NOMOS_REPOS_DIR override takes precedence; when no home directory is
// resolvable it falls back to the workspace-local <workspace>/.nomos/repos.
func ReposDir(workspace string) string {
	if dir := os.Getenv(ReposDirEnv); dir != "" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, DefaultReposDirName)
	}
	return filepath.Join(NomosDir(workspace), "repos")
}

func SelfModelDir(workspace string) string {
	return filepath.Join(ServicesDir(workspace), "nomos-core")
}

func LegacyCosmosFile(workspace string) string { return filepath.Join(workspace, "cosmos.yaml") }
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
