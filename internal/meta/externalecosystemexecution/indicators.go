package externalecosystemexecution

func indicator(c Criterion, status, reason string) Indicator {
	n := 0
	if status == "SATISFIED" {
		n = 1
	}
	return Indicator{c.ID, c.Kind, status, n, 1, reason}
}

func buildIndicators(o *Observation) []Indicator {
	c := Criteria()
	if o == nil {
		result := make([]Indicator, len(c))
		for i := range c {
			result[i] = indicator(c[i], "UNKNOWN", "EXECUTION_OBSERVATION_UNAVAILABLE")
		}
		return result
	}
	refStatus, refReason := referenceStatus(o.Reference)
	commitStatus, commitReason := exactValue(o.Reference.Available, o.Reference.Commit, ExpectedCommit, "REFERENCE_COMMIT")
	treeStatus, treeReason := exactValue(o.Reference.Available, o.Reference.Tree, ExpectedTree, "REFERENCE_TREE")
	goStatus, goReason := exactValue(true, o.GoVersion, ExpectedGoVersion, "GO_VERSION")
	run1Status, run1Reason := runStatus(o.Runs, 0)
	run2Status, run2Reason := runStatus(o.Runs, 1)
	replayStatus, replayReason := replayStatus(o.Runs)
	writeStatus, writeReason := writeStatus(o)
	values := [][2]string{{refStatus, refReason}, {commitStatus, commitReason}, {treeStatus, treeReason},
		{goStatus, goReason}, {run1Status, run1Reason}, {run2Status, run2Reason}, {replayStatus, replayReason}, {writeStatus, writeReason}}
	result := make([]Indicator, len(c))
	for i := range c {
		result[i] = indicator(c[i], values[i][0], values[i][1])
	}
	return result
}

func referenceStatus(r ReferenceReceipt) (string, string) {
	if !r.Available {
		return "UNKNOWN", "REFERENCE_EVIDENCE_UNAVAILABLE"
	}
	if r.Decision != ExpectedReferenceDecision {
		return "UNKNOWN", "REFERENCE_DECISION_UNKNOWN"
	}
	if !r.BindingExact || r.ContractVersion != ReferenceContractVersion || r.Resolution != "EXACT" ||
		r.URL != ExpectedReferenceURL || r.ModuleGo != ExpectedModuleGo {
		return "UNSATISFIED", "REFERENCE_BINDING_MISMATCH"
	}
	return "SATISFIED", "REFERENCE_BINDING_EXACT"
}

func exactValue(available bool, actual, expected, name string) (string, string) {
	if !available || actual == "" {
		return "UNKNOWN", name + "_UNAVAILABLE"
	}
	if actual != expected {
		return "UNSATISFIED", name + "_MISMATCH"
	}
	return "SATISFIED", name + "_EXACT"
}

func runStatus(runs []RunObservation, index int) (string, string) {
	if len(runs) <= index {
		return "UNKNOWN", "EXTERNAL_RUN_UNAVAILABLE"
	}
	if len(runs[index].UnknownEvents) > 0 {
		return "UNSATISFIED", "EXTERNAL_EVENT_UNSUPPORTED"
	}
	if !runs[index].Passed || runs[index].ExitCode != 0 {
		return "UNSATISFIED", "EXTERNAL_RUN_FAILED"
	}
	return "SATISFIED", "EXTERNAL_RUN_PASSED"
}

func replayStatus(runs []RunObservation) (string, string) {
	if len(runs) < 2 || runs[0].NormalizedSHA256 == "" || runs[1].NormalizedSHA256 == "" {
		return "UNKNOWN", "NORMALIZED_REPLAY_UNAVAILABLE"
	}
	if runs[0].NormalizedSHA256 != runs[1].NormalizedSHA256 {
		return "UNSATISFIED", "NORMALIZED_REPLAY_MISMATCH"
	}
	return "SATISFIED", "NORMALIZED_REPLAY_EQUAL"
}
