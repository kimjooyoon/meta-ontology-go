package proposalpredecessor

import (
	"fmt"
	"reflect"
	"regexp"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

var shaPattern = regexp.MustCompile("^[0-9a-f]{40}$")

func validSHA(value string) bool {
	return shaPattern.MatchString(value)
}

func sealReport(report Report) (Report, error) {
	report.ReportDigest = ""
	digest, err := artifact.Digest(report)
	report.ReportDigest = digest
	return report, err
}

func Validate(report Report) error {
	if report.Schema != Schema || report.Repository == "" || !validSHA(report.CurrentSubjectSHA) || !validSHA(report.PredecessorSHA) {
		return fmt.Errorf("proposal predecessor report identity is invalid")
	}
	if report.Summary.RepositoryWrites != 0 || report.Summary.ValidCandidates < 0 ||
		report.Summary.AmbiguousCandidates < 0 || report.Summary.UnresolvedCandidates < 0 ||
		report.Summary.ObservedJobs < 0 || report.Summary.ExactJobs < 0 ||
		report.Summary.ProofsTotal != 5 {
		return fmt.Errorf("proposal predecessor summary is invalid")
	}
	ready := report.Decision == "SELECTED" && report.Reason == "PROPOSAL_PREDECESSOR_SELECTED"
	if ready != (report.Selected != nil) || ready != (report.Summary.SelectionBPS == 10000) || ready != (report.Summary.ProofsPassed == 5) {
		return fmt.Errorf("proposal predecessor decision diverged")
	}
	if ready && (!candidateSelectedReady(*report.Selected, report.PredecessorSHA) ||
		report.Summary.ExactJobs != 1 || report.Summary.ValidCandidates != 1 ||
		report.Summary.AmbiguousCandidates != 0 || report.Summary.UnresolvedCandidates != 0) {
		return fmt.Errorf("proposal predecessor selected evidence diverged")
	}
	if !reflect.DeepEqual(report.Indicators, buildIndicators(report.Summary)) {
		return fmt.Errorf("proposal predecessor indicators diverged")
	}
	proofs, err := buildProofs(report.Selected, report.PredecessorSHA, report.Summary)
	if err != nil || !reflect.DeepEqual(report.Proofs, proofs) {
		return fmt.Errorf("proposal predecessor proofs diverged")
	}
	digest := report.ReportDigest
	sealed, err := sealReport(report)
	if err != nil || digest == "" || digest != sealed.ReportDigest {
		return fmt.Errorf("proposal predecessor digest diverged")
	}
	return nil
}

func candidateSelectedReady(selected Selected, predecessorSHA string) bool {
	return candidateReady(Candidate{Selected: selected}, predecessorSHA)
}
