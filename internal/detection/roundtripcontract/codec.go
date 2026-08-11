package roundtripcontract

import "encoding/json"

// Normalize returns a detached evidence value with deterministic finding and
// artifact ordering. It does not alter semantic content.
func (e Evidence) Normalize() Evidence {
	result := e
	result.Artifacts = append([]ArtifactRef(nil), e.Artifacts...)
	result.Findings = normalizedFindings(e.Findings)
	sortArtifacts(result.Artifacts)
	return result
}

// CanonicalJSON returns stable, newline-terminated evidence for CI, cache, or
// provenance storage. It validates before serialization so invalid evidence
// cannot masquerade as a durable record.
func CanonicalJSON(e Evidence) ([]byte, error) {
	e = e.Normalize()
	if err := e.Validate(); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func sortArtifacts(artifacts []ArtifactRef) {
	for left := 0; left < len(artifacts); left++ {
		for right := left + 1; right < len(artifacts); right++ {
			if artifactLess(artifacts[right], artifacts[left]) {
				artifacts[left], artifacts[right] = artifacts[right], artifacts[left]
			}
		}
	}
}

func artifactLess(left, right ArtifactRef) bool {
	if left.Stage != right.Stage {
		return left.Stage < right.Stage
	}
	if left.URI != right.URI {
		return left.URI < right.URI
	}
	if left.Format != right.Format {
		return left.Format < right.Format
	}
	return left.Digest < right.Digest
}
