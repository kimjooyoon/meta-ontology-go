package operationconformance

import (
	"path/filepath"
	"sort"
)

func observeAtomic(evidence SplitGoEvidence) Decision {
	receipt := evidence.Write
	if !receipt.Complete {
		return DecisionUnknown
	}
	targets := candidatePaths(evidence.Candidates)
	if !receipt.ExecutionSucceeded || len(targets) < 2 ||
		receipt.WritesOutsideDeclaredTargets != 0 || receipt.TemporaryFilesRemaining != 0 ||
		!sameStrings(receipt.DeclaredTargets, targets) {
		return DecisionFail
	}
	staged := map[string]int{}
	committed, syncs := make([]string, 0, len(targets)), 0
	for index, event := range receipt.Events {
		if !event.Success {
			return DecisionFail
		}
		switch event.Kind {
		case "STAGE":
			if event.Temporary == "" || filepath.Dir(event.Temporary) != filepath.Dir(event.Target) {
				return DecisionFail
			}
			staged[event.Target] = index
		case "RENAME_CREATE", "RENAME_REPLACE":
			if stage, ok := staged[event.Target]; !ok || stage >= index {
				return DecisionFail
			}
			committed = append(committed, event.Target)
		case "DIRECT_WRITE":
			return DecisionFail
		case "DIRECTORY_SYNC":
			syncs++
		}
	}
	if !sameStrings(committed, targets) || committed[len(committed)-1] != evidence.Source.Path || syncs != 1 {
		return DecisionFail
	}
	return DecisionPass
}

func candidatePaths(files []FileEvidence) []string {
	result := make([]string, len(files))
	for index, file := range files {
		result[index] = file.Path
	}
	return result
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	a, b := append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(a)
	sort.Strings(b)
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
