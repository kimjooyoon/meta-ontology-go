package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func callbackDependencyFixture(t *testing.T) (callbackListedPackage, callbackListedPackage) {
	t.Helper()
	root := t.TempDir()
	for name, raw := range map[string]string{
		"go.mod":   "module example.invalid/dependency\n\ngo 1.27.3\n",
		"value.go": "package dependency\nconst Value = 7\n",
		"other.go": "package dependency\nconst Other = 8\n",
		"data.txt": "embedded-input",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	module := &callbackListedModule{Path: "example.invalid/dependency", Version: "v1.2.3", Sum: "h1:fixture",
		Replace: &callbackListedModule{Path: root, GoMod: filepath.Join(root, "go.mod")}}
	return callbackListedPackage{Dir: root, ImportPath: module.Path, Module: module, GoFiles: []string{"value.go"},
		EmbedFiles: []string{"data.txt"}}, callbackListedPackage{Dir: root, ImportPath: module.Path + "/transitive",
		Module: module, GoFiles: []string{"other.go"}}
}

func callbackDependencyFixtureJSON(t *testing.T, packages ...callbackListedPackage) []byte {
	t.Helper()
	var raw bytes.Buffer
	for _, item := range packages {
		if err := json.NewEncoder(&raw).Encode(item); err != nil {
			t.Fatal(err)
		}
	}
	return raw.Bytes()
}

func TestCallbackDependencyManifestBindsSelectedBytes(t *testing.T) {
	first, transitive := callbackDependencyFixture(t)
	relocated, relocatedTransitive := callbackDependencyFixture(t)
	observe := func(packages ...callbackListedPackage) CallbackDependencyInputs {
		t.Helper()
		result, err := callbackDependencyInputManifest(context.Background(), callbackDependencyFixtureJSON(t, packages...))
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	original := observe(first, transitive)
	if original.Scope != "SELECTED_EXTERNAL_PACKAGE_FILES" || original.Bytes == 0 || len(original.Packages) != 2 ||
		original.Digest != observe(relocatedTransitive, relocated).Digest {
		t.Fatal("relocation or package ordering changed selected content identity")
	}
	publicJSON, _ := json.Marshal(original)
	if strings.Contains(string(publicJSON), first.Dir) || strings.Contains(string(publicJSON), relocated.Dir) {
		t.Fatal("local checkout location escaped into dependency identity")
	}
	changedSelection := first
	changedSelection.GoFiles = []string{"other.go"}
	if observe(changedSelection, transitive).Digest == original.Digest {
		t.Fatal("different selected build files retained the original identity")
	}
	unselectedTests := first
	unselectedTests.TestGoFiles = []string{"unexecuted_dependency_test.go"}
	unselectedTests.TestEmbedFiles = []string{"unexecuted_test_data.txt"}
	if observe(unselectedTests, transitive).Digest != original.Digest {
		t.Fatal("unused external package tests were treated as selected build inputs")
	}
	for _, name := range []string{"value.go", "other.go", "data.txt", "go.mod"} {
		t.Run(name, func(t *testing.T) {
			file := filepath.Join(first.Dir, name)
			before, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(file, append(bytes.Clone(before), '\n'), 0o600); err != nil {
				t.Fatal(err)
			}
			if observe(first, transitive).Digest == original.Digest {
				t.Fatal("changed dependency bytes retained the original identity")
			}
			if err := os.WriteFile(file, before, 0o600); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCallbackDependencyManifestEnforcesBudgets(t *testing.T) {
	first, _ := callbackDependencyFixture(t)
	file := filepath.Join(first.Dir, "oversized.go")
	if err := os.WriteFile(file, bytes.Repeat([]byte{'x'}, moduleFileMax+1), 0o600); err != nil {
		t.Fatal(err)
	}
	first.GoFiles = []string{"oversized.go"}
	if _, err := callbackDependencyInputManifest(context.Background(), callbackDependencyFixtureJSON(t, first)); err == nil {
		t.Fatal("oversized input file was accepted")
	}
	packages := make([]callbackListedPackage, moduleListPackagesMax+1)
	for index := range packages {
		packages[index] = callbackListedPackage{ImportPath: "runtime", Goroot: true}
	}
	if _, err := callbackDependencyInputManifest(context.Background(), callbackDependencyFixtureJSON(t, packages...)); err == nil {
		t.Fatal("unbounded package selection was accepted")
	}
}

func TestCallbackDependencyObservationClearsInheritedFlags(t *testing.T) {
	t.Setenv("GOFLAGS", "-overlay=unrecorded.json -tags=unrecorded")
	command := exec.Command("go")
	command.Env = callbackObservationEnvironment()
	for _, value := range command.Environ() {
		if strings.HasPrefix(value, "GOFLAGS=") && value != "GOFLAGS=" {
			t.Fatal("unrecorded overlay or build flags survived observer configuration")
		}
	}
}

func TestCallbackDependencyInputDriftPreservesEvidence(t *testing.T) {
	first, _ := callbackDependencyFixture(t)
	before, err := callbackDependencyInputManifest(context.Background(), callbackDependencyFixtureJSON(t, first))
	if err != nil {
		t.Fatal(err)
	}
	first.GoFiles = []string{"other.go"}
	after, err := callbackDependencyInputManifest(context.Background(), callbackDependencyFixtureJSON(t, first))
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][3]CallbackDependencyInputs{{before, before, after}, {before, after, after}} {
		run := CallbackPackageRun{InputsBefore: &pair[1], InputsAfter: &pair[2], ExitCode: 0,
			TestEventsComplete: true, Events: []CallbackPackageTestEvent{{Name: "TestRequired", Action: "pass"}}}
		observation := CallbackExtractionObservation{Decision: "UNKNOWN", DependencyBinding: "UNBOUND",
			OperationAdmission: "UNKNOWN", ApplyPermission: "FORBIDDEN", Runs: []CallbackPackageRun{run}}
		if err := compareCallbackDependencyInputs(&observation, pair[0].Digest, pair[1], pair[2]); err == nil ||
			observation.Decision != "REFUTED" || observation.Frontier.Reason != "SELECTED_DEPENDENCY_INPUTS_CHANGED" ||
			observation.DependencyBinding != "UNBOUND" || observation.OperationAdmission != "UNKNOWN" ||
			observation.ApplyPermission != "FORBIDDEN" || len(observation.Runs[0].Events) != 1 {
			t.Fatal("dependency drift lost its counterexample or changed operation authority")
		}
	}
}

func TestCallbackDependencyManifestRejectsIncompleteInputs(t *testing.T) {
	first, _ := callbackDependencyFixture(t)
	for name, alter := range map[string]func(*callbackListedPackage){
		"missing-module": func(item *callbackListedPackage) { item.Module = nil },
		"incomplete":     func(item *callbackListedPackage) { item.Incomplete = true },
		"escaping-file":  func(item *callbackListedPackage) { item.GoFiles = []string{"../escape.go"} },
		"missing-file":   func(item *callbackListedPackage) { item.GoFiles = []string{"absent.go"} },
	} {
		t.Run(name, func(t *testing.T) {
			item := first
			alter(&item)
			if _, err := callbackDependencyInputManifest(context.Background(), callbackDependencyFixtureJSON(t, item)); err == nil {
				t.Fatal("incomplete dependency acquired a content identity")
			}
		})
	}
	for _, raw := range [][]byte{nil, []byte("{"), callbackDependencyFixtureJSON(t, first, first)} {
		if _, err := callbackDependencyInputManifest(context.Background(), raw); err == nil {
			t.Fatal("invalid dependency selection was accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := callbackDependencyInputManifest(ctx, callbackDependencyFixtureJSON(t, first)); err != context.Canceled {
		t.Fatalf("dependency inspection ignored cancellation: %v", err)
	}
}

func TestCallbackDependencySelectionNativeCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("native dependency selection regression is CI-only")
	}
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                         "module example.invalid/subject\n\ngo 1.27.3\nrequire example.invalid/dependency v1.2.3\nreplace example.invalid/dependency => ./dependency\n",
		"subject_test.go":                "package subject\nimport (\"testing\"; \"example.invalid/dependency\")\nfunc TestRequired(t *testing.T) { if dependency.Value != 7 { t.Fatal(dependency.Value) } }\n",
		"dependency/go.mod":              "module example.invalid/dependency\n\ngo 1.27.3\n",
		"dependency/value.go":            "package dependency\nimport (\"example.invalid/dependency/transitive\"; _ \"embed\")\n//go:embed data.txt\nvar Data string\nconst Value = transitive.Value\n",
		"dependency/data.txt":            "native-embedded-input",
		"dependency/transitive/value.go": "package transitive\nconst Value = 7\n",
	}
	for name, raw := range files {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	inputs, err := callbackObservationDependencyInputs(ctx, root)
	if err != nil || len(inputs.Packages) != 2 {
		t.Fatalf("native selected dependency closure unavailable: %+v, %v", inputs, err)
	}
	for _, item := range inputs.Packages {
		if item.ImportPath == "example.invalid/dependency" && item.Files["data.txt"] == "" {
			t.Fatal("native embed input was omitted")
		}
	}
}
