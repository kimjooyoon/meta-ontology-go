package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func sourceFillReplayFixture(t *testing.T, name, activity string) ([]byte, Result) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/" + name + ".gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	spec, err := SourceAssembly(context.Background(), "fill.gooo", source, activity)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateWithSourceIRBodyFill(context.Background(), "fill.gooo", source, activity, spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.MarshalIndent(prior, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &prior); err != nil {
		t.Fatal(err)
	}
	return source, prior
}

func TestSourceFillRealizationReplaysScalarAndDerivedRecord(t *testing.T) {
	for _, fixture := range []struct{ name, activity string }{
		{"source-ir-fill", "Lift"}, {"source-ir-fill-record-tiny", "ReviewCandidate"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			source, prior := sourceFillReplayFixture(t, fixture.name, fixture.activity)
			realized, err := RealizeSourceAssembly(context.Background(), "fill.gooo", source, prior)
			if err != nil {
				t.Fatal(err)
			}
			if realized.ModelCalls != 0 || strings.Contains(realized.Source, "__GOOO_BODY_HOLE_") || strings.Contains(realized.Source, "assembling") {
				t.Fatal("fill did not produce a model-free reusable checkpoint")
			}
			generated, err := Generate("fill.gooo", []byte(realized.Source), fixture.activity)
			if err != nil || generated.Source != prior.Source {
				t.Fatal("realized projection differs", err)
			}
		})
	}
}

func TestSourceFillRealizationRejectsChangedObservations(t *testing.T) {
	for _, field := range []string{"source", "request", "candidate", "selected", "probe", "completeness", "checkpoint"} {
		t.Run(field, func(t *testing.T) {
			source, prior := sourceFillReplayFixture(t, "source-ir-fill", "Lift")
			switch field {
			case "source":
				source = append(source, '\n')
			case "request":
				prior.Report.BodyFill.Decision.RequestSHA256 = strings.Repeat("0", 64)
			case "candidate":
				prior.Report.BodyFill.CandidateScores[0].TestCasesPassed--
			case "selected":
				prior.Report.BodyFill.SelectedCaseResults[0].Actual++
			case "probe":
				prior.Report.BodyFill.BehavioralProbes.ProbeInputs[0]++
			case "completeness":
				prior.Report.CompletenessReceipt = nil
			case "checkpoint":
				prior.GoooSource += "\n"
			}
			if _, err := RealizeSourceAssembly(context.Background(), "fill.gooo", source, prior); err == nil {
				t.Fatal("changed observation replayed", field)
			}
		})
	}
}
