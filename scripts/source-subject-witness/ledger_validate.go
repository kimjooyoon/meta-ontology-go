package main

import "fmt"

func validateLedger(ledger witnessLedger) error {
	if ledger.Schema != "gooo/source-subject-witness-ledger/v3" || ledger.Status != "BOUND" {
		return fmt.Errorf("ledger schema or status is not bound")
	}
	if !ledger.RootTopologyExempt || !ledger.RootREADMEExempt || len(ledger.Witnesses) != ledger.Counts.SubjectWitnesses {
		return fmt.Errorf("ledger root policy or witness count is invalid")
	}
	if ledger.Counts.WorkflowDiscoveryExemptions != 1 {
		return fmt.Errorf("ledger has %d workflow discovery exemptions, want 1", ledger.Counts.WorkflowDiscoveryExemptions)
	}
	if ledger.Counts.FunctionSourceBindings != ledger.Counts.FunctionWitnesses || ledger.Counts.RootSummaryIndicators != rootSummaryCount {
		return fmt.Errorf("ledger function or root summary meta coverage is incomplete")
	}
	if ledger.Counts.SourceIndicatorsApplicable+ledger.Counts.SourceIndicatorsNotApplicable != ledger.Counts.MetaIndicators {
		return fmt.Errorf("ledger source indicator applicability partition is incomplete")
	}
	previous, rootSummaries := "", 0
	for _, witness := range ledger.Witnesses {
		if witness.Space == rootSummarySpace {
			rootSummaries++
			if witness.Path != "." || witness.SubjectKind != "PROJECT_ROOT" || witness.Meta.IndicatorCount != rootSummaryCount {
				return fmt.Errorf("logical source summary witness is outside the exact catalog")
			}
		}
		key := witnessKey(witness)
		if previous != "" && key <= previous {
			return fmt.Errorf("witness order is not canonical at %q", key)
		}
		previous = key
		if sealWitness(witness).WitnessDigest != witness.WitnessDigest {
			return fmt.Errorf("witness digest mismatch at %q", key)
		}
	}
	if rootSummaries != 1 {
		return fmt.Errorf("ledger has %d logical source summary witnesses, want 1", rootSummaries)
	}
	counts := countWitnesses(ledger.Witnesses)
	counts.MetaIndicators = ledger.Counts.MetaIndicators
	if counts != ledger.Counts || digestValues(ledger.Witnesses) != ledger.SubjectWitnessDigest {
		return fmt.Errorf("ledger counts or subject digest mismatch")
	}
	if digestValues(ledger.Indicators) != ledger.IndicatorDigest {
		return fmt.Errorf("ledger indicator digest mismatch")
	}
	if ledger.SourceObservationDigest != digestValues(ledger.SourceObservations) {
		return fmt.Errorf("ledger source observation digest mismatch")
	}
	if len(ledger.SourceObservations) == 0 {
		if ledger.ObservationState != "" || ledger.ClaimState != "" {
			return fmt.Errorf("empty source observations have a non-empty state")
		}
	} else {
		if ledger.ObservationState != "OBSERVED" || ledger.ClaimState != "OPEN" {
			return fmt.Errorf("unsatisfied source observations are not OPEN/OBSERVED")
		}
		seen := make(map[string]bool, len(ledger.SourceObservations))
		for _, observation := range ledger.SourceObservations {
			if observation.Satisfied || observation.Value <= observation.Limit {
				return fmt.Errorf("source observation %q is not an unsatisfied driver", observation.Subject)
			}
			if err := validateLineCapIndicator(observation); err != nil {
				return err
			}
			key := observation.MetricID + "\x00" + observation.Subject
			if seen[key] {
				return fmt.Errorf("source observation %q is duplicated", observation.Subject)
			}
			seen[key] = true
		}
	}
	for _, indicator := range ledger.Indicators {
		if indicator.Verdict != "PASS" {
			return fmt.Errorf("ledger indicator %q did not pass", indicator.ID)
		}
	}
	if ledgerSemanticDigest(ledger) != ledger.SemanticDigest {
		return fmt.Errorf("ledger semantic digest mismatch")
	}
	return nil
}
