package metainvocation

import "sort"

type rule struct {
	operation string
	checkID   string
	command   string
	activity  string
}

var rules = []rule{
	{operation: operationDocsRule, checkID: "docs-check", command: "gooo check docs", activity: "SelectDocsCheck"},
	{operation: operationGoRule, checkID: "go-test", command: "go test ./...", activity: "SelectGoCheck"},
	{operation: operationYAMLRule, checkID: "yaml-check", command: "gooo check yaml", activity: "SelectYAMLCheck"},
}

func selectChecks(program Program, changeSet ChangeSet) ([]PlannedCheck, []UnknownCause) {
	filesByRule := map[string][]string{}
	unknowns := []UnknownCause{}
	for _, file := range changeSet.Files {
		matched := false
		for _, candidate := range rules {
			if ruleMatches(candidate.operation, file) {
				filesByRule[candidate.operation] = append(filesByRule[candidate.operation], file)
				matched = true
			}
		}
		if !matched {
			unknowns = append(unknowns, UnknownCause{Stage: "RULE_SELECTION", Step: "classify-change", Reason: "NO_REGISTERED_RULE", File: file})
		}
	}
	if len(unknowns) != 0 {
		return nil, unknowns
	}
	checks := make([]PlannedCheck, 0, len(filesByRule))
	for _, candidate := range rules {
		files := filesByRule[candidate.operation]
		if len(files) == 0 {
			continue
		}
		sort.Strings(files)
		checks = append(checks, plannedCheck(program, candidate, files))
	}
	return checks, nil
}

func plannedCheck(program Program, candidate rule, files []string) PlannedCheck {
	bound := program.Operations[candidate.activity]
	reasons := make([]RuleEvidence, 0, len(files))
	for _, file := range files {
		reasons = append(reasons, RuleEvidence{ID: candidate.operation + ":" + file, Operation: candidate.operation, File: file, SpecDigest: bound.SpecDigest, Source: bound.Source})
	}
	return PlannedCheck{ID: candidate.checkID, Command: candidate.command, Files: files, Reasons: reasons}
}
