package dmn

import (
	"bytes"
	"regexp"
	"strings"
)

// NormalizeInputIdentifiers keeps each <inputData>'s descriptive pill label
// (the name attribute) and its machine-readable identifier in sync.
//
// DMN editors expose two distinct identifiers per InputData: the displayed
// pill name (free-form, may contain spaces and umlauts) and the FEEL-callable
// variable name used by decision tables. Modellers routinely type into the
// pill and forget that the variable plus every decision-table input column
// also has to follow. We derive variable.name = snakeCase(inputData.name),
// rewrite any decision-table <inputExpression> that referenced the previous
// variant, and leave the pill itself untouched so non-technical readers keep
// their human label.
//
// The rewrite is surgical: only the specific variable.name attributes and
// inputExpression text contents change, so dmn-js's namespace prefixes,
// whitespace, and DRD layout survive a round-trip.
func NormalizeInputIdentifiers(xmlData []byte) ([]byte, error) {
	defs, err := ParseDefinitions(xmlData)
	if err != nil {
		return xmlData, err
	}
	rename := map[string]string{}
	for _, in := range defs.InputData {
		pill := strings.TrimSpace(in.Name)
		if pill == "" {
			continue
		}
		newVar := SnakeCase(pill)
		if newVar == "" {
			continue
		}
		oldVar := strings.TrimSpace(in.Variable.Name)
		if oldVar != newVar {
			xmlData = rewriteInputDataVariableName(xmlData, in.ID, newVar)
		}
		// Record every prior alias a decision-table column might be
		// referencing so we can rewrite them to the canonical machine name.
		if oldVar != "" && oldVar != newVar {
			rename[oldVar] = newVar
		}
		if pill != newVar {
			rename[pill] = newVar
		}
	}
	if len(rename) > 0 {
		xmlData = rewriteInputExpressionTexts(xmlData, rename)
	}
	return xmlData, nil
}

// SnakeCase converts a free-form descriptive label to a lowercase snake_case
// identifier safe for use as a DMN variable name. Spaces and other special
// characters collapse to a single underscore; German umlauts are
// transliterated (ä→ae, ö→oe, ü→ue, ß→ss); the result is trimmed of leading
// and trailing underscores. Empty / digit-only input yields an empty string,
// signalling to the caller that no rename is possible.
func SnakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = umlautReplacer.Replace(s)
	var b strings.Builder
	b.Grow(len(s))
	lastUnderscore := true
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
			lastUnderscore = false
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

var umlautReplacer = strings.NewReplacer(
	"Ä", "Ae", "ä", "ae",
	"Ö", "Oe", "ö", "oe",
	"Ü", "Ue", "ü", "ue",
	"ß", "ss",
)

var variableNameAttrRe = regexp.MustCompile(`(<variable\b[^>]*\bname=")[^"]*(")`)

func rewriteInputDataVariableName(xml []byte, inputDataID, newVar string) []byte {
	if inputDataID == "" {
		return xml
	}
	blockRe := regexp.MustCompile(
		`(?s)<inputData\b[^>]*\bid="` + regexp.QuoteMeta(inputDataID) + `"[^>]*>.*?</inputData>`,
	)
	return blockRe.ReplaceAllFunc(xml, func(block []byte) []byte {
		return variableNameAttrRe.ReplaceAll(block, []byte("${1}"+newVar+"${2}"))
	})
}

// inputExpressionTextRe matches a decision-table input column's FEEL
// expression body — <inputExpression ...> ... <text>X</text> ... </inputExpression>.
// We keep the prefix and suffix verbatim so any attributes, whitespace,
// or sibling elements (e.g. <inputValues>) survive untouched.
var inputExpressionTextRe = regexp.MustCompile(
	`(?s)(<inputExpression\b[^>]*>\s*<text>)\s*(.*?)\s*(</text>)`,
)

func rewriteInputExpressionTexts(xml []byte, rename map[string]string) []byte {
	return inputExpressionTextRe.ReplaceAllFunc(xml, func(m []byte) []byte {
		groups := inputExpressionTextRe.FindSubmatch(m)
		if groups == nil {
			return m
		}
		current := string(groups[2])
		newVal, ok := rename[current]
		if !ok {
			return m
		}
		var buf bytes.Buffer
		buf.Grow(len(groups[1]) + len(newVal) + len(groups[3]))
		buf.Write(groups[1])
		buf.WriteString(newVal)
		buf.Write(groups[3])
		return buf.Bytes()
	})
}
