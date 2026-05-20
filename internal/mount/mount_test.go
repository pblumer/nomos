package mount

import (
	"errors"
	"testing"
)

func TestStoreAddListRemove(t *testing.T) {
	s := NewStore(t.TempDir())

	if got, err := s.Mounts(); err != nil || len(got) != 0 {
		t.Fatalf("empty store: got %v err %v", got, err)
	}

	added, err := s.Add(Mount{Endpoint: "nomos.blumer.cloud:7373", Label: "Prod"})
	if err != nil {
		t.Fatal(err)
	}
	if added.ID == "" {
		t.Fatal("expected derived id")
	}

	got, err := s.Mounts()
	if err != nil || len(got) != 1 || got[0].Endpoint != "nomos.blumer.cloud:7373" {
		t.Fatalf("after add: got %v err %v", got, err)
	}

	if _, err := s.Add(Mount{Endpoint: "nomos.blumer.cloud:7373"}); !errors.Is(err, ErrExists) {
		t.Fatalf("duplicate endpoint err = %v, want ErrExists", err)
	}

	if err := s.Remove(added.ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Mounts(); len(got) != 0 {
		t.Fatalf("after remove: got %v", got)
	}

	if err := s.Remove("does-not-exist"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("remove missing err = %v, want ErrNotFound", err)
	}
	if err := s.Remove(LocalID); !errors.Is(err, ErrLocalReserved) {
		t.Fatalf("remove local err = %v, want ErrLocalReserved", err)
	}
	if _, err := s.Add(Mount{ID: LocalID, Endpoint: "x:7373"}); !errors.Is(err, ErrLocalReserved) {
		t.Fatalf("add reserved id err = %v, want ErrLocalReserved", err)
	}
}
