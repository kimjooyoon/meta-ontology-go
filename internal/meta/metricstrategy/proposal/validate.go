package proposal

import (
	"fmt"
	"reflect"
)

func Validate(report Report) error {
	digest, err := registryDigest()
	if err != nil || report.Schema != Schema || report.RegistrySchema != RegistrySchema || report.RegistryDigest != digest || report.Repository == "" || report.SubjectSHA == "" {
		return fmt.Errorf("change proposal report identity is invalid")
	}
	if len(report.Coordinates) != len(registry) {
		return fmt.Errorf("change proposal coordinate denominator changed")
	}
	for index, coordinate := range report.Coordinates {
		if coordinate.CoordinateSpec != registry[index] || coordinate.EvidenceDigest == "" || !knownStatus(coordinate.Status) {
			return fmt.Errorf("change proposal coordinate %d is invalid", index)
		}
	}
	summary := summarize(report.Coordinates)
	decision, reason := decisionFor(summary)
	if report.Summary != summary || report.Decision != decision || report.Reason != reason || report.RepositoryWrites < 0 {
		return fmt.Errorf("change proposal summary diverged")
	}
	if !reflect.DeepEqual(report.Indicators, buildIndicators(summary, report.SelectedActions, report.RepositoryWrites, report.PromotionAuthorized)) {
		return fmt.Errorf("change proposal indicators diverged")
	}
	proofs, err := buildProofs(report.Coordinates)
	if err != nil || !reflect.DeepEqual(report.Proofs, proofs) {
		return fmt.Errorf("change proposal proofs diverged")
	}
	digestValue := report.ReportDigest
	sealed, err := sealReport(report)
	if err != nil || digestValue == "" || sealed.ReportDigest != digestValue {
		return fmt.Errorf("change proposal report digest diverged")
	}
	return nil
}

func knownStatus(value string) bool {
	return value == "SATISFIED" || value == "NOT_SATISFIED" || value == "UNRESOLVED"
}

func summarize(coordinates []Coordinate) Summary {
	summary := Summary{Total: len(coordinates), RatioDenominator: len(coordinates)}
	for _, coordinate := range coordinates {
		switch coordinate.Status {
		case "SATISFIED":
			summary.Satisfied++
		case "UNRESOLVED":
			summary.Unresolved++
		default:
			summary.NotSatisfied++
		}
	}
	summary.ReadinessBPS = summary.Satisfied * 10000 / max(summary.Total, 1)
	summary.RatioNumerator = summary.Satisfied
	return summary
}
