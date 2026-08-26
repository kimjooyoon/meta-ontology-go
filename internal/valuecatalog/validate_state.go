package valuecatalog

import "fmt"

func validateCatalogState(report Report) error {
	extension := report.Improvement.After.Satisfied == 1
	if extension {
		if report.Decision != DecisionExtensionProven || report.Reason != ReasonExtensionExact || report.Improvement.After != coordinate(1, 1) {
			return fmt.Errorf("extension decision is invalid")
		}
		if report.Extension.Program != "int.add:2" || report.Extension.Passed != 3 || report.ExtensionCoreProgram != "int.add:2" {
			return fmt.Errorf("extension evidence is not exact")
		}
		return validateCounts(report, 13, []int{3, 10, 13}, []bool{true, true, true})
	}
	if report.Decision != DecisionBaselineObserved || report.Reason != ReasonBaselineObserved || report.Improvement.After != coordinate(0, 1) {
		return fmt.Errorf("baseline observation decision is invalid")
	}
	if report.Extension.CompileReason != "VALUE_PROGRAM_MISSING" || report.Extension.Passed != 0 || report.ExtensionCoreProgram != "" {
		return fmt.Errorf("extension slot is not an explicit baseline")
	}
	return validateCounts(report, 9, []int{1, 7, 9}, []bool{true, false, true})
}

func validateCounts(report Report, indicatorSatisfied int, viewSatisfied []int, proofPassed []bool) error {
	if countSatisfied(report.Indicators) != indicatorSatisfied {
		return fmt.Errorf("indicator satisfied count changed")
	}
	for index, view := range report.Views {
		if view.Satisfied != viewSatisfied[index] || view.BasisPoints != coordinate(view.Satisfied, view.Total).BasisPoints {
			return fmt.Errorf("view %d changed", index)
		}
	}
	for index, proof := range report.Proofs {
		if proof.Passed != proofPassed[index] || !validDigest(proof.EvidenceDigest) {
			return fmt.Errorf("proof %d changed", index)
		}
	}
	return nil
}

func countSatisfied(indicators []Indicator) int {
	count := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			count++
		}
	}
	return count
}
