package externalecosystemexecution

func RunSuite() SuiteReport {
	names := []string{"exact", "observation-unavailable", "reference-unavailable", "reference-decision-unknown",
		"commit-mismatch", "go-version-mismatch", "execution-failure", "replay-mismatch", "source-write", "unknown-event"}
	suite := SuiteReport{Schema: "external-ecosystem-execution-suite/v1", Total: len(names), UnknownExpected: 3, InvariantExpected: 6}
	for _, name := range names {
		var observation *Observation
		if name != "observation-unavailable" {
			value := exactObservation()
			observation = &value
			mutateCase(name, observation)
		}
		report := Evaluate(observation)
		wantDecision, wantResolution := DecisionFailClosed, "EXACT"
		if name == "exact" {
			wantDecision = DecisionConfirmed
		}
		if name == "observation-unavailable" || name == "reference-unavailable" || name == "reference-decision-unknown" {
			wantResolution = "COARSE"
		}
		passed := report.Decision == wantDecision && report.Resolution == wantResolution
		if passed {
			suite.Passed++
		} else {
			suite.Unresolved++
		}
		suite.Cases = append(suite.Cases, SuiteCase{name, wantDecision, wantResolution, report.Decision, report.Resolution, passed})
	}
	return suite
}

func exactObservation() Observation {
	outcomes := []Outcome{{Package: "github.com/cosmos72/gomacro", Action: "pass"}}
	digest := Digest(outcomes)
	state := RepositoryState{Available: true, Commit: "source", Tree: "tree", StatusSHA256: Digest("")}
	external := RepositoryState{Available: true, Commit: ExpectedCommit, Tree: ExpectedTree, StatusSHA256: Digest("")}
	return Observation{Schema: ObservationSchema, GoVersion: ExpectedGoVersion,
		Reference: ReferenceReceipt{Available: true, BindingExact: true, ContractVersion: ReferenceContractVersion,
			Decision: ExpectedReferenceDecision, Resolution: "EXACT", URL: ExpectedReferenceURL, Commit: ExpectedCommit,
			Tree: ExpectedTree, ModuleGo: ExpectedModuleGo, EvidencePath: referenceEvidencePath, EvidenceSHA256: Digest("evidence")},
		Runs: []RunObservation{{Index: 1, Passed: true, EventCount: 1, NormalizedSHA256: digest, Outcomes: outcomes},
			{Index: 2, Passed: true, EventCount: 1, NormalizedSHA256: digest, Outcomes: outcomes}},
		SourceBefore: state, SourceAfter: state, ExternalBefore: external, ExternalAfter: external,
		Regression: RegressionReceipt{Passed: 10, Total: 10}}
}

func mutateCase(name string, o *Observation) {
	switch name {
	case "reference-unavailable":
		o.Reference.Available = false
	case "reference-decision-unknown":
		o.Reference.Decision = "UNRECOGNIZED"
	case "commit-mismatch":
		o.Reference.Commit = "different"
	case "go-version-mismatch":
		o.GoVersion = "go1.26.0"
	case "execution-failure":
		o.Runs[1].Passed, o.Runs[1].ExitCode = false, 1
	case "replay-mismatch":
		o.Runs[1].NormalizedSHA256 = Digest("different")
	case "source-write":
		o.SourceAfter.Dirty = true
	case "unknown-event":
		o.Runs[0].Passed, o.Runs[0].UnknownEvents = false, []string{"future-action"}
	}
}
