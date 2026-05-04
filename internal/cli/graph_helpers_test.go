package cli

import (
	"strings"
	"testing"
)

func TestMermaidID(t *testing.T) {
	if got := mermaidID("domain", "identity.blumer.cloud"); got != "domain_identity_blumer_cloud" {
		t.Fatalf("unexpected id: %s", got)
	}
	if got := mermaidID("service", "identity.blumer.cloud", "user-account"); got != "service_identity_blumer_cloud_user_account" {
		t.Fatalf("unexpected id: %s", got)
	}
	if got := mermaidID("123"); strings.HasPrefix(got, "1") {
		t.Fatalf("id starts with digit: %s", got)
	}
	if got := mermaidID("Hello---World"); strings.Contains(got, "-") {
		t.Fatalf("id contains hyphen: %s", got)
	}
}

func TestMermaidLabelEscaping(t *testing.T) {
	if got := mermaidLabel("Service \"A\""); !strings.Contains(got, `\"`) {
		t.Fatalf("expected escaped quote in %q", got)
	}
	if got := mermaidLabel("Line1\nLine2"); strings.Contains(got, "\n") {
		t.Fatalf("expected newline escaped in %q", got)
	}
}
