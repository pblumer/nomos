package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

// ── AppError ─────────────────────────────────────────────────────────────────

func TestAppError_ErrorString(t *testing.T) {
	err := Error(CodeDomainNotFound, "Domain not found: foo", 404, nil)
	if err.Error() != "DOMAIN_NOT_FOUND: Domain not found: foo" {
		t.Fatalf("unexpected error string: %s", err.Error())
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := Error(CodeInternalError, "wrapped", 500, cause)
	if err.Unwrap() != cause {
		t.Fatal("expected Unwrap to return the cause")
	}
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	err := Error(CodeInvalidInput, "no cause", 400, nil)
	if err.Unwrap() != nil {
		t.Fatal("expected nil Unwrap for error without cause")
	}
}

func TestErrorResponse_AppError(t *testing.T) {
	err := Error(CodeBlueprintNotFound, "not found", 404, nil)
	resp := ErrorResponse(err)
	inner, ok := resp["error"].(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string inner, got %T", resp["error"])
	}
	if inner["code"] != CodeBlueprintNotFound {
		t.Fatalf("expected code %s, got %s", CodeBlueprintNotFound, inner["code"])
	}
}

func TestErrorResponse_PlainError(t *testing.T) {
	err := errors.New("plain error")
	resp := ErrorResponse(err)
	inner, ok := resp["error"].(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string inner, got %T", resp["error"])
	}
	if inner["code"] != CodeInternalError {
		t.Fatalf("expected INTERNAL_ERROR, got %s", inner["code"])
	}
}

func TestAsAppError_Nil(t *testing.T) {
	ae, ok := AsAppError(nil)
	if ok || ae != nil {
		t.Fatal("expected (nil, false) for nil error")
	}
}

// ── DoctorCosmos ─────────────────────────────────────────────────────────────

func TestDoctorCosmos_ValidCosmos(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := DoctorCosmos(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.Status == "" {
		t.Fatal("expected non-empty status")
	}
	if len(dto.Checks) == 0 {
		t.Fatal("expected at least one check")
	}
	for _, c := range dto.Checks {
		if c.Name == ".nomos/cosmos.yaml" && c.Status != "ok" {
			t.Fatalf("cosmos.yaml check should be ok, got %s", c.Status)
		}
	}
}

func TestDoctorCosmos_MissingCosmos(t *testing.T) {
	p := t.TempDir()
	dto, err := DoctorCosmos(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range dto.Checks {
		if c.Name == ".nomos/cosmos.yaml" && c.Status == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected error check for missing cosmos.yaml")
	}
	if dto.Status != "error" {
		t.Fatalf("expected overall status error, got %s", dto.Status)
	}
}

func TestDoctorCosmos_WithGitRepo(t *testing.T) {
	p := createAppTestCosmos(t)
	if err := os.MkdirAll(filepath.Join(p, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	dto, err := DoctorCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range dto.Checks {
		if c.Name == "git repository" && c.Status != "ok" {
			t.Fatalf("git check should be ok, got %s", c.Status)
		}
	}
}

// ── AddService ────────────────────────────────────────────────────────────────

func TestAddService_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	got, err := AddService(p, "identity.blumer.cloud", "auth-api", "Auth Team", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "auth-api" {
		t.Fatalf("expected name auth-api, got %s", got.Name)
	}
	if got.Domain != "identity.blumer.cloud" {
		t.Fatalf("unexpected domain: %s", got.Domain)
	}
}

func TestAddService_EmptyName(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddService(p, "identity.blumer.cloud", "", "Team", false)
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestAddService_InvalidName(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddService(p, "identity.blumer.cloud", "bad/name", "Team", false)
	if err == nil {
		t.Fatal("expected error for name with slash")
	}
}

func TestAddService_DuplicateWithoutForce(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddService(p, "identity.blumer.cloud", "my-svc", "Team", false); err != nil {
		t.Fatal(err)
	}
	_, err := AddService(p, "identity.blumer.cloud", "my-svc", "Team", false)
	if err == nil {
		t.Fatal("expected error for duplicate service without force")
	}
}

func TestAddService_DuplicateWithForce(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddService(p, "identity.blumer.cloud", "my-svc", "Team", false); err != nil {
		t.Fatal(err)
	}
	_, err := AddService(p, "identity.blumer.cloud", "my-svc", "New Team", true)
	if err != nil {
		t.Fatalf("force should overwrite: %v", err)
	}
}

func TestAddService_DomainNotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddService(p, "does-not-exist.example", "svc", "Team", false)
	if err == nil {
		t.Fatal("expected error for non-existent domain")
	}
}

// createRenameCosmos builds a cosmos with properly tree-structured domains
// (created via AddDomain so the directory layout matches what RenameDomain expects).
func createRenameCosmos(t *testing.T) string {
	t.Helper()
	p := createTestCosmosWithCatalog(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := AddDomain(p, "platform.blumer.cloud", "Platform Team", false)
	must(err)
	_, err = AddDomain(p, "identity.blumer.cloud", "Identity Team", false)
	must(err)
	return p
}

// ── RenameDomain ─────────────────────────────────────────────────────────────

func TestRenameDomain_Success(t *testing.T) {
	p := createRenameCosmos(t)
	if err := RenameDomain(p, "platform.blumer.cloud", "platform2.blumer.cloud"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := GetDomain(p, "platform2.blumer.cloud"); err != nil {
		t.Fatalf("renamed domain not found: %v", err)
	}
	if _, err := GetDomain(p, "platform.blumer.cloud"); err == nil {
		t.Fatal("old domain should no longer exist")
	}
}

func TestRenameDomain_InvalidNewName(t *testing.T) {
	p := createRenameCosmos(t)
	err := RenameDomain(p, "platform.blumer.cloud", "bad/name")
	if err == nil {
		t.Fatal("expected error for invalid new name")
	}
}

func TestRenameDomain_OldNotFound(t *testing.T) {
	p := createRenameCosmos(t)
	err := RenameDomain(p, "nonexistent.blumer.cloud", "new.blumer.cloud")
	if err == nil {
		t.Fatal("expected error when old domain not found")
	}
}

func TestRenameDomain_NewAlreadyExists(t *testing.T) {
	p := createRenameCosmos(t)
	err := RenameDomain(p, "platform.blumer.cloud", "identity.blumer.cloud")
	if err == nil {
		t.Fatal("expected error when new name already exists")
	}
}

// ── RenameService ─────────────────────────────────────────────────────────────

func TestRenameService_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	if err := RenameService(p, "identity.blumer.cloud", "user-account", "user-accounts"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := GetService(p, "identity.blumer.cloud", "user-accounts"); err != nil {
		t.Fatalf("renamed service not found: %v", err)
	}
	if _, err := GetService(p, "identity.blumer.cloud", "user-account"); err == nil {
		t.Fatal("old service should no longer exist")
	}
}

func TestRenameService_InvalidNewName(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "identity.blumer.cloud", "user-account", "bad/name")
	if err == nil {
		t.Fatal("expected error for invalid new name")
	}
}

func TestRenameService_EmptyNewName(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "identity.blumer.cloud", "user-account", "")
	if err == nil {
		t.Fatal("expected error for empty new name")
	}
}

func TestRenameService_OldNotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "identity.blumer.cloud", "does-not-exist", "new-name")
	if err == nil {
		t.Fatal("expected error when old service not found")
	}
}

func TestRenameService_NewAlreadyExists(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "identity.blumer.cloud", "user-account", "privileged-account")
	if err == nil {
		t.Fatal("expected error when new service name already exists")
	}
}

func TestRenameService_DomainNotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "does-not-exist.example", "svc", "new")
	if err == nil {
		t.Fatal("expected error for non-existent domain")
	}
}

// ── ListVerificationEvidence ──────────────────────────────────────────────────

func TestListVerificationEvidence_EmptyDir(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := ListVerificationEvidence(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.Count != 0 {
		t.Fatalf("expected 0 evidence, got %d", dto.Count)
	}
}

func TestListVerificationEvidence_WithEvidence(t *testing.T) {
	p := createAppTestCosmos(t)
	evidenceDir := storage.EvidenceDir(p)
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evidenceDir, "ev-001.yaml"),
		[]byte("id: ev-001\ntype: dns_txt\ndomain: example.com\nstatus: verified\ntimestamp: 2024-01-01T00:00:00Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dto, err := ListVerificationEvidence(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.Count != 1 {
		t.Fatalf("expected 1 evidence, got %d", dto.Count)
	}
	if dto.Evidence[0].ID != "ev-001" {
		t.Fatalf("unexpected evidence ID: %s", dto.Evidence[0].ID)
	}
}

func TestListVerificationEvidence_SkipsNonYAML(t *testing.T) {
	p := createAppTestCosmos(t)
	evidenceDir := storage.EvidenceDir(p)
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evidenceDir, "notes.txt"), []byte("ignore me"), 0o644); err != nil {
		t.Fatal(err)
	}
	dto, err := ListVerificationEvidence(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto.Count != 0 {
		t.Fatalf("expected 0 evidence (txt file ignored), got %d", dto.Count)
	}
}
