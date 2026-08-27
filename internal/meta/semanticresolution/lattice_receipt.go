package semanticresolution

import "fmt"

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

func metric(id, class string, numerator int, unit, relation, producer, consumer, operation string, proof ProofLevel) LatticeMetric {
	return LatticeMetric{ID: id, Class: class, Numerator: numerator, Denominator: LatticeCaseDenominator, Unit: unit, Relation: relation, Producer: producer, Consumer: consumer, MetaOperation: operation, Proof: proof}
}

func validateReceiptIdentity(receipt LatticeReceipt) error {
	if receipt.Schema != LatticeSchema || receipt.CaseDenominator != LatticeCaseDenominator || receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		return fmt.Errorf("lattice receipt identity or effect guardrail is invalid")
	}
	if len(receipt.Cases) != LatticeCaseDenominator || receipt.Counts.CasesTotal != LatticeCaseDenominator {
		return fmt.Errorf("lattice receipt denominator is not fixed")
	}
	if receipt.Counts.Pass != 1 || receipt.Counts.FailClosed != 2 || receipt.Counts.Unknown != 1 {
		return fmt.Errorf("lattice receipt decision counts are invalid")
	}
	return nil
}
