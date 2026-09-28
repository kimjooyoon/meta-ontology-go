package domaincapability

import (
	"sort"
	"strings"
)

// CapabilityBinding ties a capability to its declaration and both evidence layers.
type CapabilityBinding struct {
	CapabilityID  string
	DeclarationID string
	SourceDigest  string
	EvidenceDigest string
}

// NormalizeBindings keeps only fully bound observations in deterministic order.
func NormalizeBindings(bindings []CapabilityBinding) []CapabilityBinding {
	seen := make(map[string]struct{}, len(bindings))
	out := make([]CapabilityBinding, 0, len(bindings))
	for _, binding := range bindings {
		binding.CapabilityID = strings.TrimSpace(binding.CapabilityID)
		binding.DeclarationID = strings.TrimSpace(binding.DeclarationID)
		binding.SourceDigest = strings.TrimSpace(binding.SourceDigest)
		binding.EvidenceDigest = strings.TrimSpace(binding.EvidenceDigest)
		if binding.CapabilityID == "" || binding.DeclarationID == "" || binding.SourceDigest == "" || binding.EvidenceDigest == "" {
			continue
		}
		key := binding.CapabilityID + "\x00" + binding.DeclarationID + "\x00" + binding.SourceDigest + "\x00" + binding.EvidenceDigest
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, binding)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CapabilityID != out[j].CapabilityID {
			return out[i].CapabilityID < out[j].CapabilityID
		}
		return out[i].DeclarationID < out[j].DeclarationID
	})
	return out
}

// SourceCapabilityIDs returns IDs that have an explicit declaration and evidence binding.
func SourceCapabilityIDs(bindings []CapabilityBinding) []string {
	bindings = NormalizeBindings(bindings)
	seen := make(map[string]struct{}, len(bindings))
	ids := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if _, ok := seen[binding.CapabilityID]; ok {
			continue
		}
		seen[binding.CapabilityID] = struct{}{}
		ids = append(ids, binding.CapabilityID)
	}
	return ids
}