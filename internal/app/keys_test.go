package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func keysTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".nomos"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCreateAndListKeys(t *testing.T) {
	dir := keysTestDir(t)
	res, err := CreateKey(dir, "ci-token")
	if err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	if !strings.HasPrefix(res.Key, "nomos_") {
		t.Errorf("unexpected key prefix: %q", res.Key)
	}
	if res.Name != "ci-token" {
		t.Errorf("unexpected name: %q", res.Name)
	}

	keys, err := ListKeys(dir)
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if len(keys) != 1 || keys[0].Name != "ci-token" {
		t.Errorf("unexpected keys: %+v", keys)
	}
}

func TestCreateKey_DuplicateRejected(t *testing.T) {
	dir := keysTestDir(t)
	if _, err := CreateKey(dir, "dup"); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateKey(dir, "dup"); err == nil {
		t.Error("expected error for duplicate key name")
	}
}

func TestValidateKey(t *testing.T) {
	dir := keysTestDir(t)
	res, err := CreateKey(dir, "valid-key")
	if err != nil {
		t.Fatal(err)
	}
	if !ValidateKey(dir, res.Key) {
		t.Error("expected ValidateKey to return true for valid key")
	}
	if ValidateKey(dir, "nomos_invalid") {
		t.Error("expected ValidateKey to return false for unknown key")
	}
}

func TestValidateKey_Empty(t *testing.T) {
	dir := keysTestDir(t)
	if ValidateKey(dir, "") {
		t.Error("expected false for empty key")
	}
}

func TestRevokeKey(t *testing.T) {
	dir := keysTestDir(t)
	if _, err := CreateKey(dir, "to-revoke"); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateKey(dir, "keep"); err != nil {
		t.Fatal(err)
	}
	if err := RevokeKey(dir, "to-revoke"); err != nil {
		t.Fatalf("RevokeKey: %v", err)
	}
	keys, err := ListKeys(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0].Name != "keep" {
		t.Errorf("unexpected keys after revoke: %+v", keys)
	}
}

func TestRevokeKey_NotFound(t *testing.T) {
	dir := keysTestDir(t)
	if err := RevokeKey(dir, "ghost"); err == nil {
		t.Error("expected error revoking nonexistent key")
	}
}

func TestListKeys_Empty(t *testing.T) {
	dir := keysTestDir(t)
	keys, err := ListKeys(dir)
	if err != nil {
		t.Fatalf("ListKeys on empty dir: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty, got %v", keys)
	}
}
