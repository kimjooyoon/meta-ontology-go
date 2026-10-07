package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func constructionProfilesFixture(t *testing.T) packageExecutionReceipt {
	t.Helper()
	root := filepath.Join("..", "..", "examples", "construction-observation")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"),
		filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var saved packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Result.Runtime.FinitePassed != 4 || saved.Result.Runtime.FiniteTotal != 4 {
		t.Fatal("constructed program did not satisfy its four runtime outputs", saved.Result.Runtime)
	}
	return saved
}

func TestPackageConstructionFillSetAndSearchPrefix(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	saved := constructionProfilesFixture(t)
	raw, _ := json.Marshal(saved)
	path := filepath.Join(t.TempDir(), "construction.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--construction-receipt", path,
		filepath.Join("..", "..", "examples", "assembly-explainer", "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var explained packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &explained); err != nil {
		t.Fatal(err)
	}
	checkConstructionProfileObservations(t, explained)
	saved.Result.BodyFills[0].Generation.Report.BodyFill.CandidateScores[0].TestCasesPassed++
	changed, _ := json.Marshal(saved)
	if _, _, err := packageConstructionInputs(context.Background(), changed, packageruntime.EntrySpec{}); err == nil {
		t.Fatal("changed source-fill candidate score was interpreted")
	}
}

func checkConstructionProfileObservations(t *testing.T, saved packageExecutionReceipt) {
	t.Helper()
	o := saved.ConstructionInput
	if o == nil || o.Schema != "gooo/construction-input/v2" || len(o.Rows) < 4 || o.ModelCalls != 0 {
		t.Fatal("construction profiles missing", o)
	}
	best, searches := 0, 0
	for i, row := range o.Rows {
		if !row.ScoringCompleted || row.InputIndex == nil || *row.InputIndex != i {
			t.Fatal("scored row lost its policy input mapping", row)
		}
		if i < 3 {
			if row.AttemptIndex != nil || row.CandidateIndex == nil || *row.CandidateIndex != i {
				t.Fatal("fill candidate index became a chronological attempt", row)
			}
			if row.Profile != "source_fill" || row.View != "scored_set" || row.Counts.Total != 3 || row.Counts.Best != 3 || row.Counts.Scored != 3 || row.Counts.Budget != 3 {
				t.Fatal("all-candidate scoring became chronological search or included holdout", row)
			}
		} else {
			searches++
			best = max(best, row.Counts.Matched)
			if row.Profile != "ir_search" || row.View != "attempt_prefix" || row.Counts.Best != best || row.Counts.Total != 5 || row.Counts.Scored != searches {
				t.Fatal("search prefix saw future results or holdout", row)
			}
		}
	}
	var first map[string]string
	if err := json.Unmarshal(saved.Result.Runtime.Traces[0].Deliveries[0].Actual, &first); err != nil || first["next_operation"] != "USE_OBSERVED_CANDIDATE" {
		t.Fatal("Gooo failed to use the already scored better candidate", first, err)
	}
	if saved.Decision != "OBSERVED" || saved.Result.Runtime.FiniteTotal != 0 || saved.Result.Runtime.ModelCalls != 0 {
		t.Fatal("interpretation became original runtime success", saved)
	}
}
