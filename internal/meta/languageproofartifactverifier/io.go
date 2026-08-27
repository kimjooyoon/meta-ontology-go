package languageproofartifactverifier

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func WriteReport(path string, report Report) error {
	if err := Validate(report); err != nil {
		if ValidatePreliminary(report) == nil {
			raw, marshalErr := json.MarshalIndent(report, "", "  ")
			if marshalErr != nil {
				return marshalErr
			}
			return os.WriteFile(path, append(raw, '\n'), 0o644)
		}
		caseReasons := make([]string, 0, len(report.Cases))
		for _, item := range report.Cases {
			caseReasons = append(caseReasons, item.ID+"="+item.ObservedReason)
		}
		return fmt.Errorf("validate proof-carrying report: %w (conformance=%s/%s authority=%q cases=%d/%d claims=%d/%d discharged=%d open=%d refuted=%d transitions=%d indicators=%d proofs=%d interventions=%d case_reasons=%s)", err,
			report.ConformanceDecision, report.ConformanceResolution, report.ArtifactUseAuthority,
			report.Summary.CasesSatisfied, report.Summary.CasesTotal, report.Summary.ClaimInstances, report.Summary.ClaimTemplates,
			report.Summary.CaseDischargedClaims, report.Summary.CaseOpenClaims, report.Summary.CaseRefutedClaims,
			len(report.Transitions), len(report.Indicators), len(report.Proofs), len(report.Interventions), strings.Join(caseReasons, ","))
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func LoadReport(path string) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	return decodeStrict[Report](raw)
}
