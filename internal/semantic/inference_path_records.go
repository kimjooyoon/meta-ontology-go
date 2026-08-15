package semantic

import (
	"errors"
	"fmt"
	"sort"
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

func (c InferenceControls) normalized() (InferenceControls, error) {
	out := InferenceControls{
		CatalogDigest: strings.TrimSpace(c.CatalogDigest), PolicyDigest: strings.TrimSpace(c.PolicyDigest),
	}
	if out.CatalogDigest != "" {
		digest, err := normalizeDigest(out.CatalogDigest)
		if err != nil {
			return InferenceControls{}, fmt.Errorf("catalog digest: %w", err)
		}
		out.CatalogDigest = digest
	}
	if out.PolicyDigest != "" {
		digest, err := normalizeDigest(out.PolicyDigest)
		if err != nil {
			return InferenceControls{}, fmt.Errorf("policy digest: %w", err)
		}
		out.PolicyDigest = digest
	}
	profile, err := c.Profile.normalized()
	if err != nil {
		return InferenceControls{}, err
	}
	out.Profile = profile
	return out, nil
}

func (r EvidenceReference) normalized() (EvidenceReference, error) {
	id, err := ParseIdentity(r.ID.String())
	if err != nil {
		return EvidenceReference{}, fmt.Errorf("evidence ID: %w", err)
	}
	digest, err := normalizeDigest(r.Digest)
	if err != nil {
		return EvidenceReference{}, fmt.Errorf("evidence digest: %w", err)
	}
	return EvidenceReference{ID: id, Digest: digest}, nil
}

func normalizeEvidenceReferences(raw []EvidenceReference) ([]EvidenceReference, error) {
	if len(raw) == 0 {
		return nil, errors.New("at least one evidence reference is required")
	}
	out := make([]EvidenceReference, 0, len(raw))
	seen := make(map[ID]struct{}, len(raw))
	for _, ref := range raw {
		normalized, err := ref.normalized()
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized.ID]; exists {
			return nil, fmt.Errorf("duplicate evidence reference %s", normalized.ID)
		}
		seen[normalized.ID] = struct{}{}
		out = append(out, normalized)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r InferenceRecord) normalized() (InferenceRecord, error) {
	var out InferenceRecord
	var err error
	if out.RecordID, err = ParseIdentity(r.RecordID.String()); err != nil {
		return InferenceRecord{}, fmt.Errorf("record ID: %w", err)
	}
	if out.SubjectID, err = ParseIdentity(r.SubjectID.String()); err != nil {
		return InferenceRecord{}, fmt.Errorf("subject ID: %w", err)
	}
	if out.ObjectID, err = ParseIdentity(r.ObjectID.String()); err != nil {
		return InferenceRecord{}, fmt.Errorf("object ID: %w", err)
	}
	if out.Rule, err = r.Rule.normalized(); err != nil {
		return InferenceRecord{}, err
	}
	out.Phase = r.Phase
	if !out.Phase.Phase.Valid() || out.Phase.Ordinal == 0 {
		return InferenceRecord{}, errors.New("phase and non-zero phase ordinal are required")
	}
	if out.Before, err = r.Before.normalized(); err != nil {
		return InferenceRecord{}, fmt.Errorf("before: %w", err)
	}
	if out.After, err = r.After.normalized(); err != nil {
		return InferenceRecord{}, fmt.Errorf("after: %w", err)
	}
	out.Authority = r.Authority
	if !out.Authority.Layer.Valid() || !out.Authority.Effect.Valid() {
		return InferenceRecord{}, errors.New("unknown authority layer or effect")
	}
	if out.Evidence, err = normalizeEvidenceReferences(r.Evidence); err != nil {
		return InferenceRecord{}, err
	}
	if out.Controls, err = r.Controls.normalized(); err != nil {
		return InferenceRecord{}, err
	}
	return out, nil
}

func (e InferenceEvidence) normalized() (InferenceEvidence, error) {
	id, err := ParseIdentity(e.ID.String())
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence record ID: %w", err)
	}
	digest, err := normalizeDigest(e.Digest)
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence record digest: %w", err)
	}
	before, err := e.Before.normalized()
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence before: %w", err)
	}
	after, err := e.After.normalized()
	if err != nil {
		return InferenceEvidence{}, fmt.Errorf("evidence after: %w", err)
	}
	controls, err := e.Controls.normalized()
	if err != nil {
		return InferenceEvidence{}, err
	}
	return InferenceEvidence{
		ID: id, Digest: digest, Before: before, After: after,
		SourceBacked: e.SourceBacked, Independent: e.Independent, Controls: controls,
	}, nil
}

func (e InferenceEvidence) Normalized() (InferenceEvidence, error) { return e.normalized() }

func (e InferenceEvidence) Validate() error { _, err := e.normalized(); return err }
