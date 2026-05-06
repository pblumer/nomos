package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestDomainsPageRendersContextualCreateUI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/domains?selected=domain:identity.blumer.cloud")
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	hasAll(t, rr.Body.String(), "Add child domain", "Add service", "Create child domain", "Parent domain", "New segment", "Resulting canonical name", "identity.blumer.cloud", "Hint: Enter only the new segment")
	if strings.Contains(rr.Body.String(), "DNS/canonical name") {
		t.Fatalf("global technical create form should not be primary")
	}
}

func TestChildDomainFormSubmissionComposesCanonicalName(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)
	rr := postForm(h, "/domains/create-child?selected=domain:identity.blumer.cloud", "parent=identity.blumer.cloud&segment=iam&owner=Web")
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/domains?selected=domain:iam.identity.blumer.cloud" {
		t.Fatalf("redirect=%q", got)
	}
	if rr := get(h, "/api/v1/domains/iam.identity.blumer.cloud"); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "iam.identity.blumer.cloud") {
		t.Fatalf("created domain missing: %d %s", rr.Code, rr.Body.String())
	}
}

func TestChildDomainValidationKeepsInput(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postForm(h, "/domains/create-child?selected=domain:identity.blumer.cloud", "parent=identity.blumer.cloud&segment=iam.identity.blumer.cloud&owner=Web")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "Could not create child domain", "Invalid segment", "iam.identity.blumer.cloud", "Do not enter the full domain name here")
}

func TestTopLevelAdvancedAndContextualServiceCreation(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)
	if rr := postForm(h, "/domains/create-top-level?selected=cosmos", "canonical=example.com&owner=Web"); rr.Code != http.StatusSeeOther {
		t.Fatalf("top-level status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/domains/create-advanced?selected=cosmos", "canonical=advanced.example.com&owner=Web&mode=advanced"); rr.Code != http.StatusSeeOther {
		t.Fatalf("advanced status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/services/create?selected=domain:example.com", "domain=example.com&name=identity&owner=Web"); rr.Code != http.StatusSeeOther {
		t.Fatalf("service status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := get(h, "/api/v1/domains/example.com/services/identity"); rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "identity") {
		t.Fatalf("service missing: %d %s", rr.Code, rr.Body.String())
	}
}
