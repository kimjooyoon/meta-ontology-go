package bodyexecution

import (
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRecordObservationSeparatesProposalFromSelection(t *testing.T) {
	for _, test := range []struct {
		name       string
		prediction *jointdecision.ThreePrediction
		proposed   int
	}{
		{"deterministic", nil, -1},
		{"zero-mask", &jointdecision.ThreePrediction{Mask: 0}, 0},
		{"incomplete-proposal", &jointdecision.ThreePrediction{Mask: 7}, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			report := bodycodegen.Report{RecordAssembly: &bodycodegen.RecordAssemblyReceipt{
				Prediction: test.prediction, SelectedMask: 3, Ranking: []uint16{0, 7, 3},
				Attempts: []bodycodegen.RecordAssemblyAttempt{{Mask: 0, Passed: 0, Total: 1}, {Mask: 7, Passed: 0, Total: 1}, {Mask: 3, Passed: 1, Total: 1}}}}
			rows := observeRecordAttempts(report, 3)
			for i, row := range rows {
				if row.Proposed != (i == test.proposed) || row.Selected != (i == 2) {
					t.Fatal("recorded proposal was lost or confused with final selection", i, row)
				}
			}
		})
	}
}
