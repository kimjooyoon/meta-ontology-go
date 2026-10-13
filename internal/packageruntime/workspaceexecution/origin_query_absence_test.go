package workspaceexecution

import (
	"context"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRecordedOriginAbsenceAndSchema5879(t *testing.T) {
	const activity = "urn:absence5879:activity"
	const candidate = "candidate5879"
	fill := &bodycodegen.IRBodyFillReceipt{SelectedCandidateID: candidate, SourceOrigin: &bodycodegen.SourceFillOrigin{ActivityID: activity, CandidateID: candidate}}
	result := Result{Schema: "gooo/workspace-body-execution/v1", BodyFills: []BodyFillStep{{Generation: bodycodegen.Result{Report: bodycodegen.Report{ActivityID: activity, BodyFill: fill}}}}}
	report, err := InspectRecordedOrigins(context.Background(), result, OriginQuery{CandidateID: candidate, Limit: 8})
	if err != nil || report.OriginsMatched != 1 || report.GraphsAvailable != 0 || len(report.Rows) != 0 || len(report.Issues) != 1 || report.Issues[0].Reason != "NO_RETAINED_VALUE_GRAPH" {
		t.Fatal("absent graph", err, report)
	}
	fill.SourceOrigin = nil
	filtered, err := InspectRecordedOrigins(context.Background(), result, OriginQuery{CandidateID: "different5879", Limit: 8})
	if err != nil || filtered.OriginsSeen != 1 || filtered.OriginsMatched != 0 || len(filtered.Issues) != 0 {
		t.Fatal("unmatched missing origin leaked into filter", err, filtered)
	}
	missing, err := InspectRecordedOrigins(context.Background(), result, OriginQuery{CandidateID: candidate, Limit: 8})
	if err != nil || len(missing.Issues) != 1 || missing.Issues[0].Reason != "NO_RETAINED_SOURCE_ORIGIN" {
		t.Fatal("missing source origin", err, missing)
	}
	flow := bodycodegen.InspectBodyValueOrigins("return input + 199", 1)
	flow.Schema = "future5879"
	if _, _, err := bodycodegen.SourceFillUses(bodycodegen.SourceFillOrigin{SelectedBodySHA256: flow.BodySHA256, DataFlow: &flow}, 8); err == nil {
		t.Fatal("unknown graph meaning interpreted")
	}
	t.Log("5879: synthetic absent-metadata query states retain missing reasons and candidate filter; unknown graph schema reported; no body or model execution")
}
