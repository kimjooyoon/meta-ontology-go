package bodyexecution

import (
	"context"
	"strings"
	"testing"
)

const constructionObservationSource = `package observations
namespace observations
entity Integer id "observation://integer"
entity Report id "observation://report" fields {
    field state id "observation://report/state" type string required one
}
activity Select(Integer) -> Report computes "return Report{state: \"waiting\"}" assembling {
    choice "state" field_value at "0" alternative "\"ready\"" intent "Set the state to ready."
    value_case "[1]" -> "{\"state\":\"ready\"}"
    attempts "1"
}
`

func constructionObservationFixture(t *testing.T, source string) Composition {
	t.Helper()
	suite, err := DecodeCompositionInputs([]byte(`{"schema":"gooo/body-composition-inputs/v1","inputs":[{"Select":7}]}`))
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "workspace.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	return prior
}

func TestRecordConstructionObservationKeepsZeroAndEffectiveBudget(t *testing.T) {
	for _, budget := range []string{"1", "64"} {
		source := strings.Replace(constructionObservationSource, `attempts "1"`, `attempts "`+budget+`"`, 1)
		if budget == "64" {
			source = strings.Replace(source, `-> "{\"state\":\"ready\"}"`, `-> "{\"state\":\"absent\"}"`, 1)
		}
		prior := constructionObservationFixture(t, source)
		rows, err := ObserveRecordConstruction(context.Background(), []byte(source), prior)
		if err != nil || len(rows) < 1 {
			t.Fatal(err, rows)
		}
		row := rows[len(rows)-1]
		if row.Counts.Matched != 0 || row.Counts.Best != 0 || row.Counts.Total != 1 || row.CandidateCount != 2 {
			t.Fatal("observed zero or candidate space changed", row)
		}
		if row.Counts.Scored != row.Counts.Budget || row.Counts.Budget != min(row.DeclaredBudget, 2) {
			t.Fatal("exhausted candidates became available budget", row)
		}
	}
}

func TestRecordConstructionObservationRecomputesTheReceipt(t *testing.T) {
	prior := constructionObservationFixture(t, constructionObservationSource)
	prior.Steps[0].Generation.Report.RecordAssembly.Passed = 1
	if _, err := ObserveRecordConstruction(context.Background(), []byte(constructionObservationSource), prior); err == nil {
		t.Fatal("edited case count became a tool observation")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ObserveRecordConstruction(ctx, []byte(constructionObservationSource), prior); err == nil {
		t.Fatal("canceled observation continued")
	}
}
