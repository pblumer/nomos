package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func folderTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	if err := os.MkdirAll(storage.DomainsDir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(storage.CosmosFile(p), []byte("id: c\nname: C\nversion: 0.1.0\nstatus: draft\nowner: t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCreateAndGetFolder(t *testing.T) {
	p := folderTestCosmos(t)
	f, err := CreateFolder(p, "", "cloud")
	if err != nil {
		t.Fatal(err)
	}
	if f.Canonical != "cloud" {
		t.Fatalf("canonical = %q, want cloud", f.Canonical)
	}
	child, err := CreateFolder(p, "cloud", "blumer")
	if err != nil {
		t.Fatal(err)
	}
	if child.Canonical != "blumer.cloud" {
		t.Fatalf("canonical = %q, want blumer.cloud", child.Canonical)
	}
	if _, err := os.Stat(storage.FolderFile(child.Path)); err != nil {
		t.Fatalf("folder.yaml should materialize the dir: %v", err)
	}
	if got, err := GetFolder(p, "blumer.cloud"); err != nil || got.Label != "blumer" {
		t.Fatalf("GetFolder = %+v err %v", got, err)
	}
	if _, err := CreateFolder(p, "cloud", "blumer"); err == nil {
		t.Fatal("duplicate folder should fail")
	}
}

func TestMoveFolderReparentsSubtree(t *testing.T) {
	p := folderTestCosmos(t)
	for _, c := range [][2]string{{"", "cloud"}, {"cloud", "blumer"}, {"blumer.cloud", "identity"}, {"", "acme"}} {
		if _, err := CreateFolder(p, c[0], c[1]); err != nil {
			t.Fatalf("create %v: %v", c, err)
		}
	}
	// Move blumer.cloud under acme → blumer.acme, child identity.blumer.cloud → identity.blumer.acme.
	if _, err := MoveFolder(p, "blumer.cloud", "acme"); err != nil {
		t.Fatal(err)
	}
	if _, err := GetFolder(p, "blumer.acme"); err != nil {
		t.Fatalf("moved folder should be blumer.acme: %v", err)
	}
	if _, err := GetFolder(p, "identity.blumer.acme"); err != nil {
		t.Fatalf("descendant should follow to identity.blumer.acme: %v", err)
	}
	if _, err := GetFolder(p, "blumer.cloud"); err == nil {
		t.Fatal("old blumer.cloud should be gone")
	}
}

func TestRenameAndDeleteFolder(t *testing.T) {
	p := folderTestCosmos(t)
	if _, err := CreateFolder(p, "", "cloud"); err != nil {
		t.Fatal(err)
	}
	if _, err := RenameFolder(p, "cloud", "wolke"); err != nil {
		t.Fatal(err)
	}
	if _, err := GetFolder(p, "wolke"); err != nil {
		t.Fatalf("renamed folder should be wolke: %v", err)
	}
	if err := DeleteFolder(p, "wolke"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(storage.DomainsDir(p), "wolke")); !os.IsNotExist(err) {
		t.Fatal("folder dir should be removed")
	}
}
