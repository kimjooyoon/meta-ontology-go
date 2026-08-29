package rollbackfixedpoint

func Build(source Source) Report {
	coordinates := Coordinates(source)
	summary := summarize(coordinates, source)
	report := Report{Schema: Schema, Decision: DecisionFailClosed, Reason: ReasonRejected,
		Resolution: ResolutionExact, Producer: "rollbackfixedpoint.Build",
		Consumer: "transformation-effect-ledger", MetaOperation: "recover-guarded-fixed-point",
		Source: source, Summary: summary, Coordinates: coordinates,
		Indicators: indicators(source, summary), Proofs: proofs(coordinates),
		RepositoryWrites: source.RepositoryWrites}
	if mixedTerminal(source) && summary.Unresolved == 0 {
		report.Reason, report.Mode = ReasonMixed, ModeMixedTerminal
	} else if summary.Unresolved > 0 {
		report.Reason, report.Resolution = ReasonUnknown, ResolutionLower
	} else if summary.Satisfied == totalCoordinates {
		report.Decision = DecisionPass
		if recoverable(source.Guard) {
			report.Reason, report.Mode = ReasonRecovered, ModeRecovered
		} else {
			report.Reason, report.Mode = ReasonAuthorized, ModeAuthorized
		}
	}
	return seal(report)
}

func mixedTerminal(source Source) bool {
	t := source.Transformation
	return source.CollectionError == "" && t.Decision == "APPLIED" &&
		t.Reason == "SANDBOX_EFFECTS_VERIFIED" && t.OperationOutcome == "MIXED_CLOSED_REFUTED" &&
		t.ReceiptDecision == "REFUTED" && t.Effects == t.AppliedEffects+t.RefutedEffects &&
		t.AppliedEffects > 0 && t.RefutedEffects > 0 &&
		t.ReceiptCount == t.AppliedEffects && t.FailureCount == t.RefutedEffects &&
		t.DirectUnknownCount == 0 && t.DependencyBlockedUnknownCount == t.UnknownCount &&
		validDigest(t.UnknownCausalDigest) &&
		t.SourceWorkspaceUnchanged && t.WriteBoundary == "SANDBOX_ONLY" &&
		!t.PromotionAuthorized
}

func IsKnownMixedTerminal(report Report) bool {
	return report.Decision == DecisionFailClosed && report.Resolution == ResolutionExact &&
		report.Reason == ReasonMixed && report.Mode == ModeMixedTerminal && mixedTerminal(report.Source)
}

func summarize(values []Coordinate, source Source) Summary {
	result := Summary{Total: len(values), RepositoryWrites: source.RepositoryWrites}
	for _, value := range values {
		switch value.Status {
		case "SATISFIED":
			result.Satisfied++
		case "UNRESOLVED":
			result.Unresolved++
		default:
			result.NotSatisfied++
		}
	}
	result.ReadinessBPS = result.Satisfied * 10000 / result.Total
	if result.Satisfied == totalCoordinates && recoverable(source.Guard) {
		result.RecoveredFixedPoints = 1
	}
	if result.Satisfied == totalCoordinates && authorized(source.Guard) {
		result.AuthorizedPromotions = 1
	}
	return result
}
