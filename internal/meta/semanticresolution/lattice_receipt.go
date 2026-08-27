package semanticresolution

import (
	"fmt"
	"reflect"
	"strings"
)

func BuildLatticeReceipt(source, sourceSHA256 string) LatticeReceipt {
	cases := CanonicalLatticeCases()
	counts := LatticeCounts{CasesTotal: len(cases)}
	for _, item := range cases {
		switch item.Decision {
		case DecisionPass:
			counts.Pass++
		case DecisionFailClosed:
			counts.FailClosed++
		case DecisionUnknown:
			counts.Unknown++
		}
	}
	return LatticeReceipt{
		Schema: LatticeSchema, Source: source, SourceSHA256: sourceSHA256,
		RepositoryWrites: 0, MutationAuthority: false, CaseDenominator: LatticeCaseDenominator,
		Counts: counts, Cases: cases, Claims: CanonicalClaims(), Metrics: canonicalLatticeMetrics(),
	}
}

func canonicalLatticeMetrics() []LatticeMetric {
	return []LatticeMetric{
		metric("gooo.metric.meta-resolution-lattice.exact-observation.count.v1", "outcome", 1, "cases", "greater_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "observe-exact-or-partial", ProofLevelFoundation),
		metric("gooo.metric.meta-resolution-lattice.invariant-descent.count.v1", "driver", 1, "cases", "greater_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "lower-to-invariant-only", ProofLevelCoherence),
		metric("gooo.metric.meta-resolution-lattice.claim-preservation.count.v1", "driver", 4, "cases", "greater_or_equal", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "preserve-claim-state", ProofLevelRegression),
		metric("gooo.metric.meta-resolution-lattice.replay.count.v1", "driver", 4, "cases", "greater_or_equal", "semanticresolution.ReplayPartialObservation", "semantic-resolution-lattice-judge", "replay-resolution-descent", ProofLevelRegression),
		metric("gooo.metric.meta-resolution-lattice.write-guardrail.v1", "guardrail", 0, "repository_writes", "less_or_equal", "semanticresolution.ResolvePartialObservation", "semantic-resolution-lattice-judge", "preserve-read-only-resolution", ProofLevelFoundation),
	}
}

func metric(id, class string, numerator int, unit, relation, producer, consumer, operation string, proof ProofLevel) LatticeMetric {
	return LatticeMetric{ID: id, Class: class, Numerator: numerator, Denominator: LatticeCaseDenominator, Unit: unit, Relation: relation, Producer: producer, Consumer: consumer, MetaOperation: operation, Proof: proof}
}

func ValidateLatticeReceipt(receipt LatticeReceipt) error {
	if receipt.Schema != LatticeSchema || receipt.CaseDenominator != LatticeCaseDenominator || receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		return fmt.Errorf("lattice receipt identity or effect guardrail is invalid")
	}
	if len(receipt.Cases) != LatticeCaseDenominator || receipt.Counts.CasesTotal != LatticeCaseDenominator {
		return fmt.Errorf("lattice receipt denominator is not fixed")
	}
	if receipt.Counts.Pass != 1 || receipt.Counts.FailClosed != 2 || receipt.Counts.Unknown != 1 {
		return fmt.Errorf("lattice receipt decision counts are invalid")
	}
	if err := validateLatticeCases(receipt.Cases); err != nil {
		return err
	}
	if err := validateClaims(receipt.Claims); err != nil {
		return err
	}
	return validateMetrics(receipt.Metrics)
}

func validateLatticeCases(cases []LatticeCase) error {
	for _, item := range cases {
		if item.ID == "" || item.ClaimID == "" || item.Transition.FromResolution != ResolutionExactOperation {
			return fmt.Errorf("lattice case identity is invalid")
		}
		if item.Decision == DecisionUnknown {
			unknown := item.Transition.Unknown
			if item.Transition.Decision != DecisionLowerResolution || item.Transition.ToResolution != ResolutionInvariantOnly || unknown == nil || unknown.Stage != StagePartialObservation || unknown.Step != 1 || unknown.Reason == "" {
				return fmt.Errorf("unknown case did not carry a deterministic descent")
			}
		}
		if item.Decision == DecisionPass && item.Transition.Decision != DecisionPass {
			return fmt.Errorf("pass case is not exact")
		}
		if item.Decision == DecisionFailClosed && item.Transition.Decision != DecisionFailClosed {
			return fmt.Errorf("fail-closed case has an open transition")
		}
		if !reflect.DeepEqual(ReplayPartialObservation(item.Observation), item.Transition) {
			return fmt.Errorf("lattice transition is not replayable")
		}
	}
	return nil
}

func validateClaims(claims []ClaimRecord) error {
	if len(claims) != LatticeCaseDenominator {
		return fmt.Errorf("claim ledger denominator is not fixed")
	}
	for _, item := range claims {
		if item.ID == "" || !validClaimState(item.State) || item.State != item.BeforeState || item.State != item.AfterState || !item.Preserved {
			return fmt.Errorf("claim state was not preserved")
		}
	}
	return nil
}

func validateMetrics(metrics []LatticeMetric) error {
	if len(metrics) != 5 {
		return fmt.Errorf("lattice metric cardinality is invalid")
	}
	proofs := map[ProofLevel]bool{}
	seen := map[string]bool{}
	for _, item := range metrics {
		if item.ID == "" || seen[item.ID] || item.Denominator != LatticeCaseDenominator || item.Numerator < 0 || item.Numerator > item.Denominator || item.Producer == "" || item.Consumer == "" || item.MetaOperation == "" {
			return fmt.Errorf("lattice metric binding is invalid")
		}
		if item.Proof != ProofLevelFoundation && item.Proof != ProofLevelCoherence && item.Proof != ProofLevelRegression {
			return fmt.Errorf("lattice metric proof level is invalid")
		}
		seen[item.ID], proofs[item.Proof] = true, true
	}
	if len(proofs) != 3 || !strings.Contains(metrics[0].ID, "exact-observation") {
		return fmt.Errorf("lattice proof trilemma is incomplete")
	}
	return nil
}

func validClaimState(state ClaimState) bool {
	return state == ClaimOpen || state == ClaimDischarged || state == ClaimRefuted
}
