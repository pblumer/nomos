package app

import "github.com/nomos/nomos/internal/repo"

// ListRepositories returns the repositories the server manages (ADR-0022 §2).
// PR 1 covers the local server: a single default repository for the active
// workspace, its name taken from the workspace cosmos when present.
func ListRepositories(path string) (RepositoriesDTO, error) {
	cosmos, _ := GetCosmos(path)
	out := RepositoriesDTO{Repositories: []RepositoryDTO{}}
	for _, r := range repo.NewLocalRegistry(path).List() {
		name := r.Name
		if name == "" {
			name = fallback(cosmos.Name, "Local Repository")
		}
		out.Repositories = append(out.Repositories, RepositoryDTO{
			ID:            r.ID,
			Name:          name,
			Kind:          string(r.Kind),
			Location:      r.Location,
			DefaultBranch: r.DefaultBranch,
			Status:        r.Status,
			Head:          r.Head,
		})
	}
	return out, nil
}
