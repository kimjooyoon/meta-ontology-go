package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestSourceObserveCandidateKeepsAcquiredCases(t *testing.T) {
	ctx := context.Background()
	source := []byte(sourceObserveFixture)
	document, err := DecodeSourcePathDocument(ctx, "candidate.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := GenerateWithTypedPaths(ctx, "candidate.gooo", source, "Fee", document, "")
	if err != nil {
		t.Fatal(err)
	}
	prior := initial.Report.BodyPaths
	before, _ := json.Marshal(prior)
	for _, mask := range []uint16{1, 0} {
		candidate, _, err := RealizeSourcePathCandidate(ctx, "candidate.gooo", source, "Fee", prior, mask)
		if err != nil || candidate.Total != 2 || candidate.Passed != 1+int(mask) ||
			len(candidate.Cases) != 2 || candidate.Cases[1].Input != 17 || candidate.Cases[1].Expected != -6 {
			t.Fatalf("acquired obligation was lost for mask %d: %+v, %v", mask, candidate, err)
		}
		if candidate.Schema != "gooo/path-candidate/v2" || candidate.Observation == nil ||
			candidate.Observation.InitialCases != 1 || candidate.Observation.AddedCases != 1 ||
			len(candidate.Observation.OracleCases) != 2 || !candidate.Observation.OracleCases[1].Passed {
			t.Fatal("candidate did not retain observation provenance", candidate)
		}
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("candidate realization rewrote the original observation")
	}
}

func TestSourceObserveCandidateRebindsCurrentSource(t *testing.T) {
	ctx := context.Background()
	source := []byte(sourceObserveFixture)
	document, err := DecodeSourcePathDocument(ctx, "rebind.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := GenerateWithTypedPaths(ctx, "rebind.gooo", source, "Fee", document, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source string
		wantError    bool
	}{
		{"unrelated source", sourceObserveFixture + "\nactivity Unrelated(Integer) -> Integer computes \"return input + 19\"", false},
		{"changed answer", strings.Replace(sourceObserveFixture, "return 11 - input", "return input - 11", 1), true},
		{"changed identity", strings.Replace(sourceObserveFixture, "observed://reference", "observed://different-reference", 1), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, err := RealizeSourcePathCandidate(ctx, "rebind.gooo", []byte(tc.source), "Fee", initial.Report.BodyPaths, 1)
			if (err != nil) != tc.wantError {
				t.Fatal(r, err)
			}
			if tc.wantError && r.Stage != "OBSERVED_CASES" {
				t.Fatal("conflict lost its stage", r)
			}
			if !tc.wantError && (r.Total != 2 || r.Passed != 2) {
				t.Fatal("unrelated source erased acquired answers", r)
			}
		})
	}
}

func TestSourceObserveCandidateCannotDropOrRelabelReceipt(t *testing.T) {
	ctx := context.Background()
	source := []byte(sourceObserveFixture)
	doc, err := DecodeSourcePathDocument(ctx, "receipt.gooo", source, "Fee", nil)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := GenerateWithTypedPaths(ctx, "receipt.gooo", source, "Fee", doc, "")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(initial.Report.BodyPaths)
	for name, change := range map[string]func(*BodyPathReceipt){
		"missing": func(r *BodyPathReceipt) { r.Observation = nil },
		"options": func(r *BodyPathReceipt) { r.Observation.Options.MaxRounds++ },
		"answer": func(r *BodyPathReceipt) {
			o := r.Observation
			o.Rounds[0].Observation.Expected = 6
			cases := append([]pathplan.TestCase(nil), o.InitialCases...)
			cases = append(cases, *o.Rounds[0].Observation)
			o.EffectiveCasesSHA = pathCasesDigest(cases)
			// Keep the subsequent ranking's case digest coherent too. The actual
			// source oracle must still refute the substituted label.
			o.Rounds[1].Ranking.CasesSHA256 = strings.TrimPrefix(o.EffectiveCasesSHA, "sha256:")
		},
	} {
		t.Run(name, func(t *testing.T) {
			var prior BodyPathReceipt
			if err := json.Unmarshal(raw, &prior); err != nil {
				t.Fatal(err)
			}
			change(&prior)
			if _, _, err := RealizeSourcePathCandidate(ctx, "receipt.gooo", source, "Fee", &prior, 1); err == nil {
				t.Fatal("changed observation was accepted")
			}
		})
	}
}
