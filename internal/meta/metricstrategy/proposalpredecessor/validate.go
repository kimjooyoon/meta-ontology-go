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

func validRoute(value string) bool {
	return value == RouteDev || value == RouteMain
}

func sealReport(report Report) (Report, error) {
	report.ReportDigest = ""
	digest, err := artifact.Digest(report)
	report.ReportDigest = digest
	return report, err
}

func Validate(report Report) error {
	if report.Schema != Schema || report.Repository == "" || !validSHA(report.CurrentSubjectSHA) || !validSHA(report.PredecessorSHA) || !validRoute(report.RequestedRoute) {
		return fmt.Errorf("proposal predecessor report identity is invalid")
	}
	if report.Summary.RepositoryWrites != 0 || report.Summary.ValidCandidates < 0 ||
		report.Summary.AmbiguousCandidates < 0 || report.Summary.UnresolvedCandidates < 0 ||
		report.Summary.ObservedRuns < 0 || report.Summary.ExactRuns < 0 ||
		report.Summary.OtherRouteRuns < 0 || report.Summary.RouteUnknownRuns < 0 ||
		report.Summary.Contradictions < 0 || report.Summary.ObservedJobs < 0 || report.Summary.ExactJobs < 0 ||
		report.Summary.ProofsTotal != 5 {
		return fmt.Errorf("proposal predecessor summary is invalid")
	}
	ready := report.Decision == "SELECTED" && report.Reason == "PROPOSAL_PREDECESSOR_SELECTED"
	if ready != (report.Selected != nil) || ready != (report.Summary.SelectionBPS == 10000) || ready != (report.Summary.ProofsPassed == 5) {
		return fmt.Errorf("proposal predecessor decision diverged")
	}
	if ready {
		if report.ObservationDecision != DecisionClosed || report.ObservationResolution != ResolutionExact || report.Unknown != nil || report.Selected.HeadBranch != report.RequestedRoute {
			return fmt.Errorf("proposal predecessor closed observation diverged")
		}
	} else if report.Reason == ReasonRouteContradiction {
		if report.ObservationDecision != DecisionRefuted || report.ObservationResolution != ResolutionExact || report.Unknown != nil {
			return fmt.Errorf("proposal predecessor refuted observation diverged")
		}
	} else if report.ObservationDecision != DecisionUnknown || report.ObservationResolution != ResolutionLower || !validUnknown(report.Unknown, report.Reason) {
		return fmt.Errorf("proposal predecessor unknown observation diverged")
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

func validUnknown(unknown *Unknown, reason string) bool {
	return unknown != nil && unknown.Stage == ResolutionStage && unknown.Step == ResolutionStep &&
		unknown.Reason == reason && unknown.UnknownClass != "" && unknown.NextOperation != "" && unknown.BlockedBy != nil
}
