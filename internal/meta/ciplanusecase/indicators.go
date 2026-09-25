package ciplanusecase

func buildIndicators(summary Summary, limits Limits) []Indicator {
	return []Indicator{
		equal("gooo.metric.ci-plan.cases-satisfied.v1", "USER", int64(summary.CasesSatisfied), 12, "ciplanusecase.Evaluate", "human-scorecard", "close-fixed-usecase"),
		equal("gooo.metric.ci-plan.pass-decisions.v1", "USER", int64(summary.PassDecisions), 4, "metainvocation.Invoke", "ciplanusecase.Evaluate", "classify-ci-plan-decision"),
		equal("gooo.metric.ci-plan.fail-closed-decisions.v1", "USER", int64(summary.FailClosedDecisions), 4, "metainvocation.Invoke", "ciplanusecase.Evaluate", "reject-invalid-change-set"),
		equal("gooo.metric.ci-plan.unknown-decisions.v1", "USER", int64(summary.UnknownDecisions), 4, "metainvocation.Invoke", "ciplanusecase.Evaluate", "lower-rule-resolution"),
		equal("gooo.metric.ci-plan.deterministic-replays.v1", "TOOL_AUTHOR", int64(summary.DeterministicReplays), 12, "ci-plan-usecase", "ciplanusecase.Evaluate", "replay-meta-invocation"),
		equal("gooo.metric.ci-plan.golden-plans.v1", "TOOL_AUTHOR", int64(summary.GoldenPlans), 4, "ciplanusecase.ProjectGolden", "ciplanusecase.Evaluate", "compare-generated-plan"),
		equal("gooo.metric.ci-plan.rule-evidence-refs.v1", "TOOL_AUTHOR", int64(summary.RuleEvidenceRefs), 6, "metainvocation.selectChecks", "ciplanusecase.Evaluate", "bind-rule-evidence"),
		equal("gooo.metric.ci-plan.direct-unknown-claims.v1", "GOVERNOR", int64(summary.DirectUnknownClaims), 4, "metainvocation.claimsFor", "ciplanusecase.Evaluate", "locate-unknown-causality"),
		equal("gooo.metric.ci-plan.dependency-blocked-claims.v1", "GOVERNOR", int64(summary.DependencyBlocked), 8, "metainvocation.claimsFor", "ciplanusecase.Evaluate", "propagate-claim-dependency"),
		equal("gooo.metric.ci-plan.refuted-claims.v1", "GOVERNOR", int64(summary.RefutedClaims), 4, "metainvocation.claimsFor", "ciplanusecase.Evaluate", "retain-refuted-claim"),
		equal("gooo.metric.ci-plan.persistent-claims.v1", "GOVERNOR", int64(summary.PersistentClaims), 36, "metainvocation.claimsFor", "ciplanusecase.Evaluate", "retain-claim-ledger"),
		equal("gooo.metric.ci-plan.generated-source-replays.v1", "TOOL_AUTHOR", int64(summary.GeneratedReplays), 1, "gooo.generate", "ciplanusecase.Evaluate", "replay-generated-go"),
		equal("gooo.metric.ci-plan.resource-samples.v1", "USER", int64(summary.ResourceSamples), 12, "ci-plan-usecase", "human-scorecard", "measure-user-invocation"),
		equal("gooo.metric.ci-plan.gooo-files.v1", "TOOL_AUTHOR", int64(summary.GoooFiles), 1, "ci-plan-scorecard", "human-scorecard", "count-language-source"),
		equal("gooo.metric.ci-plan.go-files.v1", "TOOL_AUTHOR", int64(summary.GoFiles), 0, "ci-plan-scorecard", "human-scorecard", "count-generated-host-source"),
		equal("gooo.metric.ci-plan.gooo-lines.v1", "TOOL_AUTHOR", int64(summary.GoooLines), 12, "ci-plan-scorecard", "human-scorecard", "count-language-lines"),
		equal("gooo.metric.ci-plan.go-lines.v1", "TOOL_AUTHOR", int64(summary.GoLines), 0, "ci-plan-scorecard", "human-scorecard", "count-generated-host-lines"),
		atMost("gooo.metric.ci-plan.max-wall-ms.v1", "USER", summary.MaxWallMS, limits.MaxWallMS, "ci-plan-usecase", "human-scorecard", "bound-user-latency"),
		atMost("gooo.metric.ci-plan.peak-rss-kib.v1", "USER", summary.MaxPeakRSSKiB, limits.MaxPeakRSSKiB, "ci-plan-usecase", "human-scorecard", "bound-user-memory"),
		atMost("gooo.metric.ci-plan.max-receipt-bytes.v1", "USER", summary.MaxReceiptBytes, limits.MaxReceiptBytes, "ci-plan-usecase", "human-scorecard", "bound-receipt-size"),
		equal("gooo.metric.ci-plan.repository-writes.guardrail.v1", "GOVERNOR", int64(summary.RepositoryWrites), 0, "metainvocation.Invoke", "governance", "preserve-zero-effect-boundary"),
		equal("gooo.metric.ci-plan.mutation-authority.guardrail.v1", "GOVERNOR", int64(summary.MutationAuthority), 0, "metainvocation.Invoke", "governance", "deny-mutation-authority"),
	}
}

func equal(id, reader string, observed, target int64, producer, consumer, operation string) Indicator {
	return Indicator{ID: id, Reader: reader, Observed: observed, Comparator: "EQ", Target: target, Status: status(observed == target), Producer: producer, Consumer: consumer, MetaOperation: operation}
}

func atMost(id, reader string, observed, target int64, producer, consumer, operation string) Indicator {
	return Indicator{ID: id, Reader: reader, Observed: observed, Comparator: "LTE", Target: target, Status: status(observed <= target), Producer: producer, Consumer: consumer, MetaOperation: operation}
}

func buildReaderViews(indicators []Indicator) []ReaderView {
	views := []ReaderView{{Reader: "USER", Resolution: "EXACT"}, {Reader: "TOOL_AUTHOR", Resolution: "EXACT"}, {Reader: "GOVERNOR", Resolution: "EXACT"}}
	for _, indicator := range indicators {
		for index := range views {
			if views[index].Reader == indicator.Reader {
				views[index].IndicatorIDs = append(views[index].IndicatorIDs, indicator.ID)
			}
		}
	}
	return views
}
