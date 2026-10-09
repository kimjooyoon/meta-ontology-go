package toolchainrelease

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlanInputsReleaseProfileOnPackagedCompiler(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "local-native"}}
	if output := os.Getenv("GOOO_PLAN_INPUT_SMOKE_OUTPUT"); output != "" {
		input.OutputDir = output
		if err := os.MkdirAll(output, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := smokePlanInputs(binary, t.TempDir(), input); err != nil {
		t.Fatal(err)
	}
	for _, example := range planInputExamples() {
		for _, mode := range []string{"plan", "template", "inputs", "inputs-replay", "scored-replay"} {
			name := "local-native-" + example.name + "-" + mode + ".json"
			if _, err := os.ReadFile(filepath.Join(input.OutputDir, name)); err != nil {
				t.Fatal("native input journey observation missing", name, err)
			}
		}
	}
}

func TestPlanInspectionReleaseRejectsMissingCountersAndChangedSource(t *testing.T) {
	example := planInputExamples()[0]
	source, err := os.ReadFile(filepath.Join("../../../..", example.source))
	if err != nil {
		t.Fatal(err)
	}
	raw := planInputsFixture(t, "record-korean-plan")
	if _, err := validatePlanInputInspection(raw, source, example); err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]nativeSmokeEdit{
		"missing model calls": {[]any{"model_calls"}, nil},
		"prediction":          {[]any{"model_calls"}, 1},
		"missing tests":       {[]any{"candidate_tests"}, nil},
		"native execution":    {[]any{"native_executions"}, 1},
		"source":              {[]any{"source_sha256"}, "sha256:changed"},
		"caller input":        {[]any{"caller_inputs", 0, "key"}, "Describe"},
		"assembly phase":      {[]any{"assemblies", 0, "phase"}, "called_body"},
		"plan entry":          {[]any{"plan", "entry_activity"}, "Other"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validatePlanInputInspection(changedNativeSmoke(t, raw, edit), source, example); err == nil {
				t.Fatal("changed input inspection accepted")
			}
		})
	}
}
