package coupling

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateReceiptReferences(receipt CouplingReceipt) *evaluationIssue {
	if _, issue := normalizeID(receipt.InferenceClaimID, "inference claim ID"); issue != nil {
		return issue
	}
	seen := make(map[semantic.ID]struct{}, len(receipt.OriginPathIDs))
	for _, pathID := range receipt.OriginPathIDs {
		if _, issue := normalizeID(pathID, "origin path ID"); issue != nil {
			return issue
		}
		if _, duplicate := seen[pathID]; duplicate {
			return failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
		}
		seen[pathID] = struct{}{}
	}
	seenEvidence := make(map[semantic.ID]struct{}, len(receipt.EvidenceRefs))
	for _, ref := range receipt.EvidenceRefs {
		if _, issue := normalizeID(ref.ID, "evidence ID"); issue != nil {
			return issue
		}
		if issue := normalizeDigestValue(ref.Digest, "evidence digest"); issue != nil {
			return issue
		}
		if _, duplicate := seenEvidence[ref.ID]; duplicate {
			return failIssue(ReasonInferencePathMalformed, receipt.SurfaceID.String())
		}
		seenEvidence[ref.ID] = struct{}{}
	}
	return nil
}

func validateExternalReceipt(
	receipt *ExternalResourceReceipt, config Config, observation *ObservationVector,
) *evaluationIssue {
	if receipt == nil {
		if config.ExternalReceiptRequired {
			return unknownIssue(ReasonExternalReceiptMissing, "external resource receipt")
		}
		return nil
	}
	if receipt.Schema != ResourceSchemaV1 {
		return failIssue(ReasonMalformedBinding, "external receipt schema")
	}
	if receipt.SnapshotDigest == "" || receipt.ProviderDigest == "" || receipt.ObserverDigest == "" || receipt.CPUWorkUnits == nil || receipt.PeakMemoryBytes == nil {
		return unknownIssue(ReasonExternalReceiptMissing, "external receipt binding or value")
	}
	if config.ExternalReceiptRequired && receipt.DeterministicWorkUnits == nil {
		return unknownIssue(ReasonExternalReceiptMissing, "external deterministic work")
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{receipt.SnapshotDigest, "external snapshot digest"},
		{receipt.ProviderDigest, "external provider digest"},
		{receipt.ObserverDigest, "external observer digest"},
		{receipt.Digest, "external receipt digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return issue
		}
	}
	if receipt.SnapshotDigest != config.SnapshotDigest || receipt.ProviderDigest != config.ExpectedProviderDigest ||
		receipt.ObserverDigest != config.ExpectedObserverDigest || stableDigest(externalCanonical(*receipt)) != receipt.Digest {
		return failIssue(ReasonDigestMismatch, "external resource receipt")
	}
	observation.CPU = knownDimension(*receipt.CPUWorkUnits)
	observation.Memory = knownDimension(*receipt.PeakMemoryBytes)
	if receipt.DeterministicWorkUnits != nil {
		observation.ResourceWork = knownDimension(*receipt.DeterministicWorkUnits)
	}
	return nil
}

func passResult(accepted []semantic.ID, observation ObservationVector) Result {
	result := Result{
		Schema: ResultSchemaV1, Status: StatusPass, AcceptedSurfaceIDs: sortedIDs(accepted),
		Observation: observation, FullSuiteRequired: false,
	}
	result.Digest = stableDigest(resultCanonical(result))
	return result
}

func resultFor(status Status, code ReasonCode, detail string, observation ObservationVector) Result {
	result := Result{
		Schema: ResultSchemaV1, Status: status,
		Reasons: []Reason{{Code: code, Detail: detail}}, Observation: observation,
		FullSuiteRequired: status != StatusPass,
	}
	result.Reasons = sortedReasons(result.Reasons)
	result.Digest = stableDigest(resultCanonical(result))
	return result
}

func SortReasons(result *Result) {
	if result == nil {
		return
	}
	result.Reasons = sortedReasons(result.Reasons)
	sort.Slice(result.AcceptedSurfaceIDs, func(i, j int) bool { return result.AcceptedSurfaceIDs[i] < result.AcceptedSurfaceIDs[j] })
	result.Digest = stableDigest(resultCanonical(*result))
}
