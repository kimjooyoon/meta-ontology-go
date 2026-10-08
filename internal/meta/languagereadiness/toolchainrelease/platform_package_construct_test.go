package toolchainrelease

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func packageConstructionFixture(t *testing.T, mode string) []byte {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "package-caller-"+mode+".json.gz"))
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
func packageConstructionTestReference(t *testing.T) packageConstructionReference {
	t.Helper()
	r, err := readPackageConstructionReference("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestPackageCallerReleasePartialCompleteAndSavedReplay(t *testing.T) {
	reference := packageConstructionTestReference(t)
	construct := packageConstructionFixture(t, "construct")
	for _, mode := range []string{"partial", "construct", "replay"} {
		budget := 6
		var saved []byte
		if mode == "partial" {
			budget = 5
		}
		if mode == "replay" {
			saved = construct
		}
		if err := validatePackageConstructionSmoke(packageConstructionFixture(t, mode), reference, budget, saved); err != nil {
			t.Fatal(mode, err)
		}
	}
	if err := validatePackageConstructionSmoke(packageConstructionFixture(t, "partial"), reference, 6, nil); err == nil {
		t.Fatal("partial accepted as complete")
	}
	if err := validatePackageConstructionSmoke(packageConstructionFixture(t, "replay"), reference, 6, nil); err == nil {
		t.Fatal("unbound replay accepted")
	}
	if err := validatePackageConstructionSmoke(construct, reference, 6, construct); err == nil {
		t.Fatal("fresh construction accepted as saved replay")
	}
}

func TestPackageCallerReleaseRejectsChangedOriginalEvidence(t *testing.T) {
	reference := packageConstructionTestReference(t)
	original := packageConstructionFixture(t, "construct")
	for name, edit := range map[string]nativeSmokeEdit{
		"envelope":            {[]any{"schema"}, "changed"},
		"decision":            {[]any{"decision"}, "PARTIAL_FINITE"},
		"manifest":            {[]any{"manifest_digest"}, "sha256:changed"},
		"case file":           {[]any{"cases_digest"}, "sha256:changed"},
		"result":              {[]any{"result", "schema"}, "changed"},
		"package":             {[]any{"result", "program", "entry", "package_path"}, "other"},
		"package source":      {[]any{"result", "program", "workspace_image", "packages", 0, "sources", 0, "source_digest"}, "sha256:changed"},
		"lowered source":      {[]any{"result", "program", "lowered_gooo_source"}, "changed"},
		"original expected":   {[]any{"result", "construction_cases", "cases", 0, "expected", packageCallerKey}, json.Number("8")},
		"translated expected": {[]any{"result", "construction", "construction_cases", "cases", 0, "expected", "GoooPackage1ActivityMain"}, json.Number("8")},
		"budget":              {[]any{"result", "construction", "program_budget"}, json.Number("5")},
		"source binding":      {[]any{"result", "construction", "original_source_sha256"}, "sha256:changed"},
		"fill activity":       {[]any{"result", "construction", "attempts", 0, "fill_candidates", 0, "activity"}, "PlanBudget"},
		"rejection activity":  {[]any{"result", "construction", "attempts", 1, "rejection", "activity"}, "PlanBudget"},
		"local expected":      {[]any{"result", "construction", "attempts", 0, "fill_candidates", 0, "value_case_results", 0, "expected", "next"}, json.Number("2")},
		"native fault left":   {[]any{"result", "construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "fault", "left"}, json.Number("9")},
		"native fault origin": {[]any{"result", "construction", "attempts", 4, "runtime", "traces", 0, "deliveries", 0, "fault", "site", "activity_id"}, "other"},
		"selected source":     {[]any{"result", "construction", "selected_source"}, "changed"},
		"last numerator":      {[]any{"result", "evaluation", "runtime", "finite_passed"}, json.Number("3")},
		"last denominator":    {[]any{"result", "evaluation", "runtime", "finite_total"}, json.Number("3")},
		"new inference":       {[]any{"result", "evaluation", "new_model_calls"}, json.Number("1")},
		"held out input":      {[]any{"result", "evaluation", "runtime", "traces", 3, "deliveries", 0, "input", "used"}, json.Number("9007199254740992")},
		"cancelled run":       {[]any{"result", "evaluation", "runtime", "runs", 0, "canceled"}, true},
	} {
		t.Run(name, func(t *testing.T) {
			raw := changedNativeSmoke(t, original, edit)
			if err := validatePackageConstructionSmoke(raw, reference, 6, nil); err == nil {
				t.Fatal("changed evidence accepted")
			}
		})
	}
}

func TestPackageCallerReleaseKeepsSavedHistoryAndReference(t *testing.T) {
	reference := packageConstructionTestReference(t)
	saved, replay := packageConstructionFixture(t, "construct"), packageConstructionFixture(t, "replay")
	for _, edit := range []nativeSmokeEdit{
		{[]any{"replayed_from_sha256"}, "sha256:changed"},
		{[]any{"result", "replayed_from_sha256"}, "sha256:changed"},
		{[]any{"result", "construction", "elapsed_ns"}, json.Number("1")},
	} {
		if err := validatePackageConstructionSmoke(changedNativeSmoke(t, replay, edit), reference, 6, saved); err == nil {
			t.Fatal("changed saved history accepted")
		}
	}
	reference.Manifest = append(reference.Manifest, '\n')
	if err := validatePackageConstructionSmoke(saved, reference, 6, nil); err == nil {
		t.Fatal("changed original manifest accepted")
	}
}

func TestNativePackageCallerReleaseProfile(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir()}
	input.Target.ID = "local-native"
	if err := smokePackageConstruction(binary, input); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"partial", "construct", "replay"} {
		if info, err := os.Stat(filepath.Join(input.OutputDir, "local-native-package-caller-"+mode+".json")); err != nil || info.Size() == 0 {
			t.Fatal("missing native observation", mode, err)
		}
	}
}
