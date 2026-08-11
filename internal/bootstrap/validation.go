package bootstrap

import (
	"crypto/sha256"
	"fmt"
	"path"
	"sort"
	"strings"
)

func (m Manifest) normalized() (Manifest, error) {
	if m.Schema == "" {
		m.Schema = SchemaVersion
	}
	if m.Schema != SchemaVersion {
		return Manifest{}, fmt.Errorf("unsupported manifest schema %q", m.Schema)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"stage", m.Stage},
		{"source digest", m.SourceDigest},
		{"compiler digest", m.CompilerDigest},
		{"semantic digest", m.SemanticDigest},
	} {
		if strings.TrimSpace(field.value) == "" {
			return Manifest{}, fmt.Errorf("manifest %s is required", field.name)
		}
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"source digest", m.SourceDigest},
		{"compiler digest", m.CompilerDigest},
		{"semantic digest", m.SemanticDigest},
	} {
		if !validDigest(field.value) {
			return Manifest{}, fmt.Errorf("manifest %s is not a lowercase SHA-256 digest", field.name)
		}
	}
	var err error
	m.Inputs, err = normalizeArtifacts(m.Inputs)
	if err != nil {
		return Manifest{}, fmt.Errorf("inputs: %w", err)
	}
	m.Outputs, err = normalizeArtifacts(m.Outputs)
	if err != nil {
		return Manifest{}, fmt.Errorf("outputs: %w", err)
	}
	return m, nil
}

func normalizeArtifacts(artifacts []Artifact) ([]Artifact, error) {
	result := append([]Artifact{}, artifacts...)
	for _, artifact := range result {
		if err := validatePath(artifact.Path); err != nil {
			return nil, err
		}
		if !validDigest(artifact.SHA256) {
			return nil, fmt.Errorf("artifact %q has an invalid SHA-256 digest", artifact.Path)
		}
		if artifact.Size < 0 {
			return nil, fmt.Errorf("artifact %q has a negative size", artifact.Path)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	for index := 1; index < len(result); index++ {
		if result[index-1].Path == result[index].Path {
			return nil, fmt.Errorf("duplicate artifact path %q", result[index].Path)
		}
	}
	return result, nil
}

func validateEvidence(evidence Evidence) error {
	if evidence.Sequence == 0 || strings.TrimSpace(evidence.Kind) == "" || strings.TrimSpace(evidence.Subject) == "" {
		return fmt.Errorf("evidence sequence, kind, and subject are required")
	}
	if !validDigest(evidence.ClaimDigest) {
		return fmt.Errorf("evidence claim digest is not a lowercase SHA-256 digest")
	}
	if evidence.PreviousDigest != "" && !validDigest(evidence.PreviousDigest) {
		return fmt.Errorf("evidence predecessor is not a lowercase SHA-256 digest")
	}
	return nil
}

func validateEvidenceChain(evidence []Evidence) error {
	previous := ""
	for index, record := range evidence {
		if record.Sequence != uint64(index+1) {
			return fmt.Errorf("evidence sequence at index %d is invalid", index)
		}
		if err := validateEvidence(record); err != nil {
			return err
		}
		if record.PreviousDigest != previous {
			return fmt.Errorf("evidence predecessor at index %d is invalid", index)
		}
		var err error
		previous, err = record.Digest()
		if err != nil {
			return err
		}
	}
	return nil
}

func validatePath(value string) error {
	if value == "" || strings.ContainsAny(value, "\\\x00") || path.IsAbs(value) || path.Clean(value) != value || value == "." || strings.HasPrefix(value, "../") || value == ".." {
		return fmt.Errorf("artifact path %q is not a normalized relative path", value)
	}
	return nil
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
