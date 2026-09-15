package semantic

import (
	"errors"
	"fmt"
	"strings"
)

func (r RuleBinding) normalized() (RuleBinding, error) {
	id, err := ParseIdentity(r.ID.String())
	if err != nil {
		return RuleBinding{}, fmt.Errorf("rule ID: %w", err)
	}
	version := strings.TrimSpace(r.Version)
	if version == "" {
		return RuleBinding{}, errors.New("rule version is required")
	}
	digest, err := normalizeDigest(r.Digest)
	if err != nil {
		return RuleBinding{}, fmt.Errorf("rule digest: %w", err)
	}
	return RuleBinding{ID: id, Version: version, Digest: digest}, nil
}
func (s SnapshotDigests) normalized() (SnapshotDigests, error) {
	out := SnapshotDigests{Source: strings.TrimSpace(s.Source), Semantic: strings.TrimSpace(s.Semantic)}
	if out.Source == "" && out.Semantic == "" {
		return SnapshotDigests{}, errors.New("source or semantic snapshot digest is required")
	}
	if out.Source != "" {
		digest, err := normalizeDigest(out.Source)
		if err != nil {
			return SnapshotDigests{}, fmt.Errorf("source snapshot: %w", err)
		}
		out.Source = digest
	}
	if out.Semantic != "" {
		digest, err := normalizeDigest(out.Semantic)
		if err != nil {
			return SnapshotDigests{}, fmt.Errorf("semantic snapshot: %w", err)
		}
		out.Semantic = digest
	}
	return out, nil
}
func (p ProfileBinding) normalized() (ProfileBinding, error) {
	out := ProfileBinding{
		ID: strings.TrimSpace(p.ID), Version: strings.TrimSpace(p.Version), Digest: strings.TrimSpace(p.Digest),
	}
	if out.ID == "" && out.Version == "" && out.Digest == "" {
		return ProfileBinding{}, nil
	}
	if out.ID == "" || out.Version == "" || out.Digest == "" {
		return ProfileBinding{}, errors.New("profile ID, version, and digest are all required")
	}
	digest, err := normalizeDigest(out.Digest)
	if err != nil {
		return ProfileBinding{}, fmt.Errorf("profile digest: %w", err)
	}
	out.Digest = digest
	return out, nil
}
