package externalcapabilityexecution

func makeIndicators(observation Observation) []Indicator {
	known := observation.Available
	runsKnown := known && len(observation.Runs) == 2
	for _, run := range observation.Runs {
		runsKnown = runsKnown && (run.Status == StatusSatisfied || run.Status == StatusUnsatisfied)
	}
	arithmeticCount := countRuns(observation.Runs, func(run CapabilityRun) bool { return run.Arithmetic == "42" })
	functionCount := countRuns(observation.Runs, func(run CapabilityRun) bool { return run.Function == "55" })
	macroCount := countRuns(observation.Runs, func(run CapabilityRun) bool {
		return run.MacroGeneratedSHA256 != "" && run.MacroGeneratedSHA256 == run.MacroExpectedSHA256
	})
	replay := runsKnown && observation.ReplayExact && observation.Runs[0].NormalizedSHA256 != "" &&
		observation.Runs[0].NormalizedSHA256 == observation.Runs[1].NormalizedSHA256
	parentKnown := known && knownParent(observation.Parent)
	parentExact := parentKnown && observation.Parent.Decision == DecisionFailClosed &&
		observation.Parent.Resolution == ResolutionExact && observation.Parent.Completed == 6 &&
		observation.Parent.Total == 8 && observation.Parent.BasisPoints == 7500 &&
		observation.Parent.OfficialMutationCount == 0 && observation.Parent.PromotionCount == 0
	return []Indicator{
		indicator("reference", "DRIVER", "FOUNDATION", "bind-pinned-repository", known, observation.Reference.RepositoryURL == ExpectedRepository, boolInt(observation.Reference.RepositoryURL == ExpectedRepository), 1),
		indicator("commit", "DRIVER", "FOUNDATION", "verify-pinned-commit", known, observation.Reference.CommitSHA == ExpectedCommit, boolInt(observation.Reference.CommitSHA == ExpectedCommit), 1),
		indicator("tree", "DRIVER", "FOUNDATION", "verify-pinned-tree", known, observation.Reference.TreeSHA == ExpectedTree, boolInt(observation.Reference.TreeSHA == ExpectedTree), 1),
		indicator("go127", "DRIVER", "FOUNDATION", "select-go-toolchain", known, observation.Reference.GoVersion == ExpectedGoVersion, boolInt(observation.Reference.GoVersion == ExpectedGoVersion), 1),
		indicator("arithmetic", "OUTCOME", "COHERENCE", "execute-embedded-eval", runsKnown, arithmeticCount == 2, arithmeticCount, 2),
		indicator("function", "OUTCOME", "COHERENCE", "execute-interpreted-function", runsKnown, functionCount == 2, functionCount, 2),
		indicator("ast-macro", "OUTCOME", "COHERENCE", "execute-ast-macro-generation", runsKnown, macroCount == 2, macroCount, 2),
		indicator("replay", "GUARDRAIL", "COHERENCE", "compare-normalized-replay", runsKnown, replay, boolInt(replay), 1),
		indicator("writes", "GUARDRAIL", "REGRESSION", "preserve-repository-boundary", known, observation.RepositoryWrites+observation.ExternalRepositoryWrites == 0, observation.RepositoryWrites+observation.ExternalRepositoryWrites, 0),
		indicator("parent", "GUARDRAIL", "REGRESSION", "preserve-parent-failure", parentKnown, parentExact, boolInt(parentExact), 1),
	}
}

func countRuns(runs []CapabilityRun, predicate func(CapabilityRun) bool) int {
	count := 0
	for _, run := range runs {
		if predicate(run) {
			count++
		}
	}
	return count
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
