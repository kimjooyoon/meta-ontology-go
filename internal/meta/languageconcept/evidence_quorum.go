package languageconcept

import quorum "github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencequorum"

var evidenceQuorumConcept = Concept{
	ID:             "deterministic-evidence-quorum",
	Problem:        "Confidence averages and duplicate receipts can make correlated evidence look independent.",
	PositiveEffect: "A bounded Gooo claim is justified only by distinct provenance groups, required roles, and a fixed minimum quorum.",
	MetaOperation:  "justify-claim-with-independent-evidence",
	Rarity:         "UNCOMMON_COMBINATION",
	Stage:          "OPERATING",
	NoveltyClaim:   false,
	CodeBindings: []string{
		"internal/meta/evidencequorum",
		"cmd/evidence-quorum-witness",
		"examples/evidence-quorum",
		"scripts/evidence-quorum",
	},
	MetricBindings: quorum.MetricIDs(),
	UseCases: []UseCase{
		{ID: "sufficient-independent", Trigger: "three required roles come from three origin groups", ExpectedOutcome: "PASS_3_OF_3_INDEPENDENT_GROUPS"},
		{ID: "same-origin-replica", Trigger: "a producer repeats the same observation under one origin group", ExpectedOutcome: "FAIL_CLOSED_2_OF_3_INDEPENDENT_GROUPS"},
		{ID: "conflicting-independent", Trigger: "an independent origin group contradicts the claim", ExpectedOutcome: "FAIL_CLOSED_CONFLICT_INVARIANT_ONLY"},
		{ID: "insufficient-independent", Trigger: "fewer than three independent groups are available", ExpectedOutcome: "FAIL_CLOSED_LOWER_RESOLUTION"},
	},
}
