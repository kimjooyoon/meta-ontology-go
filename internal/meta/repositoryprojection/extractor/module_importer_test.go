package extractor

import (
	"context"
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestModuleImporterResolvesLocalReplacementWithSharedStandardTypes(t *testing.T) {
	t.Setenv("GOWORK", filepath.Join(t.TempDir(), "ambient-workspace.work"))
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("GOFLAGS", "-modfile=/ambient/not-this-repository.mod")
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", `module example.test/main

go 1.27.0

require example.test/dependency v0.0.0
replace example.test/dependency => ./dependency
`)
	writeImporterFixture(t, root, "dependency/go.mod", `module example.test/dependency

go 1.27.0
`)
	writeImporterFixture(t, root, "dependency/dependency.go", `package dependency

import "context"

func Copy(ctx context.Context) context.Context { return ctx }
`)
	writeImporterFixture(t, root, "consumer.go", `package main

import (
	"context"
	"example.test/dependency"
)

var _ context.Context = dependency.Copy(nil)
`)

	imports := newModuleImporter(root)
	dependency, err := imports.Import("example.test/dependency")
	if err != nil {
		t.Fatalf("module dependency import failed: %v", err)
	}
	contextPackage, err := imports.Import("context")
	if err != nil {
		t.Fatalf("standard context import failed: %v", err)
	}
	copyFunction, ok := dependency.Scope().Lookup("Copy").Type().(*types.Signature)
	if !ok {
		t.Fatal("dependency Copy is not a function")
	}
	contextType := contextPackage.Scope().Lookup("Context").Type()
	if !types.Identical(copyFunction.Params().At(0).Type(), contextType) ||
		!types.Identical(copyFunction.Results().At(0).Type(), contextType) {
		t.Fatal("module source and standard importer produced incompatible context types")
	}
	if _, err := imports.Import("net"); err != nil {
		t.Fatalf("standard net package with possible cgo-backed dependencies failed: %v", err)
	}

	fset := token.NewFileSet()
	consumerFile, err := parser.ParseFile(fset, filepath.Join(root, "consumer.go"), nil, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&types.Config{Importer: imports}).Check("example.test/main", fset, []*ast.File{consumerFile}, nil); err != nil {
		t.Fatalf("consumer failed to type-check against shared imports: %v", err)
	}
}

func TestModuleListEnvironmentPinsRepositoryBuildContext(t *testing.T) {
	environment := moduleListEnvironment([]string{
		"GOWORK=/ambient/go.work",
		"GOTOOLCHAIN=auto",
		"GOMODCACHE=/cached/modules",
		"GOPROXY=off",
		"GOWORK=/second/workspace.work",
		"GOROOT=/ambient/go",
		"GOOS=windows",
		"GOARCH=386",
		"CGO_ENABLED=0",
		"GOFLAGS=-tags=windows -modfile=/ambient/go.mod",
		"GOEXPERIMENT=not-a-real-experiment",
		"GOARM=5",
		"GOAMD64=v4",
	})
	values := map[string][]string{}
	for _, entry := range environment {
		name, value, ok := strings.Cut(entry, "=")
		if ok {
			values[name] = append(values[name], value)
		}
	}
	for name, expected := range map[string]string{
		"GOWORK": "off", "GOTOOLCHAIN": "local", "GOROOT": runtime.GOROOT(),
		"GOOS": build.Default.GOOS, "GOARCH": build.Default.GOARCH,
		"CGO_ENABLED": boolSetting(build.Default.CgoEnabled),
		"GOFLAGS":     "", "GOEXPERIMENT": "",
	} {
		if len(values[name]) != 1 || values[name][0] != expected {
			t.Fatalf("repository build setting %s=%q, want %q: %#v", name, values[name], expected, values)
		}
	}
	for name, expected := range moduleListArchitectureSettings() {
		if len(values[name]) != 1 || values[name][0] != expected {
			t.Fatalf("architecture setting %s=%q, want %q: %#v", name, values[name], expected, values)
		}
	}
	if len(values["GOMODCACHE"]) != 1 || values["GOMODCACHE"][0] != "/cached/modules" ||
		len(values["GOPROXY"]) != 1 || values["GOPROXY"][0] != "off" {
		t.Fatalf("Go module cache/proxy settings were not preserved: %#v", values)
	}
}

func TestModuleImporterUsesHostBuildContextDespiteAmbientTargetAndTags(t *testing.T) {
	t.Setenv("GOWORK", filepath.Join(t.TempDir(), "ambient.work"))
	t.Setenv("GOTOOLCHAIN", "auto")
	t.Setenv("GOOS", "windows")
	t.Setenv("GOARCH", "386")
	t.Setenv("CGO_ENABLED", "0")
	t.Setenv("GOFLAGS", "-tags=windows -modfile=/ambient/go.mod")
	t.Setenv("GOEXPERIMENT", "not-a-real-experiment")
	t.Setenv("GOPROXY", "off")
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module example.test/main\n\ngo 1.27.0\n\nrequire example.test/platform v0.0.0\nreplace example.test/platform => ./dep\n")
	hostOS := build.Default.GOOS
	otherOS := "windows"
	if hostOS == otherOS {
		otherOS = "linux"
	}
	writeImporterFixture(t, root, filepath.Join("dep", "go.mod"), "module example.test/platform\n\ngo 1.27.0\n")
	writeImporterFixture(t, root, filepath.Join("dep", "selected_"+hostOS+".go"), "//go:build "+hostOS+"\n\npackage dep\n\nfunc Selected() string { return \"host\" }\n")
	writeImporterFixture(t, root, filepath.Join("dep", "selected_"+otherOS+".go"), "//go:build "+otherOS+"\n\npackage dep\n\nfunc Selected() string { return \"other\" }\n")
	imports := newModuleImporter(root)
	files, err := imports.moduleFiles("example.test/platform")
	if err != nil {
		t.Fatalf("could not list dependency files under repository build context: %v", err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "selected_"+hostOS+".go" {
		t.Fatalf("dependency files=%v, want only selected_%s.go", files, hostOS)
	}
	if _, err := imports.Import("example.test/platform"); err != nil {
		t.Fatalf("dependency did not type-check under repository build context: %v", err)
	}
}

func boolSetting(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func TestModuleImporterRejectsUnresolvedAndCGODependencies(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", `module example.test/main

go 1.27.0

require example.test/cgo v0.0.0
replace example.test/cgo => ./cgo
`)
	writeImporterFixture(t, root, "cgo/go.mod", `module example.test/cgo

go 1.27.0
`)
	writeImporterFixture(t, root, "cgo/cgo.go", `package cgo

import "C"

func Value() int { return 1 }
`)
	imports := newModuleImporter(root)
	if _, err := imports.Import("example.test/cgo"); !errors.Is(err, errModuleDependency) {
		t.Fatalf("CGO package import error=%v, want fail-closed module dependency error", err)
	}
	if _, err := imports.Import("example.test/missing"); !errors.Is(err, errModuleDependency) {
		t.Fatalf("unresolved module import error=%v, want fail-closed module dependency error", err)
	}
	if _, err := imports.Import("-mod=vendor"); !errors.Is(err, errModuleDependency) {
		t.Fatalf("invalid import path error=%v, want fail-closed module dependency error", err)
	}
}

func TestModuleImporterStopsAtCumulativeGoListBudgets(t *testing.T) {
	for name, test := range map[string]struct {
		code      moduleDependencyCauseCode
		configure func(*moduleImporter)
	}{
		"deadline": {moduleCauseTimeBudget, func(imports *moduleImporter) {
			imports.listStarted = time.Now().Add(-moduleListTotalTimeout)
		}},
		"calls": {moduleCauseCallBudget, func(imports *moduleImporter) {
			imports.listCalls = moduleListCallsMax
		}},
		"bytes": {moduleCauseByteBudget, func(imports *moduleImporter) {
			imports.listBytes = moduleListTotalMax
		}},
	} {
		t.Run(name, func(t *testing.T) {
			imports := &moduleImporter{root: t.TempDir(), listStarted: time.Now()}
			test.configure(imports)
			priorCalls := imports.listCalls
			_, err := imports.runGoList([]string{"list", "example.test/unused"})
			requireModuleFailure(t, err, test.code)
			if imports.listCalls != priorCalls {
				t.Fatalf("budget rejection launched a go list call: before=%d after=%d", priorCalls, imports.listCalls)
			}
		})
	}
	imports := &moduleImporter{}
	for range modulePackageMax {
		if !imports.allowExternalPackage() {
			t.Fatal("external package budget rejected an allowed package")
		}
	}
	if imports.allowExternalPackage() {
		t.Fatal("external package budget allowed an extra package")
	}
	blocked := &moduleImporter{externalPackages: modulePackageMax}
	if _, err := blocked.Import("example.test/blocked"); !errors.Is(err, errModuleDependency) {
		t.Fatalf("package budget error=%v, want fail-closed module dependency error", err)
	} else {
		requireModuleFailure(t, err, moduleCausePackageLimit)
	}
}

func TestModuleListFailureCodesAreSanitizedAndPreserved(t *testing.T) {
	raw := errors.New("/private/go env=TOP_SECRET executable failure")
	cases := []struct {
		name string
		err  error
		code moduleDependencyCauseCode
	}{
		{"timeout", moduleListCommandFailure(raw, context.DeadlineExceeded, false, false), moduleCauseTimeout},
		{"output", moduleListCommandFailure(raw, nil, true, false), moduleCauseOutputCap},
		{"subprocess", moduleListCommandFailure(raw, nil, false, false), moduleCauseSubprocess},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			requireModuleFailure(t, test.err, test.code)
			if strings.Contains(test.err.Error(), "/private") || strings.Contains(test.err.Error(), "TOP_SECRET") {
				t.Fatalf("failure leaked raw subprocess details: %v", test.err)
			}
		})
	}
	wrapped := errors.Join(errors.New("wrapper includes /private/path"), cases[0].err)
	requireModuleFailure(t, preserveModuleFailure(wrapped), moduleCauseTimeout)
	if got := preserveModuleFailure(raw); got != errModuleDependency {
		t.Fatalf("untyped failure was not reduced to the generic sentinel: %v", got)
	}
}

func TestModuleListTimeoutBoundsAreExplicit(t *testing.T) {
	if moduleListTimeout != 30*time.Second || moduleListTotalTimeout != 120*time.Second {
		t.Fatalf("go list bounds = %s per call / %s total", moduleListTimeout, moduleListTotalTimeout)
	}
	if moduleListCallsMax != 64 || moduleListTotalMax != 16<<20 {
		t.Fatalf("go list count/byte budgets changed: calls=%d bytes=%d", moduleListCallsMax, moduleListTotalMax)
	}
}

func requireModuleFailure(t *testing.T, err error, code moduleDependencyCauseCode) {
	t.Helper()
	if !errors.Is(err, errModuleDependency) {
		t.Fatalf("error %v is not a module dependency failure", err)
	}
	var failure *moduleDependencyFailure
	if !errors.As(err, &failure) || failure.code != code {
		t.Fatalf("error %v has cause code %v, want %v", err, failure, code)
	}
}

func TestModuleImporterRejectsLocalTraversalAndSymlinkEscapes(t *testing.T) {
	root := t.TempDir()
	writeImporterFixture(t, root, "go.mod", "module example.test/main\n\ngo 1.27.0\n")
	outside := t.TempDir()
	writeImporterFixture(t, outside, "escape.go", "package escape\n\nconst Secret = \"outside\"\n")
	imports := newModuleImporter(root)
	if _, err := imports.Import("example.test/main/../../outside"); !errors.Is(err, errModuleDependency) {
		t.Fatalf("traversal import error=%v, want fail-closed module dependency error", err)
	}

	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}
	if _, err := imports.Import("example.test/main/escape"); !errors.Is(err, errModuleDependency) {
		t.Fatalf("directory symlink import error=%v, want fail-closed module dependency error", err)
	}

	fileRoot := t.TempDir()
	writeImporterFixture(t, fileRoot, "go.mod", "module example.test/filelink\n\ngo 1.27.0\n")
	if err := os.Symlink(filepath.Join(outside, "escape.go"), filepath.Join(fileRoot, "escape.go")); err != nil {
		t.Skipf("file symlinks unavailable: %v", err)
	}
	fileImports := newModuleImporter(fileRoot)
	if _, err := fileImports.Import("example.test/filelink"); !errors.Is(err, errModuleDependency) {
		t.Fatalf("file symlink import error=%v, want fail-closed module dependency error", err)
	}
}

func TestModuleImporterAcceptsRelativeRepositoryRoot(t *testing.T) {
	absoluteRoot := t.TempDir()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativeRoot, err := filepath.Rel(workingDirectory, absoluteRoot)
	if err != nil {
		t.Fatal(err)
	}
	writeImporterFixture(t, absoluteRoot, "go.mod", "module example.test/relative\n\ngo 1.27.0\n")
	writeImporterFixture(t, absoluteRoot, "package.go", "package relative\n\nconst Value = 1\n")
	imports := newModuleImporter(relativeRoot)
	package_, err := imports.Import("example.test/relative")
	if err != nil {
		t.Fatalf("relative repository root import failed: %v", err)
	}
	if package_.Scope().Lookup("Value") == nil {
		t.Fatal("relative repository root did not expose its package declaration")
	}
}

func TestModuleImporterTypeChecksDeclaredDecisionRuntimeFromCurrentModule(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate importer test source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("test did not resolve repository root: %v", err)
	}
	imports := newModuleImporter(root)
	fset := token.NewFileSet()
	source := `package probe

import decision "github.com/kimjooyoon/gooo-decision-runtime"

func use(model *decision.Model, workspace *decision.Workspace, prediction *decision.Prediction) error {
	return model.PredictInto("input + 1", workspace, prediction)
}
`
	file, err := parser.ParseFile(fset, "sdk-consumer.go", source, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&types.Config{Importer: imports}).Check("extractor/sdkprobe", fset, []*ast.File{file}, nil); err != nil {
		t.Fatalf("declared public SDK did not type-check through module importer: %v", err)
	}
}

func writeImporterFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
