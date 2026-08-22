package guardedcapability

import (
	_ "embed"
	"encoding/json"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
)

//go:embed foundation.json
var foundationRaw []byte

func foundationReport() (guardedpromotion.Report, error) {
	report := guardedpromotion.Report{}
	if err := json.Unmarshal(foundationRaw, &report); err != nil {
		return report, err
	}
	return report, nil
}

func foundationExact(source Source) bool {
	expected, err := foundationReport()
	if err != nil || guardedpromotion.Validate(source.FoundationReport) != nil ||
		!reflect.DeepEqual(source.FoundationReport, expected) {
		return false
	}
	report := source.FoundationReport
	return source.WorkflowRunID == FoundationWorkflowRunID &&
		source.ArtifactID == FoundationArtifactID &&
		source.ArtifactDigest == FoundationArtifactDigest &&
		source.ReportFileSHA == FoundationReportFileSHA &&
		report.ReportDigest == FoundationReportDigest &&
		report.Decision == guardedpromotion.DecisionAuthorized &&
		report.Resolution == guardedpromotion.ResolutionExact &&
		report.Source.CurrentHeadSHA == FoundationSubjectSHA &&
		report.Summary.Satisfied == 12 && report.Summary.Total == 12 &&
		report.Summary.Unresolved == 0 && report.Summary.RepositoryWrites == 0 &&
		report.Summary.ReadinessPromotionAuthorized &&
		!report.Summary.RepositoryMutationAuthorized
}
