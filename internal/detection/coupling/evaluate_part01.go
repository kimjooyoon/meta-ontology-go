package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func Evaluate(input Input, authority AuthorityContext) (result Result) {
	inputDigest := inputIdentityDigest(input, authority)
	defer func() {
		result.InputDigest = inputDigest
		result.Digest = stableDigest(resultCanonical(result))
	}()
	if authority.Schema == "" {
		return resultFor(StatusUnknown, ReasonAuthorityInputSelfBound, "evaluator authority context is missing", ObservationVector{})
	}
	authorityConfig, issue := normalizeAuthorityContext(authority)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	if input.Schema != InputSchemaV1 {
		return resultFor(StatusFailClosed, ReasonMalformedBinding, "input schema", ObservationVector{})
	}
	if issue := normalizeConfig(input.Config); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	if issue := validatePacketApplicability(input, authority); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	if issue := comparePacketAuthority(input, authorityConfig, authority); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	registry, issue := normalizeRegistry(input.Registry, authorityConfig)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	changed, issue := normalizeManifest(input.Manifest, authorityConfig, registry)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	observation := baseObservation(len(changed), len(input.Receipts))
	if issue := validateExternalReceipt(input.ExternalReceipt, authorityConfig, &observation); issue != nil {
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
		if issue := validateReceipt(receipt, entry, authorityConfig, registry[entry.SurfaceID]); issue != nil {
			return resultFor(issue.status, issue.code, issue.detail, observation)
		}
		if issue := validateReceiptPath(receipt, entry, path); issue != nil {
			return resultFor(issue.status, issue.code, issue.detail, observation)
		}
		accepted = append(accepted, entry.SurfaceID)
	}
	return passResult(accepted, observation)
}
