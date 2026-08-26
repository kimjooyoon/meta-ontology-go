package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func indicatorDecisionLedgerFixture() ([]sourcepolicy.Indicator, []Action) {
	exempt := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityNotApplicable,
		Satisfied:     true,
		Proof:         sourcepolicy.ProofFoundation,
	}
	conforming := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Satisfied:     true,
		Proof:         sourcepolicy.ProofCoherence,
	}
	repair := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Blocking:      true,
		Proof:         sourcepolicy.ProofRegression,
	}
	action := Action{
		IndicatorID:         indicatorID(repair),
		MetricID:            repair.MetricID,
		Subject:             repair.Subject,
		SubjectKind:         repair.SubjectKind,
		Applicability:       repair.Applicability,
		ApplicabilityRule:   repair.ApplicabilityRule,
		ApplicabilityReason: repair.ApplicabilityReason,
		Blocking:            repair.Blocking,
		SourceIndicator:     repair,
		IndicatorOutcome:    repair.Outcome(),
		MetricProofChoice:   repair.Proof,
		MetricProducer:      repair.Producer,
		MetricConsumer:      repair.Consumer,
		Operation:           repair.Operation,
	}
	return []sourcepolicy.Indicator{repair, exempt, conforming}, []Action{action}
}
