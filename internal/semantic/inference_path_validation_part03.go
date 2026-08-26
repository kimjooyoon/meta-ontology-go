package semantic

func validateInferenceEvidence(
	binding InferenceRecord, receipt ID, requireReceipt bool,
	records map[ID]InferenceEvidence, issues *InferencePathErrors,
) {
	refs := make(map[ID]struct{}, len(binding.Evidence))
	for _, ref := range binding.Evidence {
		if _, duplicate := refs[ref.ID]; duplicate {
			issues.add("duplicate-evidence", binding.RecordID, ref.ID.String())
			continue
		}
		refs[ref.ID] = struct{}{}
		record, ok := records[ref.ID]
		if !ok {
			issues.add("orphan-evidence", binding.RecordID, ref.ID.String())
			continue
		}
		if record.Digest != ref.Digest || record.Before != binding.Before ||
			record.After != binding.After || record.Controls != binding.Controls {
			issues.add("stale-evidence", binding.RecordID, ref.ID.String())
		}
	}
	if requireReceipt {
		if receipt == "" {
			issues.add("missing-acceptance-receipt", binding.RecordID, "empty receipt")
		} else if _, ok := refs[receipt]; !ok {
			issues.add("orphan-acceptance-receipt", binding.RecordID, receipt.String())
		} else if !records[receipt].SourceBacked || records[receipt].Before.Source == "" ||
			records[receipt].After.Source == "" {
			issues.add("unbacked-acceptance-receipt", binding.RecordID, receipt.String())
		}
	}
}
func hasIndependentEvidence(refs []EvidenceReference, records map[ID]InferenceEvidence) bool {
	for _, ref := range refs {
		if records[ref.ID].Independent {
			return true
		}
	}
	return false
}
func (p InferencePathV1) Validate() error    { _, err := p.Normalized(); return err }
func (p InferencePathV1) StableHash() string { return StableHashString(p.Canonical()) }

// InferencePathChain proves one finite, unambiguous chain over typed edges.
type InferencePathChain struct{ Edges []InferenceEdge }
