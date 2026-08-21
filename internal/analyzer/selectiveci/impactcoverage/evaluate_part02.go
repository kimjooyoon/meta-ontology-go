package impactcoverage

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"sort"
)

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
