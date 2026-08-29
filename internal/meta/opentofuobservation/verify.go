package opentofuobservation

import "fmt"

func ValidateReport(report Report, expectedSubject, expectedContract string) error {
	if report.Schema != ReportSchema || report.ContractID != expectedContract || report.SubjectSHA != expectedSubject || report.MetaOperation != MetaOperation {
		return fmt.Errorf("report context is invalid")
	}
	if len(report.UserPaths) != len(fixedPaths) {
		return fmt.Errorf("report user-path denominator is invalid")
	}
	for index, path := range fixedPaths {
		if report.UserPaths[index] != path {
			return fmt.Errorf("report user-path identity is invalid")
		}
	}
	if err := ValidateObservation(observationFromReport(report)); err != nil {
		return fmt.Errorf("report evidence is invalid: %w", err)
	}
	if len(report.Cells) != len(fixedCells) || report.Summary.CellsTotal != len(fixedCells) {
		return fmt.Errorf("report cell denominator is invalid")
	}
	seen := map[string]bool{}
	for index, cell := range report.Cells {
		spec := fixedCells[index]
		if cell.ID != spec.ID || cell.MetaOperation != spec.MetaOperation || cell.ProofChoice != spec.ProofChoice || cell.Indicator != spec.Indicator || seen[cell.ID] || !validDigest(cell.EvidenceDigest) {
			return fmt.Errorf("report cell identity is invalid")
		}
		if report.CellEvidenceDigests[cell.ID] != cell.EvidenceDigest {
			return fmt.Errorf("report cell evidence binding is invalid")
		}
		seen[cell.ID] = true
		if cell.Decision == DecisionPass && cell.State != "CLOSED" {
			return fmt.Errorf("closed cell state is invalid")
		}
	}
	if report.Summary.FoundationClosed != 4 || report.Summary.CoherenceClosed != 4 || report.Summary.RegressionClosed != 4 {
		return fmt.Errorf("proof denominator is invalid")
	}
	if !validCellEvidenceDigests(report.CellEvidenceProjections, report.CellEvidenceDigests) {
		return fmt.Errorf("cell evidence denominator is invalid")
	}
	if report.Summary.ThreePaths != 3 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.PromotionAuthorized || report.ReleaseBinaryBuilds != 0 || report.ReleaseBinaryBuildReason != "NOT_EXECUTED_RELEASE_BINARY_BOUNDARY" {
		return fmt.Errorf("boundary fields are invalid")
	}
	if len(report.Executions) != 2 {
		return fmt.Errorf("report execution denominator is invalid")
	}
	if err := validateExecutions(report.Executions, report.Executions[0].FixtureDigest); err != nil {
		return fmt.Errorf("report executions are invalid: %w", err)
	}
	if err := validateReuse(Observation{Reuse: report.Reuse}); err != nil {
		return fmt.Errorf("report reuse is invalid: %w", err)
	}
	if err := validateGraph(report.Graph); err != nil {
		return fmt.Errorf("report graph is invalid: %w", err)
	}
	if err := validateRuntime(Observation{Runtime: report.Runtime}); err != nil {
		return fmt.Errorf("report runtime is invalid: %w", err)
	}
	expected, err := sealedReportDigest(report)
	if err != nil || expected != report.ReportDigest || !validDigest(report.ReportDigest) {
		return fmt.Errorf("report digest is not sealed")
	}
	return validateDecision(report)
}

func observationFromReport(report Report) Observation {
	return Observation{Schema: ObservationSchema, ContractID: report.ContractID, SubjectSHA: report.SubjectSHA,
		UserPaths: report.UserPaths, Release: report.Release, FixtureDigest: report.FixtureDigest,
		FixtureFiles: report.FixtureFiles, FixturePhysicalLines: report.FixturePhysicalLines,
		Executions: report.Executions, Reuse: report.Reuse, Runtime: report.Runtime, Inventory: report.Inventory,
		ObserverGoVersion: report.ObserverGoVersion, ObserverGOVERSION: report.ObserverGOVERSION,
		ObserverToolchainDigest: report.ObserverToolchainDigest, CellEvidenceProjections: report.CellEvidenceProjections,
		CellEvidenceDigests: report.CellEvidenceDigests, Graph: report.Graph, RepositoryWrites: report.RepositoryWrites,
		LocalTestExecutions: report.LocalTestExecutions, ReleaseBinaryBuilds: report.ReleaseBinaryBuilds,
		ReleaseBinaryBuildReason: report.ReleaseBinaryBuildReason, HumanReportReady: report.HumanReportReady}
}

func validateDecision(report Report) error {
	refuted, unknown := 0, 0
	for _, cell := range report.Cells {
		if cell.Decision == DecisionRefuted {
			refuted++
		}
		if cell.Decision == DecisionUnknown {
			unknown++
		}
	}
	if refuted > 0 && report.Decision != DecisionRefuted {
		return fmt.Errorf("REFUTED precedence is missing")
	}
	if refuted == 0 && unknown > 0 && report.Decision != DecisionUnknown {
		return fmt.Errorf("UNKNOWN decision is missing")
	}
	if refuted == 0 && unknown == 0 && report.Decision != DecisionPass {
		return fmt.Errorf("PASS decision is missing")
	}
	return nil
}

func validCellEvidenceDigests(projections, digests map[string]string) bool {
	if len(projections) != len(fixedCells) || len(digests) != len(fixedCells) {
		return false
	}
	seen := map[string]bool{}
	for _, cell := range fixedCells {
		if projections[cell.ID] == "" || digests[cell.ID] != DigestBytes([]byte(projections[cell.ID])) || seen[digests[cell.ID]] {
			return false
		}
		seen[digests[cell.ID]] = true
	}
	return true
}
