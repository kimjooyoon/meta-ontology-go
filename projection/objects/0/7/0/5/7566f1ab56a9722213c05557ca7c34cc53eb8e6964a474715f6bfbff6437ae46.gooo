package impactcoverage

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
)

// Observe evaluates exact snapshot facts without selection, execution, or
// authority writes. Evaluate is the descriptive alias used by callers.
func Observe(input Input) Result {
	result := resultFor(input)
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return seal(result, DecisionUnknown, ReasonInvalidSnapshot)
	}
	result.InputDigest = digestBytes(canonical)
	scan, err := scanSnapshots(input.Base, input.Head)
	if err != nil {
		return seal(result, DecisionUnknown, ReasonInvalidSnapshot)
	}
	result.ChangedBlobCount = scan.changedBlobs
	result.CoveredChangedBlobCount = scan.coveredBlobs
	result.UncoveredChangedBlobCount = scan.uncoveredBlobs
	result.ChangedBindingCount = uint64(len(scan.changedIDs))
	result.DeterministicWorkUnits = scan.workUnits
	result.UncoveredPaths = append([]string{}, scan.uncoveredPaths...)
	if input.Base.SourceMapDigest != input.Head.SourceMapDigest ||
		input.Base.RegistryDigest != input.Head.RegistryDigest {
		return seal(result, DecisionUnknown, ReasonAuthorityDrift)
	}
	if len(scan.uncoveredPaths) != 0 {
		return seal(result, DecisionUnknown, ReasonMissingBinding)
	}
	result.ChangedStableIDs = append([]string{}, scan.changedIDs...)
	if scan.changedBlobs == 0 {
		return seal(result, DecisionExact, ReasonNoChange)
	}
	return seal(result, DecisionExact, ReasonComplete)
}
func Evaluate(input Input) Result { return Observe(input) }
func ObserveSnapshots(base, head selectiveci.Snapshot) Result {
	return Observe(NewInput(&base, &head))
}
func EvaluateSnapshots(base, head selectiveci.Snapshot) Result {
	return ObserveSnapshots(base, head)
}

type scanResult struct {
	changedBlobs   uint64
	coveredBlobs   uint64
	uncoveredBlobs uint64
	workUnits      uint64
	changedIDs     []string
	uncoveredPaths []string
}
type sourcePair struct {
	before *selectiveci.Source
	after  *selectiveci.Source
}
