package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"html"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func pathReportCondition(passed, reached bool) pathplan.ConditionResult {
	return pathplan.ConditionResult{Case: pathplan.ConditionCase{ChoiceID: "comparison", Input: 66, Expected: true},
		Observation: bodyplan.ConditionObservation{Reached: reached, Value: passed}, Passed: passed}
}

func pathReportFixture() bodyConstructOutput {
	first := pathplan.SearchAttempt{Mask: 3, Status: "CONDITION_REJECTED", Passed: 8, Total: 8,
		Conditions: []pathplan.ConditionResult{pathReportCondition(false, true), pathReportCondition(true, true), pathReportCondition(false, true)}}
	p := &bodycodegen.BodyPathReceipt{Search: pathplan.SearchResult{Schema: "gooo/typed-path-tdd-search/v1",
		Selection: pathplan.Selection{ModelCalls: 1}, Attempts: []pathplan.SearchAttempt{first, {}, {}, {}}},
		Conditions: &bodycodegen.PathConditionReceipt{Passed: 3, Declared: 3}}
	step := bodyexecution.CompositionStep{Generation: bodycodegen.Result{Report: bodycodegen.Report{
		Activity: "Choose", ActivityID: "test://choose", BodyPaths: p}}}
	return bodyConstructOutput{GeneratedNow: true, Construction: bodyexecution.JointConstruction{
		Initial: bodyexecution.Composition{Steps: []bodyexecution.CompositionStep{step}}, SelectedAttempt: 1,
		Attempts: []bodyexecution.JointAttempt{{}, {PathCandidates: []bodycodegen.PathCandidate{{
			Activity: "Choose", ActivityID: "test://choose", Mask: 5, Passed: 1, Total: 2,
			Conditions: &bodycodegen.PathConditionReceipt{Passed: 1, Declared: 2, NotReached: 1}}}}}}}
}

func TestConstructionPathReportFirstAndSelectedStaySeparate(t *testing.T) {
	r := pathReportFixture()
	before, _ := json.Marshal(r)
	for _, format := range []string{"text", "markdown"} {
		var out bytes.Buffer
		if err := writeBodyConstructResult(&out, r, format); err != nil {
			t.Fatal(err)
		}
		visible := html.UnescapeString(out.String())
		for _, want := range []string{"Choose", "4 recorded candidates; 1 recorded local model calls",
			"mask 3; CONDITION_REJECTED", "8/8 (100.00%)", "1/3 (33.33%)", "input 66; expected true; observed false",
			"mask 5; outputs 1/2 (50.00%); conditions 1/2 (50.00%); 1 not reached", "caller-program"} {
			if !strings.Contains(visible, want) {
				t.Fatalf("%s missing %q: %s", format, want, visible)
			}
		}
	}
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("report mutated original search or caller selection")
	}
}

func TestConstructionPathReportMissingAndRejectedHistory(t *testing.T) {
	r := pathReportFixture()
	p := r.Construction.Initial.Steps[0].Generation.Report.BodyPaths
	p.Search.Schema = ""
	if got := constructionInitialPathRows(p); len(got) != 1 || !strings.Contains(got[0][1], "not recorded") {
		t.Fatal(got)
	}
	p.Search.Schema, p.Search.Attempts = "gooo/typed-path-tdd-search/v1", nil
	if got := constructionInitialPathRows(p); !strings.Contains(got[1][1], "none recorded") {
		t.Fatal(got)
	}
	p.Search.Attempts = []pathplan.SearchAttempt{{Status: "TYPE_REJECTED"}}
	if got := constructionInitialPathRows(p); !strings.Contains(got[2][1], "not measured") || got[3][1] != "not recorded" {
		t.Fatal(got)
	}
	for _, index := range []int{-1, 5} {
		r.Construction.SelectedAttempt = index
		if constructionSelectedPath(r.Construction, "test://choose") != nil {
			t.Fatal("invented selection")
		}
	}
	r.Construction.SelectedAttempt = 1
	candidate := r.Construction.Attempts[1].PathCandidates[0]
	r.Construction.Attempts[1].PathCandidates = append(r.Construction.Attempts[1].PathCandidates, candidate)
	if constructionSelectedPath(r.Construction, "test://choose") != nil {
		t.Fatal("accepted ambiguous candidate")
	}
	if constructionSelectedPathShare(nil) != "not uniquely recorded; inspect JSON" {
		t.Fatal("missing data became a score")
	}
}

func TestConstructionPathReportConditionsAndEscaping(t *testing.T) {
	c := pathReportCondition(false, false)
	c.Case.Input, c.Case.Expected, c.Case.ChoiceID = 9007199254740993, false, "a|b\n<script>"
	if got := constructionRecordedConditions([]pathplan.ConditionResult{c}); got != "0/1 (0.00%); 1 not reached" {
		t.Fatal(got)
	}
	if got := constructionConditionDetail(c); !strings.Contains(got, "9007199254740993; expected false; observed not reached") {
		t.Fatal(got)
	}
	c.Passed = true
	if got := constructionRecordedConditions([]pathplan.ConditionResult{c}); !strings.Contains(got, "inconsistent") {
		t.Fatal(got)
	}
	c.Passed = false
	r := pathReportFixture()
	r.Construction.Initial.Steps[0].Generation.Report.BodyPaths.Search.Attempts[0].Conditions = []pathplan.ConditionResult{c}
	var out bytes.Buffer
	if err := writeBodyConstructResult(&out, r, "markdown"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"&#124;", "&#92;u003cscript&#92;u003e", "not reached"} {
		if !strings.Contains(out.String(), want) {
			t.Fatal(want, out.String())
		}
	}
	c.Case.ChoiceID = "choice\a"
	if got := constructionConditionDetail(c); strings.Contains(got, "invalid recorded JSON") || !strings.Contains(got, `\u0007`) {
		t.Fatal("condition ID was encoded as a Go string rather than JSON", got)
	}
}

func TestConstructionPathReportIncludesHelpersAndBoundsRows(t *testing.T) {
	r := pathReportFixture()
	step := r.Construction.Initial.Steps[0]
	r.Construction.Initial.Preparations = []bodyexecution.CompositionStep{step}
	for range 7 {
		r.Construction.Initial.Steps = append(r.Construction.Initial.Steps, step)
	}
	rows := constructionPathRows(r.Construction)
	bodies := 0
	for _, row := range rows {
		if row[0] == "Initial typed path body" {
			bodies++
		}
	}
	if bodies != 8 || rows[len(rows)-1] != [2]string{"Additional initial typed path bodies", "1; use --format json"} {
		t.Fatal(bodies, rows)
	}
}

//go:embed testdata/canonical-native-construct-20261011.json.gz
var canonicalReportObservation []byte

// This is the once-recorded public-model observation, never a fresh execution.
func TestConstructionPathReportRetainedCanonicalObservation(t *testing.T) {
	z, err := gzip.NewReader(bytes.NewReader(canonicalReportObservation))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var r bodyConstructOutput
	if err := json.NewDecoder(z).Decode(&r); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(r)
	for _, replay := range []bool{false, true} {
		r.GeneratedNow = !replay
		var out bytes.Buffer
		if err := writeBodyConstructResult(&out, r, "text"); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"4 recorded candidates; 1 recorded local model calls", "mask 3; CONDITION_REJECTED",
			"1/3 (33.33%)", "mask 0; outputs 8/8 (100.00%); conditions 3/3 (100.00%)", "initial local search history"} {
			if !strings.Contains(out.String(), want) {
				t.Fatal(want, out.String())
			}
		}
		if replay && !strings.Contains(out.String(), "New model calls during replay/evaluation: \"0\"") {
			t.Fatal("historical calls counted as new")
		}
		if !replay {
			t.Log(out.String())
		}
	}
	r.GeneratedNow = true
	after, _ := json.Marshal(r)
	if !bytes.Equal(before, after) {
		t.Fatal("rendering changed native evidence")
	}
}
