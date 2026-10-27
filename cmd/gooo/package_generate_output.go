package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func writePackageLibrary(options packageGenerateOptions, program workspaceexecution.Program, source []byte) (string, error) {
	missing := missingDirectoryChain(options.outputDir)
	published := false
	defer func() {
		if !published {
			_ = cleanupGenerateDirectories(missing)
		}
	}()
	if err := ensureOutputDirectory(options.outputDir); err != nil {
		return "", err
	}
	output, err := resolveOutputPath(options.outputDir, generatedFileName)
	if err != nil {
		return "", err
	}
	if err := protectPackageLibraryInputs(output, options.manifest, program); err != nil {
		return "", err
	}
	before, err := captureOutputSnapshot(output)
	if err != nil {
		return "", err
	}
	write := atomicWrite{path: output, data: source}
	if before.exists {
		write.before = &before
	}
	if err := writeAtomicFiles([]atomicWrite{write}); err != nil {
		return "", err
	}
	published = true
	return output, nil
}

func protectPackageLibraryInputs(output, manifest string, program workspaceexecution.Program) error {
	target, err := os.Stat(output)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	inputs := []string{manifest}
	for _, pkg := range program.Workspace.Packages {
		for _, source := range pkg.Sources {
			inputs = append(inputs, filepath.Join(filepath.Dir(manifest), filepath.FromSlash(source.Filename)))
		}
	}
	for _, input := range inputs {
		original, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("inspect workspace input %q: %w", input, err)
		}
		if os.SameFile(target, original) {
			return fmt.Errorf("generated output would replace workspace input %q", input)
		}
	}
	return nil
}
