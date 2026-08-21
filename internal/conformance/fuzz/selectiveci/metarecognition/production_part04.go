package metarecognition

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"math"
)

func evaluatePath(b BaselineConfig) Outcome {
	pathID := "path://unknown"
	if len(b.Path.IDs) > 0 {
		pathID = b.Path.IDs[0]
	}
	id := semantic.MustIdentity(pathID)
	requirement := pathclosure.Requirement{PathID: id, RecordIDs: []semantic.ID{semantic.MustIdentity("record://one")}, ExpectedKinds: []semantic.InferenceKind{semantic.InferenceAuthoritativeDeclaration}, StartID: semantic.MustIdentity("node://start"), EndID: semantic.MustIdentity("node://end")}
	requirements := []pathclosure.Requirement{requirement}
	if b.Path.Duplicate {
		requirements = append(requirements, requirement)
	}
	if b.Path.Conflict {
		requirement.RecordIDs = []semantic.ID{semantic.MustIdentity("record://two")}
		requirements = append(requirements, requirement)
	}
	path := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}
	result := pathclosure.Evaluate(path, requirements)
	work := Work{Units: result.Numerator, Selected: result.Numerator, Full: result.Denominator, ProvRecords: len(path.Edges) + len(path.Evidence), ProvPaths: result.Denominator}
	if result.Status == pathclosure.FAIL_CLOSED && b.Path.Conflict {
		return productionOutcome(FailClosedUnsound, ReasonConflictingReceipt, []string{id.String()}, work)
	}
	if result.Status == pathclosure.FAIL_CLOSED {
		return productionOutcome(FailClosedUnsound, ReasonDuplicateReceipt, []string{id.String()}, work)
	}
	return productionOutcome(UnknownFullSuiteRequired, ReasonExternalMissing, []string{id.String()}, work)
}
func evaluateResource(b BaselineConfig) Outcome {
	samples := make([]resourceenvelope.Sample, 6)
	for index := range samples {
		samples[index] = resourceenvelope.Sample{CPUCoreNS: math.MaxUint64, WallNS: 1}
	}
	envelope := resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: digest("r"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5, Limits: resourceenvelope.Limits{CPUCoreNS: math.MaxUint64, PeakRSSBytes: math.MaxUint64, ReadBytes: math.MaxUint64, WriteBytes: math.MaxUint64}, Samples: samples}
	result := resourceenvelope.Evaluate(envelope)
	work := Work{Units: int(envelope.SampleCount), Selected: int(envelope.SampleCount), Full: len(envelope.Samples), ProvRecords: len(envelope.Samples), ProvPaths: 1}
	if result.Status != resourceenvelope.PASS {
		return productionOutcome(UnknownFullSuiteRequired, ReasonInvalidResource, []string{"receipt-1"}, work)
	}
	return productionOutcome(ClosedSound, ReasonExactBinding, nil, work)
}
func commandIDs(values []CommandAssertion, include func(CommandAssertion) bool) []string {
	ids := make([]string, 0, len(values))
	for _, value := range values {
		if include(value) {
			ids = append(ids, value.ID)
		}
	}
	return sorted(ids)
}
func externalMissing(value ExternalAssertion) bool {
	return !(value.Authenticity && value.Provider && value.Phase && value.Observer)
}
func externalInputID(value ExternalAssertion) string {
	if !value.Authenticity {
		return "external-authenticity"
	}
	if !value.Provider {
		return "external-provider"
	}
	if !value.Phase {
		return "external-phase"
	}
	return "external-observer"
}
