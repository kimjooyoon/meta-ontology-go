package metacircularboundaryconsumer

import (
	"bytes"
	"encoding/json"
	"fmt"

	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
)

func parseGrantEvidence(raw []byte, semanticDigest string) (map[string]contract.ExternalGrant, string, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, "", fmt.Errorf("external grant artifact is absent")
	}
	var artifact contract.ExternalGrantArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		return nil, "", fmt.Errorf("decode external grant artifact: %w", err)
	}
	digest := digestBytes(raw)
	if artifact.Schema != grantSchema || artifact.Producer != grantProducer || artifact.Policy != "READ_ONLY_EXTERNAL_GRANT_POLICY_V1" {
		return nil, digest, fmt.Errorf("external grant artifact identity is invalid")
	}
	if artifact.ArtifactDigest != "" && artifact.ArtifactDigest != digest {
		return nil, digest, fmt.Errorf("external grant artifact digest mismatch")
	}
	grants := make(map[string]contract.ExternalGrant, len(artifact.Grants))
	for _, grant := range artifact.Grants {
		if grant.CaseID == "" || grant.GrantDigest != "" {
			return nil, digest, fmt.Errorf("external grant has invalid identity")
		}
		if _, exists := grants[grant.CaseID]; exists {
			return nil, digest, fmt.Errorf("duplicate external grant %q", grant.CaseID)
		}
		if grant.Decision != grantDecision && grant.Decision != grantDeny {
			return nil, digest, fmt.Errorf("external grant %q has unknown decision", grant.CaseID)
		}
		if grant.Decision == grantDecision && (grant.Issuer == "" || grant.SubjectDigest == "" || grant.Operation == "" || grant.Scope == "" || grant.Handle == "") {
			return nil, digest, fmt.Errorf("external grant %q is incomplete", grant.CaseID)
		}
		grant.GrantDigest = digestValue(grant)
		grants[grant.CaseID] = grant
	}
	for _, definition := range expectedCases() {
		if _, ok := grants[definition.ID]; !ok {
			return nil, digest, fmt.Errorf("external grant %q is absent", definition.ID)
		}
	}
	return grants, digest, nil
}

func parseEffectEvidence(raw []byte) (contract.EffectObservation, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return contract.EffectObservation{}, fmt.Errorf("workspace effect artifact is absent")
	}
	var evidence contract.EffectEvidence
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		return contract.EffectObservation{}, fmt.Errorf("decode workspace effect artifact: %w", err)
	}
	if evidence.Schema != effectSchema || evidence.Producer != effectProducer || evidence.OutputPath == "" || evidence.PermissionEvidence != "workflow-contents-read-only" || evidence.MutationAuthority == "" {
		return contract.EffectObservation{}, fmt.Errorf("workspace effect artifact identity is invalid")
	}
	if !validDigest(evidence.TrackedBeforeDigest) || !validDigest(evidence.TrackedAfterDigest) || !validDigest(evidence.UntrackedBeforeDigest) || !validDigest(evidence.UntrackedAfterDigest) {
		return contract.EffectObservation{}, fmt.Errorf("workspace effect snapshot digest is invalid")
	}
	if evidence.MutationAuthority != authorityDenied && evidence.MutationAuthority != authorityGranted && evidence.MutationAuthority != authorityUnknown {
		return contract.EffectObservation{}, fmt.Errorf("workspace mutation authority is invalid")
	}
	writes := 0
	if evidence.TrackedBeforeDigest != evidence.TrackedAfterDigest || evidence.UntrackedBeforeDigest != evidence.UntrackedAfterDigest {
		writes = 1
	}
	return contract.EffectObservation{Known: true, EvidenceDigest: digestBytes(raw), OutputPath: evidence.OutputPath, OutputOutsideRepository: evidence.OutputOutsideRepository, PermissionEvidence: evidence.PermissionEvidence, RepositoryWrites: writes, MutationAuthority: evidence.MutationAuthority}, nil
}

func parseReplayEvidence(raw []byte) (contract.ReplayEvidence, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return contract.ReplayEvidence{}, fmt.Errorf("replay evidence is absent")
	}
	var evidence contract.ReplayEvidence
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		return contract.ReplayEvidence{}, fmt.Errorf("decode replay evidence: %w", err)
	}
	if evidence.Schema != replaySchema || evidence.Producer != replayProducer || len(evidence.ReceiptDigestsA) != len(expectedCases()) || len(evidence.ReceiptDigestsB) != len(expectedCases()) || len(evidence.ExecutionDigestsA) != len(expectedCases()) || len(evidence.ExecutionDigestsB) != len(expectedCases()) {
		return contract.ReplayEvidence{}, fmt.Errorf("replay evidence shape is invalid")
	}
	digest := evidence.EvidenceDigest
	evidence.EvidenceDigest = ""
	if digest == "" || digest != digestValue(evidence) {
		return contract.ReplayEvidence{}, fmt.Errorf("replay evidence digest is invalid")
	}
	evidence.EvidenceDigest = digest
	return evidence, nil
}

func validExecutionArtifact(source contract.SourceObservation, grant contract.ExternalGrant, caseID string, artifact contract.ExecutionArtifact) bool {
	if artifact.Schema != executionSchema || artifact.Producer != executionProducer || artifact.CaseID != caseID || artifact.OperationID != metaOperationID || artifact.Path != "execution/"+caseID+".json" || artifact.GrantDigest != grant.GrantDigest || artifact.InputDigest != source.SemanticDigest || artifact.OutputCanonical == "" || artifact.OutputDigest != digestBytes([]byte(artifact.OutputCanonical)) {
		return false
	}
	copy := artifact
	copy.ArtifactDigest = ""
	return artifact.ArtifactDigest != "" && artifact.ArtifactDigest == digestValue(copy)
}
