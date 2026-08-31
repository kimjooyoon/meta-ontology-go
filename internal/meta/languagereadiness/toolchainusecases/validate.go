package toolchainusecases

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

func Validate(report Report, expectedHead string) error {
	if report.Schema != ReportSchema || report.Producer != "toolchainusecases.Evaluate" ||
		report.Consumer != "self-improvement-cycle" || report.MetaOperation != "execute-versioned-use-cases" ||
		report.HeadSHA != expectedHead || report.Source.ExpectedHeadSHA != expectedHead || !validHead(expectedHead) {
		return fmt.Errorf("executable use case report identity mismatch")
	}
	if len(report.Cases) != totalCases || len(report.Indicators) != 8 || len(report.Proofs) != 3 ||
		!validDigest(report.Source.ConceptArtifactDigest) || !validDigest(report.Source.CatalogDigest) ||
		!validDigest(report.Source.RegistryDigest) {
		return fmt.Errorf("executable use case report shape mismatch")
	}
	definitions := make([]CaseDefinition, 0, len(report.Cases))
	for _, item := range report.Cases {
		definitions = append(definitions, item.Definition)
		status := "NOT_SATISFIED"
		if item.ObservedDecision == "UNKNOWN" {
			status = "UNRESOLVED"
		} else if item.ObservedDecision == item.Definition.ExpectedDecision {
			status = "SATISFIED"
		}
		if item.Status != status || item.EvidenceDigest != caseDigest(item, report.Source.ConceptArtifactDigest) {
			return fmt.Errorf("executable use case result mismatch")
		}
	}
	if !reflect.DeepEqual(definitions, expectedRegistry().Cases) {
		return fmt.Errorf("executable use case definitions mismatch")
	}
	candidate := report
	candidate.Decision, candidate.Reason, candidate.Resolution, candidate.ReportDigest = "", "", "", ""
	candidate.Summary, candidate.Indicators, candidate.Proofs = Summary{}, nil, nil
	if !reflect.DeepEqual(finish(candidate), report) {
		return fmt.Errorf("executable use case derived evidence mismatch")
	}
	return nil
}

func validHead(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
