package app

import "testing"

func TestListRepositoriesReturnsLocalDefault(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := ListRepositories(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(dto.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(dto.Repositories))
	}
	r := dto.Repositories[0]
	if r.ID != "default" {
		t.Errorf("ID = %q, want default", r.ID)
	}
	if r.Kind != "filesystem" {
		t.Errorf("Kind = %q, want filesystem", r.Kind)
	}
	if r.Location != p {
		t.Errorf("Location = %q, want %q", r.Location, p)
	}
	if r.Name != "Local Cosmos" {
		t.Errorf("Name = %q, want Local Cosmos", r.Name)
	}
}

func TestListRepositoriesNameFallsBackWithoutCosmos(t *testing.T) {
	dto, err := ListRepositories(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(dto.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(dto.Repositories))
	}
	if dto.Repositories[0].Name != "Local Repository" {
		t.Errorf("Name = %q, want Local Repository", dto.Repositories[0].Name)
	}
}
