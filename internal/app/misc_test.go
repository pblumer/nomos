package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
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
	got, err := AddService(p, "auth-api", "Auth Team", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "auth-api" {
		t.Fatalf("expected name auth-api, got %s", got.Name)
	}
}

func TestAddService_EmptyName(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddService(p, "", "Team", false)
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestAddService_InvalidName(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddService(p, "bad/name", "Team", false)
	if err == nil {
		t.Fatal("expected error for name with slash")
	}
}

func TestAddService_DuplicateWithoutForce(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddService(p, "my-svc", "Team", false); err != nil {
		t.Fatal(err)
	}
	_, err := AddService(p, "my-svc", "Team", false)
	if err == nil {
		t.Fatal("expected error for duplicate service without force")
	}
}

func TestAddService_DuplicateWithForce(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddService(p, "my-svc", "Team", false); err != nil {
		t.Fatal(err)
	}
	_, err := AddService(p, "my-svc", "New Team", true)
	if err != nil {
		t.Fatalf("force should overwrite: %v", err)
	}
}

// ── RenameService ─────────────────────────────────────────────────────────────

func TestRenameService_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	if err := RenameService(p, "user-account", "user-accounts"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := GetService(p, "user-accounts"); err != nil {
		t.Fatalf("renamed service not found: %v", err)
	}
	if _, err := GetService(p, "user-account"); err == nil {
		t.Fatal("old service should no longer exist")
	}
}

func TestRenameService_InvalidNewName(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "user-account", "bad/name")
	if err == nil {
		t.Fatal("expected error for invalid new name")
	}
}

func TestRenameService_EmptyNewName(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "user-account", "")
	if err == nil {
		t.Fatal("expected error for empty new name")
	}
}

func TestRenameService_OldNotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "does-not-exist", "new-name")
	if err == nil {
		t.Fatal("expected error when old service not found")
	}
}

func TestRenameService_NewAlreadyExists(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "user-account", "privileged-account")
	if err == nil {
		t.Fatal("expected error when new service name already exists")
	}
}
