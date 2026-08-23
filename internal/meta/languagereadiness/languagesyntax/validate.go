package languagesyntax

import (
	"fmt"
	"reflect"
	"strings"
)

func Validate(report Report, expectedHead string) error {
	if report.Schema != ReportSchema || report.Producer != "languagesyntax.Evaluate" ||
		report.Consumer != "self-improvement-cycle" || report.MetaOperation != "prove-language-syntax-roundtrip" ||
		report.HeadSHA != expectedHead || report.Source.ExpectedHeadSHA != expectedHead || !validHead(expectedHead) {
		return fmt.Errorf("language syntax report identity mismatch")
	}
	if len(report.Cases) != totalCases || len(report.Indicators) != 16 || len(report.Proofs) != 3 ||
		!validDigest(report.Source.ConceptArtifactDigest) || !validDigest(report.Source.CatalogDigest) ||
		!validDigest(report.Source.RegistryDigest) || !validDigest(report.Source.CorpusDigest) {
		return fmt.Errorf("language syntax report shape mismatch")
	}
	if report.Source.ObservationKnown && report.Source.CorpusDigest != digestJSON(report.Source.GoooFiles) {
		return fmt.Errorf("language syntax corpus digest mismatch")
	}
	definitions := make([]CaseDefinition, 0, len(report.Cases))
	for _, item := range report.Cases {
		definitions = append(definitions, item.Definition)
		if item.Status != caseStatus(item) || item.EvidenceDigest != caseDigest(item, report.Source) ||
			!validCaseEvidence(item) {
			return fmt.Errorf("language syntax case evidence mismatch")
		}
	}
	if !reflect.DeepEqual(definitions, expectedRegistry().Cases) {
		return fmt.Errorf("language syntax case definitions mismatch")
	}
	for _, file := range report.Source.GoooFiles {
		if !strings.HasSuffix(file.Path, ".gooo") || file.GoooLines <= 0 || !validDigest(file.SourceDigest) {
			return fmt.Errorf("language syntax file observation mismatch")
		}
	}
	candidate := report
	candidate.Decision, candidate.Reason, candidate.Resolution, candidate.ReportDigest = "", "", "", ""
	candidate.Summary, candidate.Indicators, candidate.Proofs = Summary{}, nil, nil
	if !reflect.DeepEqual(finish(candidate), report) {
		return fmt.Errorf("language syntax derived evidence mismatch")
	}
	return nil
}

func validCaseEvidence(item CaseResult) bool {
	evidence := item.Evidence
	if evidence.ObservedDecision == "UNKNOWN" {
		return true
	}
	if !validDigest(evidence.SourceDigest) || evidence.SourceLines <= 0 {
		return false
	}
	if item.Definition.Kind == KindInvalid {
		return item.Status != "SATISFIED" || evidence.DiagnosticRejected
	}
	if item.Status != "SATISFIED" {
		return true
	}
	return validDigest(evidence.ASTDigest) && validDigest(evidence.CanonicalDigest) &&
		validDigest(evidence.SemanticDigest) && evidence.ASTReplayed && evidence.ByteReplayed &&
		evidence.SemanticReplayed && evidence.GetPut && evidence.PutGet
}
