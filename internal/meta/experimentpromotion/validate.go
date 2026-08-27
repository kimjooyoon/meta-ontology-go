package experimentpromotion

import (
	"fmt"
	"reflect"
)

func ValidateReport(report Report) error {
	if report.Schema != ReportSchema || report.Scope != PortfolioScope {
		return fmt.Errorf("REPORT_IDENTITY_INVALID")
	}
	if !validDigest(report.ObservationDigest) || !validDigest(report.Digest) {
		return fmt.Errorf("REPORT_DIGEST_INVALID")
	}
	if !report.SourceProjection.Exact || report.SourceProjection.Path != SourcePath || !validDigest(report.SourceProjection.RawDigest) || !validDigest(report.SourceProjection.SemanticDigest) || !sameStrings(report.SourceProjection.Experiments, experimentIDs()) || !sameStrings(report.SourceProjection.Gates, GateIDs) {
		return fmt.Errorf("REPORT_SOURCE_PROJECTION_INVALID")
	}
	if len(report.Experiments) != ExperimentCount || len(report.ClaimLedger) != GateSlotCount || len(report.EmittedClaims) != GateSlotCount {
		return fmt.Errorf("REPORT_CARDINALITY_INVALID")
	}
	if report.AggregateMetrics == nil || len(report.AggregateMetrics) != 0 || report.MutationAuthority || report.RepositoryWrites != report.RepositorySnapshot.ChangedPaths {
		return fmt.Errorf("REPORT_GUARDRAIL_INPUT_INVALID")
	}
	if !validDigest(report.RepositorySnapshot.BeforeDigest) || !validDigest(report.RepositorySnapshot.AfterDigest) || report.RepositorySnapshot.ChangedPaths < 0 {
		return fmt.Errorf("REPORT_REPOSITORY_SNAPSHOT_INVALID")
	}

	for experimentIndex, experiment := range report.Experiments {
		if experiment.ExperimentID != experimentIDs()[experimentIndex] || len(experiment.Gates) != GateCount {
			return fmt.Errorf("REPORT_EXPERIMENT_ORDER_INVALID: %d", experimentIndex)
		}
		for gateIndex, gate := range experiment.Gates {
			if gate.ExperimentID != experiment.ExperimentID || gate.GateID != GateIDs[gateIndex] || !validStatus(gate.Status) || gate.ClaimTransition.ExperimentID != gate.ExperimentID || gate.ClaimTransition.GateID != gate.GateID || gate.ClaimTransition.From != ClaimOpen || gate.ClaimTransition.To != claimStateFor(gate.Status) || gate.ClaimTransition.Digest != claimTransitionDigest(gate.ExperimentID, gate.GateID, gate.ClaimTransition.To) {
				return fmt.Errorf("REPORT_GATE_INVALID: %s/%s", experiment.ExperimentID, gate.GateID)
			}
			if gate.Status == StatusUnknown && gate.ObservationID != "" {
				return fmt.Errorf("REPORT_UNKNOWN_HAS_OBSERVATION: %s/%s", experiment.ExperimentID, gate.GateID)
			}
		}
		expectedStatus, expectedStage, expectedStep, expectedReason := summarizeExperiment(experiment.Gates)
		if experiment.Status != expectedStatus || experiment.Stage != expectedStage || experiment.Step != expectedStep || experiment.Reason != expectedReason {
			return fmt.Errorf("REPORT_EXPERIMENT_SUMMARY_INVALID: %s", experiment.ExperimentID)
		}
	}

	previous := ""
	sequence := 0
	for _, experiment := range report.Experiments {
		for _, gate := range experiment.Gates {
			entry := report.ClaimLedger[sequence]
			if entry.Sequence != sequence+1 || entry.ExperimentID != gate.ExperimentID || entry.GateID != gate.GateID || entry.PriorState != ClaimOpen || entry.NextState != gate.ClaimTransition.To || entry.Stage != gate.Stage || entry.Step != gate.Step || entry.Reason != gate.Reason || entry.PreviousDigest != previous || entry.Digest != ledgerDigest(entry) {
				return fmt.Errorf("REPORT_CLAIM_LEDGER_INVALID: %d", sequence+1)
			}
			claim := report.EmittedClaims[sequence]
			if claim.ExperimentID != gate.ExperimentID || claim.GateID != gate.GateID || claim.Class != "PROMOTION_GATE" || claim.State != gate.ClaimTransition.To {
				return fmt.Errorf("REPORT_EMITTED_CLAIM_INVALID: %d", sequence+1)
			}
			previous = entry.Digest
			sequence++
		}
	}

	expectedSummary := summarize(report.Experiments, report.ClaimLedger, report.EmittedClaims, report.RepositorySnapshot)
	if !reflect.DeepEqual(report.Summary, expectedSummary) {
		return fmt.Errorf("REPORT_SUMMARY_INVALID")
	}
	expectedGuards := makeGuardrails(report.EmittedClaims, report.RepositorySnapshot)
	if !reflect.DeepEqual(report.Guardrails, expectedGuards) {
		return fmt.Errorf("REPORT_GUARDRAILS_INVALID")
	}
	if report.Digest != reportDigest(report) {
		return fmt.Errorf("REPORT_SELF_DIGEST_INVALID")
	}
	return nil
}

func validStatus(value string) bool {
	switch value {
	case StatusProven, StatusOpen, StatusUnknown, StatusRefuted:
		return true
	default:
		return false
	}
}
