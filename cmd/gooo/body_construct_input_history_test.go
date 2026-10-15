package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestConstructionInputHistoryReport(t *testing.T) {
	r := constructionReportFixture()
	r.Evaluation.Runtime.ConstructionInputHistorySHA256 = "fixture-history"
	r.Evaluation.Runtime.InputSeparation = bodyexecution.CompositionInputSeparation{Status: "PASS", Reason: "DISJOINT_EXPECTATIONS_MATCHED",
		UniqueInputs: 3, OverlappingInputs: 2, DisjointInputs: 1, DisjointCasesPassed: 1, DuplicateRows: 1, Scope: "recorded arguments"}
	before, _ := json.Marshal(r)
	for _, format := range []string{"text", "markdown"} {
		var out bytes.Buffer
		if err := writeBodyConstructResult(&out, r, format); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"2 overlapping, 1 disjoint, 0 unknown", "1/1 (100.00%)", "model-training exposure remains unknown"} {
			if !strings.Contains(out.String(), want) {
				t.Fatal("missing input history", want, out.String())
			}
		}
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("report rewrote observed inputs")
	}
	r.Evaluation.Runtime.InputSeparation.UnknownInputs = 1
	if rows := constructionInputHistoryRows(r.Evaluation.Runtime); !strings.Contains(rows[0][1], "inconsistent") {
		t.Fatal("inconsistent input denominator accepted", rows)
	}
	if rows := constructionInputHistoryRows(bodyexecution.CompositionRuntime{}); rows[0][1] != "not recorded" {
		t.Fatal("legacy record acquired input evidence", rows)
	}
}

func TestConstructionInputHistoryCLI(t *testing.T) {
	fixture := "../../examples/caller-guided-construction/"
	args := []string{"body-construct", "--source", fixture + "main.gooo.fixture", "--entry", "Main",
		"--construction-cases", fixture + "construction-cases.json", "--cases", fixture + "evaluation-cases.json", "--attempts", "2"}
	var out, diagnostic bytes.Buffer
	if code := run(args, &out, &diagnostic); code != exitOK {
		t.Fatal(code, diagnostic.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	m := result.Evaluation.Runtime.InputSeparation
	if result.Evaluation.Runtime.FinitePassed != 4 || m.OverlappingInputs != 1 || m.DisjointInputs != 3 || m.DisjointCasesPassed != 3 ||
		m.UnknownInputs != 0 || result.Evaluation.Runtime.ConstructionInputHistorySHA256 == "" {
		t.Fatal("CLI lost input history", m)
	}
}
