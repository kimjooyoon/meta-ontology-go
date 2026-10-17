package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictestreuse"
)

// Observe the execution directory itself; published presentation artifacts have
// a separate reader and cannot establish execution-directory membership.
func readExecutionPackageInputs(root string, program, contract []byte) (publictestreuse.ExecutionPackageInputs, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return publictestreuse.ExecutionPackageInputs{}, err
	}
	names := publictestreuse.ExecutionPackageNames()
	if len(entries) != len(names) {
		return publictestreuse.ExecutionPackageInputs{}, errors.New("generated test package needs exactly three input files")
	}
	var contents [3][]byte
	for index, entry := range entries {
		if entry.Name() != names[index] {
			return publictestreuse.ExecutionPackageInputs{}, fmt.Errorf("unexpected generated test package entry %q", entry.Name())
		}
		contents[index], err = readRegular(filepath.Join(root, entry.Name()))
		if err != nil {
			return publictestreuse.ExecutionPackageInputs{}, err
		}
	}
	if !bytes.Equal(contents[0], contract) || !bytes.Equal(contents[2], program) {
		return publictestreuse.ExecutionPackageInputs{}, errors.New("generated test package source or contract differs from selected bytes")
	}
	return publictestreuse.ExecutionPackageInputsFromBytes(contents[1], contents[2], contents[0]), nil
}

// Reconstruct the original profile from explicitly renamed publication copies.
// Other presentation artifacts are outside the execution-package input domain.
func readPublishedPackageInputs(root string, program, contract []byte) (publictestreuse.ExecutionPackageInputs, error) {
	names := [3]string{"test-reuse-test-contract.go", "go.mod", "test-reuse-baseline-program.go"}
	var contents [3][]byte
	for index, name := range names {
		data, err := readRegular(filepath.Join(root, name))
		if err != nil {
			return publictestreuse.ExecutionPackageInputs{}, err
		}
		contents[index] = data
	}
	if !bytes.Equal(contents[0], contract) || !bytes.Equal(contents[2], program) {
		return publictestreuse.ExecutionPackageInputs{}, errors.New("published package input mapping differs from selected bytes")
	}
	return publictestreuse.ExecutionPackageInputsFromBytes(contents[1], contents[2], contents[0]), nil
}
