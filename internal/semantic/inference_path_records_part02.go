package semantic

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

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
