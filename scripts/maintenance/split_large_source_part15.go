package main

import "os"

func promoteStaged(source string, staged []generatedFile) error {
	committed := make([]string, 0, len(staged))
	for _, stage := range staged {
		destination := string(stage.contents)
		if err := os.Rename(stage.path, destination); err != nil {
			cleanupStaged(staged)
			cleanupPaths(committed)
			return err
		}
		committed = append(committed, destination)
	}
	if err := os.Remove(source); err != nil {
		cleanupPaths(committed)
		return err
	}
	return nil
}

func cleanupStaged(staged []generatedFile) {
	for _, stage := range staged {
		_ = os.Remove(stage.path)
	}
}

func cleanupPaths(paths []string) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
}
