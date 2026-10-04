package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCompositionCaseSeriesKeepsSuiteBoundariesAndTypedExpectations(t *testing.T) {
	source, suite := compositionFixture(t)
	series := CompositionCaseSeries{Schema: "gooo/body-composition-case-series/v1", Suites: []CompositionCases{suite, suite}}
	raw, err := json.Marshal(series)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCompositionCaseSeries(raw)
	if err != nil || len(decoded.Suites) != 2 {
		t.Fatal(decoded, err)
	}
	if err := ValidateCompositionSuites(context.Background(), "graph.gooo", source, decoded.Suites); err != nil {
		t.Fatal(err)
	}
	decoded.Suites[1].Cases[0].Inputs["unknown"] = json.RawMessage(`1`)
	if err := ValidateCompositionSuites(context.Background(), "graph.gooo", source, decoded.Suites); err == nil {
		t.Fatal("unknown second-suite input accepted")
	}
	for _, changed := range []CompositionCaseSeries{
		{Schema: "unknown", Suites: []CompositionCases{suite}},
		{Schema: series.Schema},
		{Schema: series.Schema, Suites: make([]CompositionCases, 17)},
		{Schema: series.Schema, Suites: []CompositionCases{{Schema: "unknown", Cases: suite.Cases}}},
	} {
		raw, err := json.Marshal(changed)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeCompositionCaseSeries(raw); err == nil {
			t.Fatal("invalid suite series accepted")
		}
	}
}
