package coupling

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func Evaluate(input Input) (result Result) {
	inputDigest := inputIdentityDigest(input)
	defer func() {
		result.InputDigest = inputDigest
		result.Digest = stableDigest(resultCanonical(result))
	}()
	if input.Schema != InputSchemaV1 {
		return resultFor(StatusFailClosed, ReasonMalformedBinding, "input schema", ObservationVector{})
	}
	if issue := normalizeConfig(input.Config); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	registry, issue := normalizeRegistry(input.Registry, input.Config)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	changed, issue := normalizeManifest(input.Manifest, input.Config, registry)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	observation := baseObservation(len(changed), len(input.Receipts))
	if issue := validateExternalReceipt(input.ExternalReceipt, input.Config, &observation); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, observation)
	}
	if len(changed) == 0 {
		if len(input.Receipts) != 0 {
			return resultFor(StatusFailClosed, ReasonOrphanReceipt, "receipt for zero-change manifest", observation)
		}
		observation.DeterministicWork = knownDimension(0)
		return passResult(nil, observation)
	}
	path, issue := normalizeInferencePath(input.InferencePath)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, observation)
	}
	observation.InferenceRecords = knownDimension(uint64(len(path.Edges) + len(path.Claims) + len(path.Evidence)))
	observation.InferencePaths = knownDimension(uint64(len(input.Receipts)))
	observation.DeterministicWork = knownDimension(uint64(len(changed) + len(input.Receipts) + len(path.Edges) + len(path.Claims) + len(path.Evidence)))
	receipts, issue := indexReceipts(input.Receipts, changed)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, observation)
	}
	accepted := make([]semantic.ID, 0, len(changed))
	for _, entry := range changed {
		receipt, exists := receipts[entry.SurfaceID]
		if !exists {
			return resultFor(StatusUnknown, ReasonRequiredInputMissing, "receipt for changed surface", observation)
		}
		if issue := validateReceipt(receipt, entry, input.Config, registry[entry.SurfaceID]); issue != nil {
			return resultFor(issue.status, issue.code, issue.detail, observation)
		}
		if issue := validateReceiptPath(receipt, entry, path); issue != nil {
			return resultFor(issue.status, issue.code, issue.detail, observation)
		}
		accepted = append(accepted, entry.SurfaceID)
	}
	return passResult(accepted, observation)
}

func baseObservation(changed, receipts int) ObservationVector {
	return ObservationVector{
		ChangedSurfaces: knownDimension(uint64(changed)),
		Receipts:        knownDimension(uint64(receipts)),
	}
}

func knownDimension(value uint64) CountDimension { return CountDimension{Known: true, Value: value} }

func indexReceipts(receipts []CouplingReceipt, changed []ManifestEntry) (map[semantic.ID]CouplingReceipt, *evaluationIssue) {
	changedIDs := make(map[semantic.ID]struct{}, len(changed))
	for _, entry := range changed {
		changedIDs[entry.SurfaceID] = struct{}{}
	}
	indexed := make(map[semantic.ID]CouplingReceipt, len(receipts))
	seenReceiptIDs := make(map[semantic.ID]struct{}, len(receipts))
	for _, receipt := range receipts {
		if _, duplicate := seenReceiptIDs[receipt.ReceiptID]; duplicate {
			return nil, failIssue(ReasonDuplicateReceipt, receipt.SurfaceID.String())
		}
		seenReceiptIDs[receipt.ReceiptID] = struct{}{}
		if _, expected := changedIDs[receipt.SurfaceID]; !expected {
			return nil, failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
		}
		if _, duplicate := indexed[receipt.SurfaceID]; duplicate {
			return nil, failIssue(ReasonDuplicateReceipt, receipt.SurfaceID.String())
		}
		indexed[receipt.SurfaceID] = receipt
	}
	return indexed, nil
}

func validateReceipt(receipt CouplingReceipt, entry ManifestEntry, config Config, surface Surface) *evaluationIssue {
	if receipt.Schema != ReceiptSchemaV1 {
		return failIssue(ReasonMalformedBinding, receipt.SurfaceID.String())
	}
	if receipt.State != ReceiptStateCurrent {
		return unknownIssue(ReasonStaleInput, receipt.SurfaceID.String())
	}
	if _, issue := normalizeID(receipt.ReceiptID, "receipt ID"); issue != nil {
		return issue
	}
	if receipt.SurfaceID != surface.SurfaceID || receipt.CodeSymbolID != surface.CodeSymbolID || receipt.SemanticOwnerID != surface.SemanticOwnerID {
		return failIssue(ReasonSourceMapMismatch, receipt.SurfaceID.String())
	}
	if receipt.SourceMapBindingDigest != surface.Binding.BindingDigest {
		return failIssue(ReasonSourceMapMismatch, receipt.SurfaceID.String())
	}
	if receipt.SnapshotDigest != config.SnapshotDigest || receipt.RegistryDigest != config.RegistryDigest ||
		receipt.ToolchainDigest != config.ToolchainDigest || receipt.ProfileDigest != config.ProfileDigest {
		return unknownIssue(ReasonStaleInput, receipt.SurfaceID.String())
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{receipt.SourceMapBindingDigest, "receipt binding digest"},
		{receipt.SnapshotDigest, "receipt snapshot digest"},
		{receipt.RegistryDigest, "receipt registry digest"},
		{receipt.ToolchainDigest, "receipt toolchain digest"},
		{receipt.ProfileDigest, "receipt profile digest"},
		{receipt.BeforeBlobDigest, "receipt before blob digest"},
		{receipt.AfterBlobDigest, "receipt after blob digest"},
		{receipt.BeforeAuthoritySourceDigest, "receipt before source digest"},
		{receipt.AfterAuthoritySourceDigest, "receipt after source digest"},
		{receipt.BeforeCanonicalSemanticDigest, "receipt before semantic digest"},
		{receipt.AfterCanonicalSemanticDigest, "receipt after semantic digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return issue
		}
	}
	if receipt.BeforeBlobDigest != entry.BeforeBlobDigest || receipt.AfterBlobDigest != entry.AfterBlobDigest {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	wantKind, validClaim := semanticKindForClaim(receipt.ChangeClaim)
	if !validClaim || receipt.ReceiptKind != wantKind {
		return failIssue(ReasonContradictoryReceipt, receipt.SurfaceID.String())
	}
	switch receipt.ChangeClaim {
	case ChangeClaimDelta:
		if receipt.BeforeCanonicalSemanticDigest == receipt.AfterCanonicalSemanticDigest {
			return failIssue(ReasonContradictoryReceipt, receipt.SurfaceID.String())
		}
		if receipt.BeforeAuthoritySourceDigest == receipt.AfterAuthoritySourceDigest {
			return failIssue(ReasonDeltaWithoutSource, receipt.SurfaceID.String())
		}
		if receipt.CanonicalDelta == "" || receipt.DeltaDigest == "" ||
			stableDigest(receipt.CanonicalDelta) != receipt.DeltaDigest {
			return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
		}
	case ChangeClaimNoDelta:
		if receipt.BeforeCanonicalSemanticDigest != receipt.AfterCanonicalSemanticDigest ||
			receipt.BeforeAuthoritySourceDigest != receipt.AfterAuthoritySourceDigest {
			return failIssue(ReasonNoDeltaWithoutEquality, receipt.SurfaceID.String())
		}
		if receipt.CanonicalDelta != "" || receipt.DeltaDigest != "" || receipt.AuthoritativeSource != nil {
			return failIssue(ReasonNoDeltaWithoutEquality, receipt.SurfaceID.String())
		}
	default:
		return failIssue(ReasonMalformedBinding, receipt.SurfaceID.String())
	}
	if receipt.AuthoritativeSource != nil {
		if _, issue := normalizeID(receipt.AuthoritativeSource.SourceID, "authority source ID"); issue != nil {
			return issue
		}
	}
	return validateReceiptReferences(receipt)
}

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
