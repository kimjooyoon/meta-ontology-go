package metacircularboundary

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func parseGrantEvidence(raw []byte, semanticDigest string) (map[string]ExternalGrant, string, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, "", fmt.Errorf("external grant artifact is absent")
	}
	var artifact ExternalGrantArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		return nil, "", fmt.Errorf("decode external grant artifact: %w", err)
	}
	digest := digestBytes(raw)
	if artifact.Schema != GrantSchema || artifact.Producer != GrantProducer || artifact.Policy != "READ_ONLY_EXTERNAL_GRANT_POLICY_V1" {
		return nil, digest, fmt.Errorf("external grant artifact identity is invalid")
	}
	if artifact.ArtifactDigest != "" && artifact.ArtifactDigest != digest {
		return nil, digest, fmt.Errorf("external grant artifact digest mismatch")
	}
	grants := make(map[string]ExternalGrant, len(artifact.Grants))
	for _, grant := range artifact.Grants {
		if grant.CaseID == "" || grant.GrantDigest != "" {
			return nil, digest, fmt.Errorf("external grant has invalid identity")
		}
		if _, exists := grants[grant.CaseID]; exists {
			return nil, digest, fmt.Errorf("duplicate external grant %q", grant.CaseID)
		}
		if grant.Decision != GrantDecision && grant.Decision != GrantDeny {
			return nil, digest, fmt.Errorf("external grant %q has unknown decision", grant.CaseID)
		}
		if grant.Decision == GrantDecision && (grant.Issuer == "" || grant.SubjectDigest == "" || grant.Operation == "" || grant.Scope == "" || grant.Handle == "") {
			return nil, digest, fmt.Errorf("external grant %q is incomplete", grant.CaseID)
		}
		grant.GrantDigest = digestValue(grant)
		grants[grant.CaseID] = grant
	}
	for _, definition := range contractCases() {
		if _, ok := grants[definition.ID]; !ok {
			return nil, digest, fmt.Errorf("external grant %q is absent", definition.ID)
		}
	}
	return grants, digest, nil
}

func parseEffectEvidence(raw []byte) (EffectObservation, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return EffectObservation{}, fmt.Errorf("workspace effect artifact is absent")
	}
	var evidence EffectEvidence
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		return EffectObservation{}, fmt.Errorf("decode workspace effect artifact: %w", err)
	}
	if evidence.Schema != EffectSchema || evidence.Producer != EffectProducer || evidence.OutputPath == "" || evidence.PermissionEvidence != "workflow-contents-read-only" || evidence.MutationAuthority == "" {
		return EffectObservation{}, fmt.Errorf("workspace effect artifact identity is invalid")
	}
	if !validDigest(evidence.TrackedBeforeDigest) || !validDigest(evidence.TrackedAfterDigest) || !validDigest(evidence.UntrackedBeforeDigest) || !validDigest(evidence.UntrackedAfterDigest) {
		return EffectObservation{}, fmt.Errorf("workspace effect snapshot digest is invalid")
	}
	if evidence.MutationAuthority != AuthorityDenied && evidence.MutationAuthority != AuthorityGranted && evidence.MutationAuthority != AuthorityUnknown {
		return EffectObservation{}, fmt.Errorf("workspace mutation authority is invalid")
	}
	writes := 0
	if evidence.TrackedBeforeDigest != evidence.TrackedAfterDigest || evidence.UntrackedBeforeDigest != evidence.UntrackedAfterDigest {
		writes = 1
	}
	return EffectObservation{
		Known:                   true,
		EvidenceDigest:          digestBytes(raw),
		OutputPath:              evidence.OutputPath,
		OutputOutsideRepository: evidence.OutputOutsideRepository,
		PermissionEvidence:      evidence.PermissionEvidence,
		RepositoryWrites:        writes,
		MutationAuthority:       evidence.MutationAuthority,
	}, nil
}

func parseReplayEvidence(raw []byte) (ReplayEvidence, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return ReplayEvidence{}, fmt.Errorf("replay evidence is absent")
	}
	var evidence ReplayEvidence
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		return ReplayEvidence{}, fmt.Errorf("decode replay evidence: %w", err)
	}
	digest := evidence.EvidenceDigest
	evidence.EvidenceDigest = ""
	if evidence.Schema != ReplaySchema || evidence.Producer != ReplayProducer || len(evidence.ReceiptDigestsA) != len(contractCases()) || len(evidence.ReceiptDigestsB) != len(contractCases()) || len(evidence.ExecutionDigestsA) != len(contractCases()) || len(evidence.ExecutionDigestsB) != len(contractCases()) {
		return ReplayEvidence{}, fmt.Errorf("replay evidence shape is invalid")
	}
	if digest == "" || digest != digestValue(evidence) {
		return ReplayEvidence{}, fmt.Errorf("replay evidence digest is invalid")
	}
	evidence.EvidenceDigest = digest
	return evidence, nil
}
