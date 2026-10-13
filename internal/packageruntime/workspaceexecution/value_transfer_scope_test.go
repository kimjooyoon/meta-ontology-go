package workspaceexecution

import (
	"context"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestSelectedValueQueryScope6571(t *testing.T) {
	const source = `package bounded
namespace bounded6571
entity Integer id "urn:gooo:type:integer"
activity Main(Integer) -> Integer id "urn:bounded6571:main" computes "let first = input + 41; return first"
`
	result := Result{Schema: "gooo/workspace-body-execution/v1", Program: Program{Entry: ActivityRef{LoweredName: "Main"}}, Composition: bodyexecution.Composition{GoooSource: source, SelectedSourceSHA256: sourceSHA256([]byte(source)), Plan: bodyexecution.CompositionPlan{Activities: []bodyexecution.CompositionActivity{{ID: "urn:bounded6571:main", Name: "Main"}}}}}
	query := OriginQuery{Follow: true, Limit: 1}
	flow, err := InspectSelectedValueOrigins(context.Background(), result, query)
	if err != nil || len(flow.Rows) != 1 || flow.Rows[0].Kind != "external_input" || flow.Truncated {
		t.Fatal("fresh static path", err, flow)
	}
	before := result.Composition.SelectedSourceSHA256
	result.Composition.GoooSource += " "
	if _, err := InspectSelectedValueOrigins(context.Background(), result, query); err == nil {
		t.Fatal("source digest difference ignored")
	}
	result.Composition.GoooSource = source
	result.Composition.Plan.Activities[0].ID = "urn:wrong6571"
	if _, err := InspectSelectedValueOrigins(context.Background(), result, query); err == nil {
		t.Fatal("source identity difference ignored")
	}
	result.Composition.Plan.Activities[0].ID = "urn:bounded6571:main"
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := InspectSelectedValueOrigins(cancelled, result, query); err != context.Canceled {
		t.Fatal("context not propagated", err)
	}
	if result.Composition.SelectedSourceSHA256 != before {
		t.Fatal("query mutated source identity")
	}
	t.Log("6571: static source path; unchanged source/plan identities required; cancellation; no execution")
}
