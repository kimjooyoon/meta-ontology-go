package externalcapabilityexecution

func exactObservation(subject string) Observation {
	run := CapabilityRun{
		Status: StatusSatisfied, Arithmetic: "42", Function: "55",
		EvaluatorExitCode: 0, MacroExitCode: 0,
		MacroGeneratedSHA256: "sha256:fixture", MacroExpectedSHA256: "sha256:fixture",
		NormalizedSHA256: "sha256:replay",
	}
	first, second := run, run
	first.RunID, second.RunID = "run-1", "run-2"
	return sealObservation(Observation{
		Schema: ObservationSchema, SubjectSHA: subject, Available: true,
		Reference: Reference{
			RepositoryURL: ExpectedRepository, CommitSHA: ExpectedCommit,
			TreeSHA: ExpectedTree, GoVersion: ExpectedGoVersion,
		},
		Parent: ParentReport{
			Decision: DecisionFailClosed, Resolution: ResolutionExact,
			Completed: 6, Total: 8, BasisPoints: 7500,
		},
		Runs: []CapabilityRun{first, second}, ReplayExact: true,
		ExternalExecutions: 4, UnknownEvents: []string{},
	})
}

func cloneObservation(source Observation) Observation {
	clone := source
	clone.Runs = append([]CapabilityRun(nil), source.Runs...)
	clone.UnknownEvents = append([]string(nil), source.UnknownEvents...)
	return clone
}
