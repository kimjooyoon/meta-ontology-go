package impactcoverage

import (
	"fmt"
	"sort"

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

func scanSnapshots(base, head *selectiveci.Snapshot) (scanResult, error) {
	pairs := make(map[string]sourcePair, len(base.Sources)+len(head.Sources))
	for index := range base.Sources {
		source := &base.Sources[index]
		pair := pairs[source.Path]
		pair.before = source
		pairs[source.Path] = pair
	}
	for index := range head.Sources {
		source := &head.Sources[index]
		pair := pairs[source.Path]
		pair.after = source
		pairs[source.Path] = pair
	}
	paths := make([]string, 0, len(pairs))
	for repoPath := range pairs {
		paths = append(paths, repoPath)
	}
	sort.Strings(paths)
	result := scanResult{}
	var bindingRecords uint64
	for _, repoPath := range paths {
		pair := pairs[repoPath]
		beforeIDs, afterIDs := sourceIDs(pair.before), sourceIDs(pair.after)
		var err error
		bindingRecords, err = addLen(bindingRecords, pair)
		if err != nil {
			return scanResult{}, err
		}
		if !changed(pair, beforeIDs, afterIDs) {
			continue
		}
		result.changedBlobs++
		union := unionIDs(beforeIDs, afterIDs)
		if len(union) == 0 {
			result.uncoveredBlobs++
			result.uncoveredPaths = append(result.uncoveredPaths, repoPath)
			continue
		}
		result.coveredBlobs++
		result.changedIDs = append(result.changedIDs, union...)
	}
	result.changedIDs = sortedUnique(result.changedIDs)
	result.uncoveredPaths = sortedUnique(result.uncoveredPaths)
	// deterministic_work_units = scanned_paths + scanned_binding_records.
	var err error
	result.workUnits, err = checkedAdd(uint64(len(paths)), bindingRecords)
	if err != nil {
		return scanResult{}, err
	}
	return result, nil
}

func sourceIDs(source *selectiveci.Source) []string {
	if source == nil {
		return nil
	}
	ids := make([]string, 0, len(source.Bindings))
	for _, binding := range source.Bindings {
		ids = append(ids, binding.ID)
	}
	return sortedUnique(ids)
}

func addLen(total uint64, pair sourcePair) (uint64, error) {
	count := uint64(0)
	if pair.before != nil {
		count = uint64(len(pair.before.Bindings))
	}
	if pair.after != nil {
		var err error
		count, err = checkedAdd(count, uint64(len(pair.after.Bindings)))
		if err != nil {
			return 0, err
		}
	}
	return checkedAdd(total, count)
}

func changed(pair sourcePair, beforeIDs, afterIDs []string) bool {
	if pair.before == nil || pair.after == nil {
		return true
	}
	return pair.before.BlobDigest != pair.after.BlobDigest ||
		!equalStrings(beforeIDs, afterIDs)
}

func unionIDs(left, right []string) []string {
	values := append(append([]string{}, left...), right...)
	return sortedUnique(values)
}

func sortedUnique(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	if len(result) < 2 {
		return result
	}
	write := 1
	for _, value := range result[1:] {
		if value != result[write-1] {
			result[write] = value
			write++
		}
	}
	return result[:write]
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func checkedAdd(left, right uint64) (uint64, error) {
	if ^uint64(0)-left < right {
		return 0, fmt.Errorf("deterministic work unit overflow")
	}
	return left + right, nil
}

func resultFor(input Input) Result {
	result := Result{Schema: SchemaV1, ChangedStableIDs: []string{}, UncoveredPaths: []string{}}
	if input.Base == nil || input.Head == nil {
		return result
	}
	result.BaseSnapshotDigest = input.Base.Digest
	result.HeadSnapshotDigest = input.Head.Digest
	result.BaseSourceMapDigest = input.Base.SourceMapDigest
	result.HeadSourceMapDigest = input.Head.SourceMapDigest
	result.BaseRegistryDigest = input.Base.RegistryDigest
	result.HeadRegistryDigest = input.Head.RegistryDigest
	return result
}

func seal(result Result, decision Decision, reason Reason) Result {
	result.Decision = decision
	result.Reason = reason
	result.FullSuiteRequired = decision == DecisionUnknown
	if decision == DecisionUnknown {
		result.ChangedStableIDs = []string{}
	}
	result.ChangedStableIDs = sortedUnique(result.ChangedStableIDs)
	result.UncoveredPaths = sortedUnique(result.UncoveredPaths)
	result.OutputDigest = ""
	canonical, err := result.CanonicalJSON()
	if err == nil {
		result.OutputDigest = digestBytes(canonical)
	}
	return result
}
