package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/exactsha"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "source-splitter:", err)
		os.Exit(1)
	}
}

func verifyRepositorySHA(root, expected string) error {
	return exactsha.Verify(root, expected)
}

type splitTempCreator func(string, string) (*os.File, error)

func stagePart(part splitPart, mode os.FileMode, create splitTempCreator) (string, error) {
	file, err := create(filepath.Dir(part.Path), ".source-split-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err = file.Chmod(mode); err == nil {
		_, err = file.Write(part.Data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(name)
		return "", errorsJoin(err, closeErr)
	}
	return name, nil
}

func errorsJoin(first, second error) error {
	if first != nil {
		return first
	}
	return second
}
