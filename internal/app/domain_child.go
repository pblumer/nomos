package app

import (
	"net/http"
	"strings"

	"github.com/nomos/nomos/internal/namespace"
)

func AddChildDomain(path, parentCanonicalName, segment, owner string, force bool) (DomainDTO, error) {
	parent := namespace.Canonical(strings.TrimSpace(parentCanonicalName))
	segment = strings.TrimSpace(segment)
	if parent == "" || strings.ContainsAny(parent, `/\\`) || strings.Contains(parent, " ") || strings.HasPrefix(parent, ".") || strings.HasSuffix(parent, ".") {
		return DomainDTO{}, Error(CodeInvalidNamespace, "Invalid parent domain: "+parent, http.StatusBadRequest, nil)
	}
	if !validDomainSegment(segment) {
		return DomainDTO{}, Error(CodeInvalidNamespace, "Invalid segment. Use only the new segment, for example: test2", http.StatusBadRequest, nil)
	}
	return AddDomain(path, segment+"."+parent, owner, force)
}

func validDomainSegment(segment string) bool {
	if segment == "" || strings.ContainsAny(segment, `. /\\`) || strings.HasPrefix(segment, "-") || strings.HasSuffix(segment, "-") {
		return false
	}
	for _, r := range segment {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}
