package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestObserverRejectsOverlappingCanonicalPaths(t *testing.T) {
	tests := overlappingObserverPaths(t)
	request := sampleRequest(StatusFail)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewNoWriteObserver(requestObservationBinding(request), test.paths)
			if err == nil || !strings.Contains(err.Error(), "paths overlap") {
				t.Fatalf("overlapping observer paths were accepted: %v", err)
			}
		})
	}
}
func overlappingObserverPaths(t *testing.T) []struct {
	name  string
	paths ObserverPaths
} {
	t.Helper()
	root := t.TempDir()
	separateTemp := filepath.Join(root, "separate-temp")
	if err := os.Mkdir(separateTemp, 0o700); err != nil {
		t.Fatal(err)
	}
	ancestorTemp := filepath.Join(root, "ancestor", "temp")
	if err := os.MkdirAll(ancestorTemp, 0o700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "source.gooo")
	output := filepath.Join(root, "output.go")
	ancestorSource := filepath.Dir(ancestorTemp)
	descendantSource := filepath.Join(ancestorTemp, "source.gooo")
	return []struct {
		name  string
		paths ObserverPaths
	}{
		{name: "equal source output", paths: ObserverPaths{SourcePath: source, OutputPath: filepath.Join(root, "nested", "..", "source.gooo"), TempRoot: separateTemp}},
		{name: "source ancestor of output", paths: ObserverPaths{SourcePath: filepath.Join(root, "input"), OutputPath: filepath.Join(root, "input", "output.go"), TempRoot: separateTemp}},
		{name: "output ancestor of source", paths: ObserverPaths{SourcePath: filepath.Join(root, "output-dir", "source.gooo"), OutputPath: filepath.Join(root, "output-dir"), TempRoot: separateTemp}},
		{name: "source ancestor of temp", paths: ObserverPaths{SourcePath: ancestorSource, OutputPath: output, TempRoot: ancestorTemp}},
		{name: "temp ancestor of source", paths: ObserverPaths{SourcePath: descendantSource, OutputPath: output, TempRoot: ancestorTemp}},
		{name: "output ancestor of temp", paths: ObserverPaths{SourcePath: source, OutputPath: ancestorSource, TempRoot: ancestorTemp}},
		{name: "temp ancestor of output", paths: ObserverPaths{SourcePath: source, OutputPath: descendantSource, TempRoot: ancestorTemp}},
		{name: "temp equal source", paths: ObserverPaths{SourcePath: ancestorTemp, OutputPath: output, TempRoot: ancestorTemp}},
		{name: "temp equal output", paths: ObserverPaths{SourcePath: source, OutputPath: ancestorTemp, TempRoot: ancestorTemp}},
	}
}
