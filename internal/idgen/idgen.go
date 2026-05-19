// Package idgen erzeugt und validiert Artefakt-IDs gemäß ADR-0020.
//
// Format: <TYP>_<6 Zeichen Crockford-Base32>, z. B. "PRD_A7K3M2".
// Crockford-Base32 lässt I, L, O, U aus, damit IDs robust beim Abtippen sind.
//
// Sonderfall: "COS_root" ist die fest vergebene ID des Cosmos (genau einer pro Repo).
package idgen

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
)

// crockfordAlphabet folgt der Crockford-Base32-Definition (ohne I, L, O, U).
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// SuffixLength ist die Anzahl der zufälligen Zeichen nach dem Präfix.
const SuffixLength = 6

// CosmosRootID ist die feste ID des Cosmos-Wurzelartefakts.
const CosmosRootID = "COS_root"

// artefactPrefixes ordnet jedem Artefakttyp einen festen 3-Buchstaben-Präfix zu.
// Neue Artefakttypen müssen hier registriert werden.
var artefactPrefixes = map[string]string{
	"cosmos":              "COS",
	"domain":              "DOM",
	"service":             "SRV", // Namespace-Service (live)
	"product":             "PRD", // legacy product artefact, gleicher Präfix wie product_blueprint
	"product_blueprint":   "PRD",
	"service_blueprint":   "SVC",
	"product_instance":    "PRI",
	"service_instance":    "SVI",
	"process":             "PRC",
	"process_step":        "STP",
	"decision":            "DEC",
	"evidence":            "EVD",
	"finding":             "FND",
	"servicegraph":        "SGR",
	"requirement":         "REQ",
	"rule":                "RUL",
	"validation_scenario": "VSC",
	"variant":             "PRV",
	"attribute":           "ATR",
	"blueprint":           "BLP",
}

var (
	idRegex     = regexp.MustCompile(`^[A-Z]{3}_[0-9A-HJKMNP-TV-Z]{6}$`)
	prefixRegex = regexp.MustCompile(`^[A-Z]{3}$`)
)

// PrefixForType liefert den kanonischen Präfix für einen Artefakttyp.
func PrefixForType(artefactType string) (string, bool) {
	p, ok := artefactPrefixes[strings.TrimSpace(artefactType)]
	return p, ok
}

// KnownTypes liefert alle registrierten Artefakttypen.
func KnownTypes() []string {
	out := make([]string, 0, len(artefactPrefixes))
	for k := range artefactPrefixes {
		out = append(out, k)
	}
	return out
}

// New erzeugt eine frische ID für einen 3-Buchstaben-Präfix.
func New(prefix string) (string, error) {
	p := strings.ToUpper(strings.TrimSpace(prefix))
	if !prefixRegex.MatchString(p) {
		return "", fmt.Errorf("idgen: invalid prefix %q (must be exactly 3 uppercase letters)", prefix)
	}
	suffix, err := randomCrockford(SuffixLength)
	if err != nil {
		return "", err
	}
	return p + "_" + suffix, nil
}

// NewForType erzeugt eine frische ID anhand des Artefakttyps.
func NewForType(artefactType string) (string, error) {
	if artefactType == "cosmos" {
		// Cosmos hat genau eine feste ID pro Repo.
		return CosmosRootID, nil
	}
	p, ok := PrefixForType(artefactType)
	if !ok {
		return "", fmt.Errorf("idgen: no prefix registered for artefact type %q", artefactType)
	}
	return New(p)
}

// IsValid prüft, ob id dem ADR-0020-Format entspricht.
func IsValid(id string) bool {
	if id == CosmosRootID {
		return true
	}
	return idRegex.MatchString(id)
}

// IsValidForType prüft Format und Präfix-Typ-Zuordnung.
// Unbekannte Typen passieren die Präfix-Prüfung; nur das Format wird geprüft.
func IsValidForType(id, artefactType string) bool {
	if !IsValid(id) {
		return false
	}
	if id == CosmosRootID {
		return artefactType == "cosmos"
	}
	p, ok := PrefixForType(artefactType)
	if !ok {
		return true
	}
	return strings.HasPrefix(id, p+"_")
}

// IsLegacy gibt true zurück, wenn id nicht leer ist und nicht dem
// neuen Format entspricht. Wird von Migration und Validator genutzt.
func IsLegacy(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	return !IsValid(id)
}

// randomCrockford erzeugt n Crockford-Base32-Zeichen kryptografisch zufällig.
// 256 ist ein exaktes Vielfaches von 32, daher ist die Modulo-Bildung bias-frei.
func randomCrockford(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, v := range buf {
		out[i] = crockfordAlphabet[int(v)%len(crockfordAlphabet)]
	}
	return string(out), nil
}
