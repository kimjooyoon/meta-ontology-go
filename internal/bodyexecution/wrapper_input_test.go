package bodyexecution

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWrapperDependencyChangeCannotReuseBuild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX wrapper fixture")
	}
	for _, kind := range []string{"scalar", "composition"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("TMPDIR", root)
			tool, dependency := filepath.Join(root, "go-wrapper"), filepath.Join(root, "dependency-state")
			quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }
			wrapper := "#!/bin/sh\nif [ \"$1\" = build ] && [ -f " + quote(dependency) +
				" ]; then echo 'changed tool dependency' >&2; exit 33; fi\nexec " + quote(nativeTool()) + " \"$@\"\n"
			if err := os.WriteFile(tool, []byte(wrapper), 0700); err != nil {
				t.Fatal(err)
			}
			run := wrapperDependencyExecution(t, kind, tool)
			warm, fresh := NewExecutor(), NewExecutor()
			defer warm.Close()
			defer fresh.Close()
			first, artifact, err := run(warm)
			if err != nil || !first.Completed || artifact == nil || artifact.Reused {
				t.Fatal("initial wrapper build failed", first, artifact, err)
			}
			if warm.artifact != nil || artifact.MissReason != "tool_dependencies_unbound" {
				t.Fatal("unbound wrapper output retained for reuse", artifact)
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), "gooo-body-runtime-") ||
					strings.HasPrefix(entry.Name(), "gooo-composition-runtime-") {
					t.Fatal("unbound executable directory was not released", entry.Name())
				}
			}
			if err := os.WriteFile(dependency, []byte("changed"), 0600); err != nil {
				t.Fatal(err)
			}
			control, _, controlErr := run(fresh)
			if controlErr == nil || !control.Started || control.ExitCode == nil || *control.ExitCode != 33 {
				t.Fatal("fresh control did not observe changed dependency", control, controlErr)
			}
			second, artifact, err := run(warm)
			if err == nil || !second.Started || artifact == nil || artifact.Reused ||
				second.ExitCode == nil || *second.ExitCode != 33 {
				t.Fatalf("unchanged wrapper hid its changed input: build=%+v artifact=%+v err=%v", second, artifact, err)
			}
			current, readErr := os.ReadFile(tool)
			if readErr != nil || string(current) != wrapper {
				t.Fatal("counterexample changed wrapper bytes", readErr)
			}
		})
	}
}

func wrapperDependencyExecution(t *testing.T, kind, tool string) func(*Executor) (ProcessObservation, *ArtifactObservation, error) {
	t.Helper()
	if kind == "composition" {
		source, suite, prior := retainedFieldFixture(t)
		return func(owner *Executor) (ProcessObservation, *ArtifactObservation, error) {
			result, err := owner.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, tool)
			return result.Build, result.Artifact, err
		}
	}
	source, doc, prior, parent := fixture(t)
	return func(owner *Executor) (ProcessObservation, *ArtifactObservation, error) {
		result, err := owner.Execute(context.Background(), "fixture.gooo", source, doc, prior, parent, doc.TestCases, tool)
		return result.Observation.Build, result.Observation.Artifact, err
	}
}
