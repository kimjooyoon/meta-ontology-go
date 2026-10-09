package toolchainrelease

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func typedPathsFixture(t *testing.T, name string) []byte {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "typed-path-"+name+".json.gz"))
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
	return raw
}

func typedPathsSource(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../../../..", typedPathsRoot, name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestTypedReleasePreflightAndExactNativeReplay(t *testing.T) {
	if err := validateTypedPreflight(typedPathsFixture(t, "unary"), typedPathsSource(t, "unary.gooo.fixture")); err != nil {
		t.Fatal(err)
	}
	source := typedPathsSource(t, "mixed-model.gooo.fixture")
	contextSHA, err := validateTypedRecordPreflight(typedPathsFixture(t, "record"), source)
	if err != nil {
		t.Fatal(err)
	}
	feedback, cases := typedPathsSource(t, "mixed-construction-cases.json"), typedPathsSource(t, "mixed-evaluation-cases.json")
	selected, err := validateTypedSmoke(typedPathsFixture(t, "construct"), source, feedback, cases, contextSHA, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := validateTypedSmoke(typedPathsFixture(t, "replay"), source, feedback, cases, contextSHA, true, selected); err != nil {
		t.Fatal(err)
	}
}

func TestTypedReleaseRejectsChangedNativeObservations(t *testing.T) {
	raw, source := typedPathsFixture(t, "construct"), typedPathsSource(t, "mixed-model.gooo.fixture")
	contextSHA, err := validateTypedRecordPreflight(typedPathsFixture(t, "record"), source)
	if err != nil {
		t.Fatal(err)
	}
	feedback, cases := typedPathsSource(t, "mixed-construction-cases.json"), typedPathsSource(t, "mixed-evaluation-cases.json")
	for name, edit := range map[string]nativeSmokeEdit{
		"schema":                    {[]any{"construction", "schema"}, "gooo/joint-construction/v6"},
		"source":                    {[]any{"construction", "original_source_sha256"}, "changed"},
		"scope":                     {[]any{"evaluation", "input_separation", "other_inputs"}, 2},
		"caller baseline":           {[]any{"construction", "attempts", 0, "runtime", "finite_passed"}, 1},
		"local baseline":            {[]any{"construction", "attempts", 0, "local_passed"}, 1},
		"model input":               {[]any{"construction", "initial", "preparations", 1, "generation", "report", "record_assembly", "model_context", "sha256"}, "changed"},
		"decline":                   {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_paths", "model_context", "reason"}, "changed"},
		"missing typed counter":     {[]any{"construction", "initial", "preparations", 0, "generation", "report", "body_paths", "search", "selection", "local_model_predictions"}, nil},
		"caller input":              {[]any{"construction", "attempts", 0, "runtime", "traces", 0, "deliveries", 0, "input"}, 4},
		"new inference":             {[]any{"evaluation", "new_model_calls"}, 1},
		"missing inference counter": {[]any{"evaluation", "new_model_calls"}, nil},
		"rounded input":             {[]any{"evaluation", "runtime", "traces", 2, "deliveries", 0, "input"}, json.Number("9007199254740992")},
		"rounded output":            {[]any{"evaluation", "runtime", "traces", 2, "deliveries", 0, "actual"}, json.Number("27021597764222980")},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validateTypedSmoke(changedNativeSmoke(t, raw, edit), source, feedback, cases, contextSHA, false, ""); err == nil {
				t.Fatal("changed typed native observation accepted")
			}
		})
	}
}

func TestTypedReleaseRejectsChangedPreflight(t *testing.T) {
	raw, source := typedPathsFixture(t, "unary"), typedPathsSource(t, "unary.gooo.fixture")
	for name, edit := range map[string]nativeSmokeEdit{
		"counter":         {[]any{"model_predictions"}, 1},
		"missing counter": {[]any{"candidate_tests"}, nil},
		"emission":        {[]any{"selected_emission"}, true},
		"source":          {[]any{"original_source_sha256"}, "changed"},
		"weights":         {[]any{"model_compatibility", "model", "weights_sha256"}, "changed"},
		"forced ranking":  {[]any{"model_compatibility", "status"}, "READY_FOR_RANKING"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateTypedPreflight(changedNativeSmoke(t, raw, edit), source); err == nil {
				t.Fatal("changed typed preflight accepted")
			}
		})
	}
}

func TestNativeTypedPathsReleaseProfile(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "local-native"}}
	if err := smokeTypedPaths(binary, t.TempDir(), input); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"unary-preflight", "record-preflight", "construct", "replay"} {
		if _, err := os.ReadFile(filepath.Join(input.OutputDir, "local-native-typed-path-"+name+".json")); err != nil {
			t.Fatal("native typed profile did not retain original output", err)
		}
	}
}
