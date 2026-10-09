package toolchainrelease

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScalarPreflightRetainsSourceModelAndAssemblyBinding(t *testing.T) {
	for i, example := range scalarIdentityExamples() {
		language := []string{"korean", "english"}[i]
		source, err := os.ReadFile(filepath.Join("../../../..", example.source))
		if err != nil {
			t.Fatal(err)
		}
		raw := scalarIdentityFixture(t, language+"-preflight")
		contextSHA, err := validateScalarModelPreflight(raw, source)
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"construct", "replay"} {
			assembly := scalarIdentityFixture(t, "model-"+mode)
			if err := validateScalarPreflightBinding(assembly, contextSHA); err != nil {
				t.Fatal(err)
			}
			if err := validateScalarPreflightBinding(assembly, "sha256:other"); err == nil {
				t.Fatal("assembly accepted a different preflight input")
			}
		}
		if _, err := validateScalarModelPreflight(raw, append(source, '\n')); err == nil {
			t.Fatal("preflight accepted a different source")
		}
	}
}

func scalarPreflightMutations() map[string]nativeSmokeEdit {
	return map[string]nativeSmokeEdit{
		"export schema":  {[]any{"schema"}, "other"},
		"activity":       {[]any{"activity_id"}, "other"},
		"prediction":     {[]any{"model_predictions"}, 1},
		"tests":          {[]any{"candidate_tests"}, 1},
		"plan exposure":  {[]any{"expanded_plan"}, map[string]any{}},
		"missing model":  {[]any{"model_compatibility"}, nil},
		"not ready":      {[]any{"model_compatibility", "status"}, "DECLINED_TO_DETERMINISTIC"},
		"wrong reason":   {[]any{"model_compatibility", "reason"}, "unknown"},
		"not loaded":     {[]any{"model_compatibility", "model", "loaded"}, false},
		"weights":        {[]any{"model_compatibility", "model", "weights_sha256"}, "other"},
		"metadata":       {[]any{"model_compatibility", "model", "metadata_sha256"}, "other"},
		"model schema":   {[]any{"model_compatibility", "model", "model_schema"}, "other"},
		"model feature":  {[]any{"model_compatibility", "model", "feature_version"}, "other"},
		"model bytes":    {[]any{"model_compatibility", "model", "resident_tensor_bytes"}, 0},
		"arithmetic":     {[]any{"model_compatibility", "model", "arithmetic_version"}, "other"},
		"input status":   {[]any{"context", "status"}, "DECLINED_TO_DETERMINISTIC"},
		"input feature":  {[]any{"context", "feature_version"}, "other"},
		"flow exposure":  {[]any{"value_flow"}, map[string]any{}},
		"context digest": {[]any{"context", "sha256"}, "sha256:other"},
		"context text":   {[]any{"context", "text"}, "changed"},
		"field identity": {[]any{"choices", 0, "field_id"}, "other"},
		"picked early":   {[]any{"choices", 0, "picked"}, "second"},
	}
}

func TestScalarPreflightRejectsMissingAndChangedEvidence(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("../../../..", scalarIdentityExamples()[0].source))
	if err != nil {
		t.Fatal(err)
	}
	raw := scalarIdentityFixture(t, "korean-preflight")
	for name, edit := range scalarPreflightMutations() {
		t.Run(name, func(t *testing.T) {
			if _, err := validateScalarModelPreflight(changedNativeSmoke(t, raw, edit), source); err == nil {
				t.Fatal("changed preflight evidence accepted")
			}
		})
	}
	for _, counter := range []string{"model_predictions", "candidate_tests"} {
		t.Run("absent "+counter, func(t *testing.T) {
			absent := strings.Replace(string(raw), `"`+counter+`":0,`, "", 1)
			if absent == string(raw) {
				t.Fatal("counter was not removed")
			}
			if _, err := validateScalarModelPreflight([]byte(absent), source); err == nil {
				t.Fatal("missing counter accepted as zero")
			}
		})
	}
}
