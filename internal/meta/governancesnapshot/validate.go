package governancesnapshot

import (
	"fmt"
	"reflect"
)

func ValidateReport(report Report, contract Contract, graph RawGraph, expectedHead string) error {
	if err := ValidateContract(contract); err != nil {
		return err
	}
	if report.Schema != ReportSchema || report.ContractID != contract.ID || report.HeadSHA != expectedHead || report.Repository != contract.Expected.Repository {
		return fmt.Errorf("governance report identity is invalid")
	}
	if len(report.Cells) != len(contract.Cells) || report.Summary != summarize(report.Cells) {
		return fmt.Errorf("governance cell summary is invalid")
	}
	if report.SettingsHealthy != (report.Decision == DecisionClosed) || report.PromotionAuthorized || report.RepositoryWrites != 0 || report.BranchSettingWrites != 0 || report.LocalTestExecutions != 0 || report.CrossProjectGates != 0 || report.Improvement != "UNKNOWN" {
		return fmt.Errorf("governance authority boundary is invalid")
	}
	if err := validateCells(report.Cells, contract.Cells); err != nil {
		return err
	}
	if err := validateSource(report.Source, contract.Source, report.Repository); err != nil {
		return err
	}
	if err := validateGraph(report.Graph, graph, contract); err != nil {
		return err
	}
	if err := validateTerminal(report); err != nil {
		return err
	}
	if err := validateCases(report.Cases, report); err != nil {
		return err
	}
	if report.ReportDigest != sealReport(report) || report.Replay.InputDigest == "" || report.Replay.ProjectionDigest == "" || !report.Replay.ReplayEqual {
		return fmt.Errorf("governance replay evidence is invalid")
	}
	return nil
}

func validateCells(cells []CellObservation, specs []CellSpec) error {
	for index, cell := range cells {
		spec := specs[index]
		if cell.ID != spec.ID || cell.MetaOperation != spec.MetaOperation || cell.ProofChoice != spec.ProofChoice || cell.Indicator != spec.Indicator || cell.Activity != spec.Activity || cell.InputID != spec.InputID || cell.OutputID != spec.OutputID || cell.EvidenceDigest == "" {
			return fmt.Errorf("cell %s is not bound to the contract", spec.ID)
		}
		switch cell.Decision {
		case DecisionClosed:
			if cell.Resolution != ResolutionExact || cell.Unknown != nil || cell.Reason != "" {
				return fmt.Errorf("closed cell %s is malformed", cell.ID)
			}
		case DecisionRefuted:
			if cell.Resolution != ResolutionExact || cell.Reason == "" || cell.Counterexample == "" || cell.Unknown != nil {
				return fmt.Errorf("refuted cell %s is malformed", cell.ID)
			}
		case DecisionUnknown:
			if cell.Resolution != ResolutionLower || !validUnknown(cell.Unknown) || cell.Reason != cell.Unknown.Reason {
				return fmt.Errorf("unknown cell %s is malformed", cell.ID)
			}
		default:
			return fmt.Errorf("cell %s has unknown decision", cell.ID)
		}
	}
	return nil
}

func validateSource(observed SourceEvidence, expected SourceContract, repository string) error {
	if !reflect.DeepEqual(observed.Documentation, expected.Documentation) || !reflect.DeepEqual(observed.APIVersions, expected.APIVersions) || len(observed.Requests) != len(expected.Endpoints) || len(observed.Payloads) != len(expected.Endpoints) {
		return fmt.Errorf("source authority is not preserved")
	}
	for index, endpoint := range expected.Endpoints {
		request := observed.Requests[index]
		if request.ID != endpoint.ID || request.Method != endpoint.Method || request.URL != endpoint.Path || request.APIVersion != endpoint.API || request.PayloadPath != endpoint.Payload {
			return fmt.Errorf("source request %s is not canonical", endpoint.ID)
		}
	}
	_ = repository
	return nil
}

func validateGraph(observed GraphEvidence, raw RawGraph, contract Contract) error {
	expected, reason := graphEvidence(raw, contract)
	if reason != "" || !reflect.DeepEqual(observed, expected) || observed.NodeCount != 36 || observed.RelationCount != 24 || observed.ActivityCount != 12 || observed.BindingCount != 12 {
		return fmt.Errorf("Gooo graph binding is invalid")
	}
	return nil
}

func validateTerminal(report Report) error {
	if report.Decision == DecisionRefuted && report.Resolution != ResolutionExact || report.Decision == DecisionUnknown && report.Resolution != ResolutionLower || report.Decision == DecisionClosed && report.Resolution != ResolutionExact {
		return fmt.Errorf("governance terminal resolution is invalid")
	}
	if report.Summary.RefutedCells > 0 && report.Decision != DecisionRefuted {
		return fmt.Errorf("refuted cell was not promoted to top-level decision")
	}
	if report.Summary.RefutedCells == 0 && report.Summary.UnknownCells > 0 && report.Decision != DecisionUnknown {
		return fmt.Errorf("unknown cell was not preserved")
	}
	if report.Summary.RefutedCells == 0 && report.Summary.UnknownCells == 0 && report.Decision != DecisionClosed {
		return fmt.Errorf("complete governance snapshot was not closed")
	}
	return nil
}

func validUnknown(unknown *Unknown) bool {
	if unknown == nil || unknown.Stage == "" || unknown.Step == "" || unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" || unknown.BlockedBy == nil {
		return false
	}
	if unknown.UnknownClass == "DIRECT_MISSING" {
		return len(unknown.BlockedBy) == 0
	}
	return unknown.UnknownClass == "DEPENDENCY_BLOCKED" && len(unknown.BlockedBy) > 0
}

func validateCases(cases []CanonicalCase, report Report) error {
	if len(cases) != 6 {
		return fmt.Errorf("canonical case denominator is invalid")
	}
	want := []string{"normal-main-match", "current-dev-drift", "missing-public-snapshot", "dependency-blocked-context-comparison", "malformed-payload", "disabled-ruleset-authority"}
	for index, item := range cases {
		if item.ID != want[index] {
			return fmt.Errorf("canonical case order is invalid")
		}
		if item.Decision == DecisionUnknown && !validUnknown(item.Unknown) {
			return fmt.Errorf("canonical case %s has invalid unknown", item.ID)
		}
		if item.Decision == DecisionRefuted && (item.Resolution != ResolutionExact || item.Reason == "" || item.Counterexample == "") {
			return fmt.Errorf("canonical case %s has invalid refutation", item.ID)
		}
	}
	if cases[0].Decision != DecisionClosed || cases[1].Decision != report.Decision || cases[2].Decision != DecisionUnknown || cases[3].Decision != DecisionUnknown || cases[4].Decision != DecisionRefuted || cases[5].Decision != DecisionRefuted {
		return fmt.Errorf("canonical case decisions are invalid")
	}
	if report.Decision == DecisionRefuted && (cases[1].Resolution != ResolutionExact || cases[1].Reason != "DEV_LIVE_PROTECTION_DRIFT") {
		return fmt.Errorf("current drift case is not canonical")
	}
	if report.Decision == DecisionClosed && (cases[1].Resolution != ResolutionExact || cases[1].Reason != "GOVERNANCE_SNAPSHOT_CONFORMANT") {
		return fmt.Errorf("current conformant case is not canonical")
	}
	if cases[2].Unknown.UnknownClass != "DIRECT_MISSING" || len(cases[2].Unknown.BlockedBy) != 0 || cases[3].Unknown.UnknownClass != "DEPENDENCY_BLOCKED" {
		return fmt.Errorf("canonical case unknown frontier is invalid")
	}
	if cases[4].Reason != "MALFORMED_PUBLIC_PAYLOAD" || cases[5].Reason != "DISABLED_RULESET_AUTHORITY" {
		return fmt.Errorf("canonical refutation reasons are invalid")
	}
	return nil
}
