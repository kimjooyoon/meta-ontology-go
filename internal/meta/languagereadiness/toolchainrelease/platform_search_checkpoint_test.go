package toolchainrelease

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func searchCheckpointFixture(t *testing.T, mode string) ([]byte, [][]byte) {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "search-checkpoint-"+mode+".json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	files := make([][]byte, 3)
	for i, name := range []string{"main.gooo.fixture", "construction-cases.json", "evaluation-cases.json"} {
		files[i], err = os.ReadFile(filepath.Join("../../../..", searchCheckpointRoot, name))
		if err != nil {
			t.Fatal(err)
		}
	}
	return raw, files
}

func TestSearchCheckpointReleaseRetainsOriginalPoolAndReplay(t *testing.T) {
	raw, files := searchCheckpointFixture(t, "construct")
	retained, err := validateSearchCheckpoint(raw, files[0], files[1], files[2], false, "")
	if err != nil {
		t.Fatal(err)
	}
	replayed, _ := searchCheckpointFixture(t, "replay")
	if got, err := validateSearchCheckpoint(replayed, files[0], files[1], files[2], true, retained); err != nil || got != retained {
		t.Fatal("saved complete history differs", got, err)
	}
	if _, err := validateSearchCheckpoint(replayed, files[0], files[1], files[2], true, ""); err == nil {
		t.Fatal("replay without the original construction was accepted")
	}
	if _, err := validateSearchCheckpoint(raw, append(files[0], '\n'), files[1], files[2], false, ""); err == nil {
		t.Fatal("changed source was accepted")
	}
	changedCases := bytes.ReplaceAll(files[2], []byte("12"), []byte("13"))
	if bytes.Equal(changedCases, files[2]) {
		t.Fatal("case mutation missed its target")
	}
	if _, err := validateSearchCheckpoint(raw, files[0], files[1], changedCases, false, ""); err == nil {
		t.Fatal("native values were not checked against the independent case file")
	}
}

func TestCheckpointHistoryDigestPreservesExactIntegers(t *testing.T) {
	original, err := checkpointConstructionDigest([]byte(`{"input":9007199254740993,"other":1}`))
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := checkpointConstructionDigest([]byte(`{ "other":1, "input":9007199254740993 }`))
	if err != nil || original != reordered {
		t.Fatal("object ordering changed the history", err)
	}
	rounded, err := checkpointConstructionDigest([]byte(`{"input":9007199254740992,"other":1}`))
	if err != nil || original == rounded {
		t.Fatal("history lost exact integer identity", err)
	}
}

func TestSearchCheckpointReleaseRejectsChangedObservations(t *testing.T) {
	raw, files := searchCheckpointFixture(t, "construct")
	retained, err := validateSearchCheckpoint(raw, files[0], files[1], files[2], false, "")
	if err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]nativeSmokeEdit{
		"native output":       {[]any{"evaluation", "runtime", "traces", 1, "deliveries", 0, "actual"}, 13},
		"native input":        {[]any{"evaluation", "runtime", "traces", 1, "deliveries", 0, "input"}, 6},
		"missing model count": {[]any{"evaluation", "new_model_calls"}, nil},
		"new prediction":      {[]any{"evaluation", "new_model_calls"}, 1},
		"unknown erased":      {[]any{"evaluation", "runtime", "input_separation", "unknown_inputs"}, 0},
		"candidate space":     {[]any{"construction", "candidate_space"}, "16"},
		"planning source":     {[]any{"construction", "attempts", 0, "search_candidates", 0, "planning_source_sha256"}, "changed"},
		"local output":        {[]any{"construction", "attempts", 0, "search_candidates", 0, "attempt", "case_results", 0, "actual"}, 99},
		"call input":          {[]any{"construction", "attempts", 0, "search_candidates", 0, "local_calls", "evaluations", 0, "calls", 0, "inputs"}, []any{99}},
		"probe repeated":      {[]any{"construction", "attempts", 0, "search_candidates", 0, "local_calls", "evaluations", 0, "phase"}, "candidate-generation"},
		"missing pool":        {[]any{"construction", "initial", "preparations", 1, "generation", "report", "body_search", "initial_candidate_set"}, nil},
		"changed pool":        {[]any{"construction", "initial", "preparations", 1, "generation", "report", "body_search", "initial_candidate_set", "candidates", 1, "expression"}, "99"},
	} {
		t.Run(name, func(t *testing.T) {
			changed := changedNativeSmoke(t, raw, edit)
			if _, err := validateSearchCheckpoint(changed, files[0], files[1], files[2], false, ""); err == nil {
				t.Fatal("changed observation accepted")
			}
		})
	}
	replayed, _ := searchCheckpointFixture(t, "replay")
	replayed = changedNativeSmoke(t, replayed, nativeSmokeEdit{[]any{"construction", "elapsed_ns"}, 0})
	if _, err := validateSearchCheckpoint(replayed, files[0], files[1], files[2], true, retained); err == nil {
		t.Fatal("replay replaced part of the original construction history")
	}
}

func TestSearchCheckpointReleaseOnPackagedCompiler(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "local-native"}}
	if output := os.Getenv("GOOO_CHECKPOINT_SMOKE_OUTPUT"); output != "" {
		input.OutputDir = output
		if err := os.MkdirAll(output, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := smokeSearchCheckpoint(binary, t.TempDir(), input); err != nil {
		t.Fatal(err)
	}
}
