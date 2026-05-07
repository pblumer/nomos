package app

import (
	"net/http"
	"strings"

	"github.com/nomos/nomos/internal/namespace"
)

func AddDomainInNamespace(path, namespaceName, label, owner string, force bool) (DomainDTO, error) {
	canonical, err := namespace.ComposeCanonical(namespaceName, label)
	if err != nil {
		return DomainDTO{}, Error(CodeInvalidNamespace, err.Error(), http.StatusBadRequest, err)
	}
	return AddDomain(path, canonical, owner, force)
}

func AddChildDomain(path, parentCanonicalName, segment, owner string, force bool) (DomainDTO, error) {
	parentCanonicalName = namespace.Canonical(parentCanonicalName)
	canonical, err := namespace.ComposeChildCanonical(parentCanonicalName, segment)
	if err != nil {
		return DomainDTO{}, Error(CodeInvalidNamespace, "Invalid segment. Use only one new label, for example: identity", http.StatusBadRequest, err)
	}
	if strings.Contains(parentCanonicalName, ".") {
		if _, err := GetDomain(path, parentCanonicalName); err != nil {
			if !isAppCode(err, CodeDomainNotFound) {
				return DomainDTO{}, err
			}
			if _, materializeErr := addDomain(path, parentCanonicalName, owner, false, true); materializeErr != nil && !isAppCode(materializeErr, CodeInvalidNamespace) {
				return DomainDTO{}, materializeErr
			} else if materializeErr != nil {
				return DomainDTO{}, Error(CodeDomainNotFound, "Cannot create child domain because "+parentCanonicalName+" is only a virtual parent. Materialize "+parentCanonicalName+" first.", http.StatusConflict, materializeErr)
			}
		}
	}
	return AddDomain(path, canonical, owner, force)
}

func isAppCode(err error, code string) bool {
	return err != nil && strings.Contains(err.Error(), code)
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
