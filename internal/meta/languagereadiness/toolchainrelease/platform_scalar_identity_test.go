package toolchainrelease

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func scalarIdentityFixture(t *testing.T, mode string) []byte {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "scalar-identity-"+mode+".json.gz"))
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

func scalarIdentityMutations() map[string]nativeSmokeEdit {
	return map[string]nativeSmokeEdit{
		"type name":         {[]any{"composition", "plan", "activities", 0, "inputs", 0, "type"}, "Integer"},
		"type ID":           {[]any{"composition", "plan", "activities", 0, "inputs", 1, "entity_id"}, "urn:gooo:type:integer"},
		"trace":             {[]any{"runtime", "traces"}, nil},
		"output ID":         {[]any{"runtime", "traces", 0, "deliveries", 0, "actual_fields", 0, "id"}, "other://text"},
		"rounded integer":   {[]any{"runtime", "traces", 0, "deliveries", 0, "actual_fields", 2, "value"}, 9007199254740992},
		"runtime inference": {[]any{"runtime", "model_calls"}, 1},
		"partial fields":    {[]any{"composition", "steps", 0, "generation", "report", "record_assembly", "fields_passed"}, 11},
		"lost inference":    {[]any{"composition", "steps", 0, "generation", "report", "record_assembly", "model_calls"}, 0},
		"prediction":        {[]any{"composition", "steps", 0, "generation", "report", "record_assembly", "prediction"}, nil},
		"decline":           {[]any{"composition", "steps", 0, "generation", "report", "record_assembly", "model_context", "status"}, "DECLINED_TO_DETERMINISTIC"},
		"weights":           {[]any{"composition", "steps", 0, "generation", "report", "record_assembly", "model", "weights_sha256"}, "changed"},
	}
}

func TestScalarIdentityReleaseRetainsNamesModelAndExactNativeValues(t *testing.T) {
	names := [3]string{"정수", "논리", "문자열"}
	for _, model := range []bool{false, true} {
		mode := "fixed"
		if model {
			mode = "model"
		}
		selected, err := validateScalarIdentitySmoke(scalarIdentityFixture(t, mode+"-construct"), names, model, false, "")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := validateScalarIdentitySmoke(scalarIdentityFixture(t, mode+"-replay"), names, model, true, selected); err != nil {
			t.Fatal(err)
		}
	}
	raw := scalarIdentityFixture(t, "model-construct")
	for name, edit := range scalarIdentityMutations() {
		t.Run(name, func(t *testing.T) {
			if _, err := validateScalarIdentitySmoke(changedNativeSmoke(t, raw, edit), names, true, false, ""); err == nil {
				t.Fatal("changed scalar identity observation accepted")
			}
		})
	}
}

func TestNativeScalarIdentityReleaseProfile(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "local-native"}}
	if err := smokeScalarIdentity(binary, t.TempDir(), input); err != nil {
		t.Fatal(err)
	}
}
