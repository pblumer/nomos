package dmn

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/nomos/nomos/internal/model"
)

// HashSHA256 returns the hex-encoded sha256 of the input bytes, prefixed with "sha256:".
func HashSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ComputeTraceID returns the content hash of a DecisionTrace.
// The TraceID field on the input is ignored (treated as empty) during hashing,
// so callers can compute and then assign it.
func ComputeTraceID(t model.DecisionTrace) (string, error) {
	t.TraceID = ""
	buf, err := canonicalJSON(t)
	if err != nil {
		return "", err
	}
	return HashSHA256(buf), nil
}

// canonicalJSON marshals v to JSON with sorted map keys at every level.
// Go's json.Marshal already sorts map[string]X keys, but values of type
// map[string]any nested inside arrays/structs are handled by re-encoding
// after a normalization pass via map[string]any.
func canonicalJSON(v any) ([]byte, error) {
	// First pass: encode to JSON to normalize concrete types.
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// Second pass: decode into an interface{} and re-marshal with sorted keys.
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := encodeCanonical(&buf, generic); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeCanonical(buf *bytes.Buffer, v any) error {
	switch x := v.(type) {
	case nil:
		buf.WriteString("null")
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return err
			}
			buf.Write(kb)
			buf.WriteByte(':')
			if err := encodeCanonical(buf, x[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case []any:
		buf.WriteByte('[')
		for i, item := range x {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := encodeCanonical(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return err
		}
		buf.Write(b)
	}
	return nil
}

// TraceBuildInput bundles everything required to build a DecisionTrace from an evaluation.
type TraceBuildInput struct {
	Domain          string
	DecisionID      string
	DecisionName    string
	DecisionVersion string
	RuleBytes       []byte
	Inputs          map[string]any
	Result          *Result
	Evaluator       model.Evaluator
	Engine          model.EngineInfo
	Timestamp       string // RFC3339Nano, supplied by caller for determinism in tests
	ParentTraceID   string
}

// BuildTrace assembles a DecisionTrace from an evaluation and computes its TraceID.
func BuildTrace(in TraceBuildInput) (model.DecisionTrace, error) {
	if in.Result == nil {
		return model.DecisionTrace{}, fmt.Errorf("BuildTrace: nil Result")
	}
	t := model.DecisionTrace{
		SchemaVersion:   1,
		ParentTraceID:   in.ParentTraceID,
		Timestamp:       in.Timestamp,
		Domain:          in.Domain,
		DecisionID:      in.DecisionID,
		DecisionName:    in.DecisionName,
		DecisionVersion: in.DecisionVersion,
		RuleHash:        HashSHA256(in.RuleBytes),
		Inputs:          in.Inputs,
		Outputs:         in.Result.Outputs,
		MatchedRules:    in.Result.MatchedRules,
		HitPolicy:       in.Result.HitPolicy,
		Evaluator:       in.Evaluator,
		Engine:          in.Engine,
	}
	id, err := ComputeTraceID(t)
	if err != nil {
		return model.DecisionTrace{}, err
	}
	t.TraceID = id
	return t, nil
}

// VerifyTrace recomputes the TraceID and returns nil if it matches the stored value.
func VerifyTrace(t model.DecisionTrace) error {
	stored := t.TraceID
	if stored == "" {
		return fmt.Errorf("trace has no trace_id")
	}
	recomputed, err := ComputeTraceID(t)
	if err != nil {
		return err
	}
	if recomputed != stored {
		return fmt.Errorf("trace_id mismatch: stored=%s recomputed=%s", stored, recomputed)
	}
	return nil
}
