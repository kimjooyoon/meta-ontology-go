package predecessorresolution

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func TestBuildDoesNotSkipUnknownCloserEvidence(t *testing.T) {
	current := fixtureSHA("d")
	first, blocked, later := fixtureSHA("a"), fixtureSHA("b"), fixtureSHA("c")
	invalid := missingSelection(current, blocked)
	invalid.Report.Reason = predecessorselection.ReasonInvalid
	invalid.Report.Summary.ObservedCandidates = 1
	invalid.Report.Summary.ExactHeadCandidates = 1
	invalid.Report.ReportDigest = ""
	invalid.Report.ReportDigest = digestJSON(invalid.Report)
	_, err := Build(Input{Repository: "owner/repo", CurrentHeadSHA: current,
		ImmediatePredecessorSHA: first, SearchLimit: SearchLimit,
		Attempts: []Attempt{
			{Depth: 0, AncestorSHA: first, ParentSHA: blocked,
				Selection: missingSelection(current, first)},
			{Depth: 1, AncestorSHA: blocked, ParentSHA: later,
				Selection: invalid},
			{Depth: 2, AncestorSHA: later,
				Selection: selectedSelection(current, later)},
		}})
	if err == nil {
		t.Fatal("expected non-missing closer evidence to stop ancestry resolution")
	}
}
