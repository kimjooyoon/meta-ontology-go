package selectiveci

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type issueClass uint8

const (
	issueNone issueClass = iota
	issueUnknown
	issueFailClosed
)

type issueState struct {
	class issueClass
	code  string
}

func (s *issueState) add(class issueClass, code string) {
	if class > s.class || class == s.class && code < s.code {
		s.class, s.code = class, code
	}
}
func normalizeSnapshot(raw semantic.SnapshotDigests, label string, state *issueState) semantic.SnapshotDigests {
	if raw.Source == "" || raw.Semantic == "" {
		state.add(issueUnknown, CodeMissing)
		return raw
	}
	source, sourceErr := normalizeDigest(raw.Source, label+" source snapshot")
	semanticDigest, semanticErr := normalizeDigest(raw.Semantic, label+" semantic snapshot")
	if sourceErr != nil || semanticErr != nil {
		state.add(issueFailClosed, CodeDigestMismatch)
	}
	return semantic.SnapshotDigests{Source: source, Semantic: semanticDigest}
}
func normalizeSequence(values []semantic.ID, label string) ([]semantic.ID, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("%s is empty", label)
	}
	out := make([]semantic.ID, len(values))
	seen := make(map[semantic.ID]struct{}, len(values))
	for i, value := range values {
		id, err := normalizeID(value, label)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate %s %s", label, id)
		}
		seen[id] = struct{}{}
		out[i] = id
	}
	return out, nil
}
