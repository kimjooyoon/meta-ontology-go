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
	if !report.SourceProjection.Exact || report.SourceProjection.Path != SourcePath || !validDigest(report.SourceProjection.RawDigest) || !validDigest(report.SourceProjection.SemanticDigest) || len(report.SourceProjection.Experiments) != ExperimentCount || !sameStrings(report.SourceProjection.Gates, GateIDs) {
		return fmt.Errorf("REPORT_SOURCE_PROJECTION_INVALID")
	}
	if len(report.Experiments) != ExperimentCount || len(report.ClaimLedger) != GateSlotCount || len(report.EmittedClaims) != GateSlotCount {
		return fmt.Errorf("REPORT_CARDINALITY_INVALID")
	}
	if len(report.AggregateMetrics) != 0 || report.MutationAuthority || report.RepositoryWrites != report.RepositorySnapshot.ChangedPaths || report.Summary.DeclaredExperimentsNumerator != ExperimentCount || report.Summary.DeclaredExperimentsDenominator != ExperimentCount || report.Summary.MaterializedClaimSlotsNumerator != GateSlotCount || report.Summary.MaterializedClaimSlotsDenominator != GateSlotCount {
		return fmt.Errorf("REPORT_GUARDRAIL_INPUT_INVALID")
	}
	if !validSnapshot(report.RepositorySnapshot) {
		return fmt.Errorf("REPORT_REPOSITORY_SNAPSHOT_INVALID")
	}
	for experimentIndex, experiment := range report.Experiments {
		if experiment.ExperimentID != report.SourceProjection.Experiments[experimentIndex].ID || len(experiment.Gates) != GateCount || !validStatus(experiment.Status) || !validStatus(experiment.EvidenceStatus) {
			return fmt.Errorf("REPORT_EXPERIMENT_ORDER_INVALID: %d", experimentIndex)
		}
		for gateIndex, gate := range experiment.Gates {
			if gate.ExperimentID != experiment.ExperimentID || gate.GateID != GateIDs[gateIndex] || !validStatus(gate.Status) || !validStatus(gate.PromotionStatus) || gate.ClaimTransition.ExperimentID != gate.ExperimentID || gate.ClaimTransition.GateID != gate.GateID || gate.ClaimTransition.From != ClaimOpen || gate.ClaimTransition.To != claimStateFor(gate.Status) || gate.ClaimTransition.Digest != claimTransitionDigest(gate.ExperimentID, gate.GateID, gate.ClaimTransition.To) {
				return fmt.Errorf("REPORT_GATE_INVALID: %s/%s", experiment.ExperimentID, gate.GateID)
			}
			if gate.Status == StatusUnknown && gate.ObservationID == "" && gate.Stage == "" {
				return fmt.Errorf("REPORT_UNKNOWN_UNLOCATED: %s/%s", experiment.ExperimentID, gate.GateID)
			}
		}
	}
	previous := report.PriorLedger.LastDigest
	for sequence, entry := range report.ClaimLedger {
		gate := report.Experiments[sequence/GateCount].Gates[sequence%GateCount]
		if entry.Sequence != sequence+1 || entry.ExperimentID != gate.ExperimentID || entry.GateID != gate.GateID || entry.PriorState != ClaimOpen || entry.NextState != gate.ClaimTransition.To || entry.Stage != gate.Stage || entry.Step != gate.Step || entry.Reason != gate.Reason || entry.PreviousDigest != previous || entry.Digest != ledgerDigest(entry) {
			return fmt.Errorf("REPORT_CLAIM_LEDGER_INVALID: %d", sequence+1)
		}
		claim := report.EmittedClaims[sequence]
		if claim.ExperimentID != gate.ExperimentID || claim.GateID != gate.GateID || claim.TargetAddress == "" || claim.State != gate.ClaimTransition.To {
			return fmt.Errorf("REPORT_EMITTED_CLAIM_INVALID: %d", sequence+1)
		}
		previous = entry.Digest
	}
	expectedSummary := summarize(report.Experiments, report.ClaimLedger, report.EmittedClaims, report.RepositorySnapshot, report.Counterexamples)
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

func validSnapshot(snapshot RepositorySnapshot) bool {
	if !validDigest(snapshot.BeforeDigest) || !validDigest(snapshot.AfterDigest) || snapshot.ChangedPaths < 0 || snapshot.ChangedPaths != len(snapshot.ChangedPathList) || snapshot.BeforeDigest != DigestBytes(snapshot.BeforeRaw) || snapshot.AfterDigest != DigestBytes(snapshot.AfterRaw) {
		return false
	}
	seen := make(map[string]bool)
	for _, path := range snapshot.ChangedPathList {
		if path == "" || seen[path] {
			return false
		}
		seen[path] = true
	}
	return true
}

func validStatus(value string) bool {
	switch value {
	case StatusProven, StatusOpen, StatusUnknown, StatusRefuted:
		return true
	default:
		return false
	}
}
