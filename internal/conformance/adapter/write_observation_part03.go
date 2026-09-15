package adapter

import (
	"fmt"
	"os"
	"strings"
)

func (b ObservationBinding) validate() error {
	if strings.TrimSpace(b.Fixture) == "" ||
		strings.TrimSpace(string(b.Operation)) == "" || strings.TrimSpace(b.RunID) == "" {
		return fmt.Errorf("observation fixture, operation, and run_id are required")
	}
	if !knownOperation(b.Operation) {
		return fmt.Errorf("unsupported observation operation %q", b.Operation)
	}
	return nil
}
func normalizeObserverPaths(paths ObserverPaths) (ObserverPaths, error) {
	values := []string{paths.SourcePath, paths.OutputPath, paths.TempRoot}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return ObserverPaths{}, fmt.Errorf("observer source, output, and temp paths are required")
		}
	}
	var err error
	paths.SourcePath, err = canonicalObserverPath(paths.SourcePath)
	if err != nil {
		return ObserverPaths{}, fmt.Errorf("source path: %w", err)
	}
	paths.OutputPath, err = canonicalObserverPath(paths.OutputPath)
	if err != nil {
		return ObserverPaths{}, fmt.Errorf("output path: %w", err)
	}
	paths.TempRoot, err = canonicalObserverPath(paths.TempRoot)
	if err != nil {
		return ObserverPaths{}, fmt.Errorf("temp root: %w", err)
	}
	if err := validateObserverPathDisjointness(paths); err != nil {
		return ObserverPaths{}, err
	}
	info, err := os.Lstat(paths.TempRoot)
	if err != nil || !info.IsDir() {
		return ObserverPaths{}, fmt.Errorf("temp root must be an existing directory")
	}
	return paths, nil
}
func captureState(paths ObserverPaths) (FilesystemState, error) {
	source, err := captureFile(paths.SourcePath)
	if err != nil {
		return FilesystemState{}, fmt.Errorf("source: %w", err)
	}
	output, err := captureFile(paths.OutputPath)
	if err != nil {
		return FilesystemState{}, fmt.Errorf("output: %w", err)
	}
	temp, err := captureTemp(paths.TempRoot)
	if err != nil {
		return FilesystemState{}, fmt.Errorf("temp: %w", err)
	}
	return FilesystemState{Source: source, Output: output, Temp: temp}, nil
}
func captureFile(path string) (FileObservation, error) {
	observation, err := capturePath(path)
	if err != nil {
		return FileObservation{}, err
	}
	if observation.Exists && observation.Kind != "file" {
		return FileObservation{}, fmt.Errorf("%s is not a regular file", path)
	}
	return observation, nil
}
