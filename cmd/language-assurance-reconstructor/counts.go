package main

import "fmt"

func countSelfMinting(routes []authorityRoute) int {
	count := 0
	for _, route := range routes { if route.AuthoredBy == route.PromotedBy { count++ } }
	return count
}

func countRoleConflicts(bindings []roleBinding) int {
	count := 0
	for _, binding := range bindings {
		roles := map[string]bool{}
		for _, role := range binding.Roles { roles[role] = true }
		for _, pair := range conflictPairs { if roles[pair[0]] && roles[pair[1]] { count++ } }
	}
	return count
}

func countUnknown(transitions []decisionTransition) (int, int) {
	laundering, top := 0, 0
	for _, transition := range transitions {
		if transition.Input == unknown {
			top++
			if launderingOutputs[transition.Output] { laundering++ }
		}
	}
	return laundering, top
}

func observeSnapshots(subjectSHA string, bindings []snapshotBinding) (*int, *int, error) {
	seen, mismatches := map[string]bool{}, 0
	for _, binding := range bindings {
		if seen[binding.EvidenceID] || !snapshotEvidenceIDs[binding.EvidenceID] || !validSHA(binding.SubjectSHA) {
			return nil, nil, fmt.Errorf("snapshot binding is malformed")
		}
		seen[binding.EvidenceID] = true
		if binding.SubjectSHA != subjectSHA { mismatches++ }
	}
	if len(bindings) != len(snapshotEvidenceIDs) { return nil, nil, nil }
	bps, paths := (len(snapshotEvidenceIDs)-mismatches)*10000/len(snapshotEvidenceIDs), mismatches
	return &bps, &paths, nil
}
