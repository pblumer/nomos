package app

import (
	"testing"
)

func TestValidDomainSegment(t *testing.T) {
	valid := []string{"identity", "user-account", "abc123", "a", "my-service"}
	for _, s := range valid {
		if !validDomainSegment(s) {
			t.Errorf("expected valid segment: %q", s)
		}
	}
	invalid := []string{
		"",
		"identity.cloud",
		"has space",
		"has/slash",
		"-leading",
		"trailing-",
		"UPPERCASE",
		"with.dot",
	}
	for _, s := range invalid {
		if validDomainSegment(s) {
			t.Errorf("expected invalid segment: %q", s)
		}
	}
}

func TestAddDomainInNamespace(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := AddDomainInNamespace(p, "cloud", "newdomain", "Owner Team", false)
	if err != nil {
		t.Fatalf("AddDomainInNamespace: %v", err)
	}
	if dto.Name != "newdomain.cloud" {
		t.Errorf("unexpected name: %q", dto.Name)
	}
}

func TestAddDomainInNamespace_InvalidLabel(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddDomainInNamespace(p, "cloud", "Invalid Label!", "Owner", false)
	if err == nil {
		t.Error("expected error for invalid label")
	}
}

func TestAddChildDomain(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := AddChildDomain(p, "identity.blumer.cloud", "sub", "Owner", false)
	if err != nil {
		t.Fatalf("AddChildDomain: %v", err)
	}
	if dto.Name != "sub.identity.blumer.cloud" {
		t.Errorf("unexpected name: %q", dto.Name)
	}
}

func TestAddChildDomain_InvalidSegment(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := AddChildDomain(p, "identity.blumer.cloud", "Invalid!", "Owner", false)
	if err == nil {
		t.Error("expected error for invalid segment")
	}
}
