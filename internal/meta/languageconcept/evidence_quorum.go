package languageconcept

func evidenceQuorumMetricIDs() []string {
	return []string{
		"gooo.metric.meta.evidence-quorum.fixed-case-coverage.v2",
		"gooo.metric.meta.evidence-quorum.independent-provenance-groups.v2",
		"gooo.metric.meta.evidence-quorum.synthetic-replica-collapse.guardrail.v2",
		"gooo.metric.meta.evidence-quorum.valid-contradiction.guardrail.v2",
		"gooo.metric.meta.evidence-quorum.unknown-open-state.guardrail.v2",
		"gooo.metric.meta.evidence-quorum.comment-invariance.v2",
		"gooo.metric.meta.evidence-quorum.claim-transitions.v2",
		"gooo.metric.meta.evidence-quorum.read-only-observation.guardrail.v2",
		"gooo.metric.meta.evidence-quorum.threshold-intervention.v2",
	}
}

var evidenceQuorumConcept = Concept{
	ID:             "deterministic-evidence-quorum",
	Problem:        "Confidence averages and duplicate receipts can make correlated evidence look independent.",
	PositiveEffect: "A bounded Gooo claim is justified only by distinct provenance groups, required roles, and a fixed minimum quorum.",
	MetaOperation:  "justify-claim-with-independent-evidence",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/meta/evidencequorumconsumer",
		"internal/meta/evidencequorumpolicy",
		"internal/meta/evidencequorumwire",
		"internal/meta/evidencequorumchannel",
		"cmd/evidence-quorum-witness",
		"cmd/evidence-quorum-source-channel",
		"cmd/evidence-quorum-reconstructor",
		"cmd/evidence-quorum-artifact-observer",
		"cmd/evidence-quorum-counterexample",
		"examples/evidence-quorum",
		"scripts/evidence-quorum",
	},
	MetricBindings: evidenceQuorumMetricIDs(),
	UseCases: []UseCase{
		{ID: "current-quorum", Trigger: "three structured provenance lineages support one exact-head claim", ExpectedOutcome: "PASS_3_OF_3_CURRENT_GROUPS"},
		{ID: "synthetic-duplicate", Trigger: "a resealed replica shares executable and dependency lineage", ExpectedOutcome: "PASS_WITH_ONE_COLLAPSED_REPLICA"},
		{ID: "synthetic-valid-conflict", Trigger: "a counterexample satisfies the policy contradiction predicate", ExpectedOutcome: "REFUTED_VALID_CONTRADICTION"},
		{ID: "synthetic-invalid-conflict", Trigger: "a contradiction lacks the policy predicate", ExpectedOutcome: "OPEN_FAIL_CLOSED"},
		{ID: "insufficient-current", Trigger: "one current provenance channel is absent", ExpectedOutcome: "OPEN_LOWER_RESOLUTION"},
		{ID: "synthetic-unknown", Trigger: "an unknown counterexample is observed", ExpectedOutcome: "OPEN_UNKNOWN_LOWER_RESOLUTION"},
	},
}
