package repo

import "testing"

func TestLocalRegistryReturnsDefaultRepository(t *testing.T) {
	ws := t.TempDir()
	repos := NewLocalRegistry(ws).List()
	if len(repos) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(repos))
	}
	r := repos[0]
	if r.ID != LocalDefaultID {
		t.Errorf("ID = %q, want %q", r.ID, LocalDefaultID)
	}
	if r.Kind != KindFilesystem {
		t.Errorf("Kind = %q, want %q", r.Kind, KindFilesystem)
	}
	if r.Location != ws {
		t.Errorf("Location = %q, want %q", r.Location, ws)
	}
	if r.Status != "unknown" {
		t.Errorf("Status = %q, want unknown for a non-git workspace", r.Status)
	}
}
