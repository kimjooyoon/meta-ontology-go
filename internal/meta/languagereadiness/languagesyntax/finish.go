package languagesyntax

func finish(report Report) Report {
	summary := Summary{Total: totalCases, UnregisteredGooo: len(report.Source.UnregisteredGooo),
		MissingRegistered: len(report.Source.MissingRegistered)}
	for _, file := range report.Source.GoooFiles {
		summary.GoooLines += file.GoooLines
	}
	for _, item := range report.Cases {
		switch item.Status {
		case "SATISFIED":
			summary.Satisfied++
			if item.Definition.Kind == KindValid {
				summary.ValidCases++
			} else {
				summary.InvalidCases++
			}
		case "UNRESOLVED":
			summary.Unresolved++
		default:
			summary.NotSatisfied++
		}
		summary.ASTReplays += boolInt(item.Evidence.ASTReplayed)
		summary.ByteReplays += boolInt(item.Evidence.ByteReplayed)
		summary.SemanticReplays += boolInt(item.Evidence.SemanticReplayed)
		summary.GetPutLaws += boolInt(item.Evidence.GetPut)
		summary.PutGetLaws += boolInt(item.Evidence.PutGet)
		summary.DiagnosticRejections += boolInt(item.Evidence.DiagnosticRejected)
	}
	summary.Executed = summary.Total - summary.Unresolved
	summary.ReadinessBPS = summary.Satisfied * 10_000 / totalCases
	report.Summary = summary
	registryDrift := boolInt(report.Source.RegistryDigest != registryDigest())
	ready := summary.Satisfied == totalCases && summary.Unresolved == 0 &&
		summary.UnregisteredGooo == 0 && summary.MissingRegistered == 0 &&
		report.RepositoryWrites == 0 && report.Source.ConceptRepositoryWrites == 0 &&
		!report.MutationAuthorized && report.Source.ConceptBound && report.Source.ObservationKnown && registryDrift == 0
	report.Decision, report.Reason, report.Resolution = DecisionClosed, "SYNTAX_ROUNDTRIP_MISMATCH", ResolutionExact
	if summary.Unresolved > 0 || !report.Source.ObservationKnown || !report.Source.ConceptBound {
		report.Reason, report.Resolution = "SYNTAX_ROUNDTRIP_EVIDENCE_UNKNOWN", ResolutionLower
	} else if ready {
		report.Decision, report.Reason = DecisionPass, "LANGUAGE_SYNTAX_ROUNDTRIP_PROVEN"
	}
	report.Indicators = indicators(report, registryDrift)
	report.Proofs = proofs(report, registryDrift)
	return seal(report)
}
