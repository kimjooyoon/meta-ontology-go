package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestSourceFillReplayPreservesHistoricalAdvisoryExclusions(t *testing.T) {
	for _, fixture := range []struct{ name, activity string }{
		{"source-ir-fill", "Lift"}, {"source-ir-fill-record-tiny", "ReviewCandidate"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			source, prior := sourceFillReplayFixture(t, fixture.name, fixture.activity)
			for _, historical := range []string{
				"calls, loops, imports, external effects, optional-field synthesis and repeated record fields",
				"recursive, indirect, external and assembling-callee calls; " +
					"loops, imports, external effects, optional-field synthesis and repeated record fields",
			} {
				prior.Report.UnsupportedConstructs = historical
				before, err := json.Marshal(prior)
				if err != nil {
					t.Fatal(err)
				}
				realized, err := RealizeSourceAssembly(context.Background(), "fill.gooo", source, prior)
				if err != nil || realized.ModelCalls != 0 {
					t.Fatal("historical explanation changed replay", err)
				}
				generated, err := Generate("fill.gooo", []byte(realized.Source), fixture.activity)
				if err != nil || generated.Source != prior.Source {
					t.Fatal("advisory text changed the generated program", err)
				}
				after, err := json.Marshal(prior)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("replay rewrote the original observation", err)
				}
			}
		})
	}
}

func TestSourceFillReplayAdvisoryChangeKeepsExecutionClaimsBound(t *testing.T) {
	for _, mutation := range []struct {
		name   string
		change func(*Result)
	}{
		{"activity", func(p *Result) { p.Report.ActivityID += "/other" }},
		{"source", func(p *Result) { p.Report.SourceDigest = strings.Repeat("0", 64) }},
		{"program", func(p *Result) { p.Report.ProgramDigest = strings.Repeat("0", 64) }},
		{"generated", func(p *Result) { p.Report.GeneratedDigest = strings.Repeat("0", 64) }},
		{"output-type", func(p *Result) { p.Report.OutputType = "Text" }},
		{"finite-count", func(p *Result) { p.Report.BodyFill.TestCasesPassed-- }},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			source, prior := sourceFillReplayFixture(t, "source-ir-fill", "Lift")
			prior.Report.UnsupportedConstructs = "historical explanatory text"
			mutation.change(&prior)
			if _, err := RealizeSourceAssembly(context.Background(), "fill.gooo", source, prior); err == nil {
				t.Fatal("advisory comparison hid changed execution claims", mutation.name)
			}
		})
	}
}
