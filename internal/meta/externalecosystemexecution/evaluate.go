package externalecosystemexecution

func Evaluate(o *Observation) Report {
	indicators := buildIndicators(o)
	report := Report{Schema: SchemaVersion, ContractVersion: ContractVersion,
		DenominatorVersion: DenominatorVersion, DenominatorDigest: DenominatorDigest(),
		Total: len(indicators), Indicators: indicators}
	for _, item := range indicators {
		if item.Status == "SATISFIED" { report.Completed++ }
		if item.Status == "UNKNOWN" { report.UnknownIndicators++ }
	}
	if report.Total > 0 { report.BasisPoints = report.Completed * 10000 / report.Total }
	if o != nil {
		report.ExternalExecutions, report.Reference = len(o.Runs), o.Reference
		report.RepositoryWrites = repositoryWrites(o)
		report.OfficialMutationCount, report.PromotionCount = o.OfficialMutationCount, o.PromotionCount
	}
	report.Proofs = proofs(o, indicators)
	if report.UnknownIndicators > 0 {
		report.Decision, report.Resolution, report.Reason = DecisionFailClosed, "COARSE", "EXECUTION_EVIDENCE_UNKNOWN"
	} else if report.Completed != report.Total {
		report.Decision, report.Resolution, report.Reason = DecisionFailClosed, "EXACT", "EXECUTION_INVARIANT_VIOLATED"
	} else {
		report.Decision, report.Resolution, report.Reason = DecisionConfirmed, "EXACT", "EXTERNAL_EXECUTION_EXACT"
	}
	return report
}

func writeStatus(o *Observation) (string, string) {
	states := []RepositoryState{o.SourceBefore, o.SourceAfter, o.ExternalBefore, o.ExternalAfter}
	for _, state := range states { if !state.Available { return "UNKNOWN", "REPOSITORY_STATE_UNAVAILABLE" } }
	if o.SourceBefore.Dirty || o.SourceAfter.Dirty || o.ExternalBefore.Dirty || o.ExternalAfter.Dirty ||
		o.SourceBefore.Commit != o.SourceAfter.Commit || o.SourceBefore.Tree != o.SourceAfter.Tree ||
		o.ExternalBefore.Commit != o.ExternalAfter.Commit || o.ExternalBefore.Tree != o.ExternalAfter.Tree ||
		o.OfficialMutationCount != 0 || o.PromotionCount != 0 { return "UNSATISFIED", "REPOSITORY_WRITE_BOUNDARY_VIOLATED" }
	return "SATISFIED", "REPOSITORY_WRITE_BOUNDARY_EXACT"
}

func repositoryWrites(o *Observation) int {
	status, _ := writeStatus(o); if status == "SATISFIED" { return 0 }; return 1
}

func proofs(o *Observation, indicators []Indicator) []Proof {
	result := []Proof{{Mode:"FOUNDATION",Status:groupStatus(indicators[:4]),Reason:"PINNED_REFERENCE_AND_TOOLCHAIN"},
		{Mode:"COHERENCE",Status:groupStatus(indicators[4:7]),Reason:"TWO_EXECUTIONS_AND_NORMALIZED_REPLAY"}}
	status := "UNKNOWN"
	if o != nil && o.Regression.Total > 0 { status = "UNSATISFIED"; if o.Regression.Passed == o.Regression.Total && o.Regression.Unresolved == 0 { status = "SATISFIED" } }
	return append(result, Proof{Mode:"REGRESSION",Status:status,Reason:"FIXED_NEGATIVE_SUITE"})
}

func groupStatus(items []Indicator) string {
	status := "SATISFIED"
	for _, item := range items { if item.Status == "UNKNOWN" { return "UNKNOWN" }; if item.Status != "SATISFIED" { status = "UNSATISFIED" } }
	return status
}
