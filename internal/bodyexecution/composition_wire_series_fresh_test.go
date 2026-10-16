package bodyexecution

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCompositionWireSeriesFresh1879(t *testing.T) {
	_, _, value := wireFixture1874(t)
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{
		Inputs: map[string]json.RawMessage{}, Expected: map[string]json.RawMessage{},
	}}}
	rawValue, _ := json.Marshal(value)
	suite.Cases[0].Inputs["First"], suite.Cases[0].Expected["Second"] = rawValue, rawValue
	for _, mode := range []string{"sixteen", "seventeen", "suite_over", "series_over"} {
		t.Run(mode, func(t *testing.T) {
			count := 16
			if mode == "seventeen" {
				count++
			}
			series := CompositionCaseSeries{Schema: "gooo/body-composition-case-series/v1"}
			for range count {
				series.Suites = append(series.Suites, suite)
			}
			if mode == "suite_over" {
				oversized := suite
				oversized.Cases = []CompositionCase{{Inputs: map[string]json.RawMessage{"First": json.RawMessage(`"` + string(bytes.Repeat([]byte("x"), CompositionInputMaxBytes)) + `"`)}}}
				series.Suites = []CompositionCases{oversized}
			}
			raw, err := json.Marshal(series)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "series_over" {
				raw = append(raw, bytes.Repeat([]byte(" "), CompositionSeriesMaxBytes+1-len(raw))...)
			}
			got, err := DecodeCompositionCaseSeries(raw)
			if mode != "sixteen" {
				if err == nil {
					t.Fatal("series or individual suite limit bypassed")
				}
				return
			}
			if err != nil || len(raw) <= 512<<10 || len(got.Suites) != 16 {
				t.Fatal("expanded aggregate series", len(raw), err)
			}
		})
	}
}
