package claimledger

import "strings"

func selectEvidence(spec ClaimSpec, sources map[string]sourceState, evidenceID string) (Evidence, any, bool, string) {
	source, exists := sources[spec.Evidence.Source]
	evidence := Evidence{
		ID: evidenceID, ClaimID: spec.ID, Source: spec.Evidence.Source,
		SourcePath: strings.Join(spec.Evidence.Paths, "|"), SourceDigest: source.Digest,
	}
	if !exists || source.Status == "MISSING" {
		evidence.Status = "MISSING"
		return evidence, nil, false, spec.UnknownReason
	}
	if source.Status != "VERIFIED" {
		evidence.Status = "REJECTED"
		return evidence, nil, false, source.Reason
	}
	path, observed, found := lookupAny(source.Value, spec.Evidence.Paths)
	if !found {
		evidence.Status = "MISSING"
		return evidence, nil, false, spec.UnknownReason
	}
	evidence.SourcePath = spec.Evidence.Source + ":" + path
	return evidence, observed, true, ""
}
