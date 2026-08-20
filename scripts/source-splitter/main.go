package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "source-splitter:", err)
		os.Exit(1)
	}
}

func repositorySHA(root string) (string, error) {
	output, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("read repository SHA: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

var errSplitBlocked = errors.New("split blocked")

type splitPart struct {
	Path    string
	Subject string
	Data    []byte
}

type splitPlan struct {
	Directory string
	Mode      os.FileMode
	Parts     []splitPart
}

func stagePart(part splitPart, mode os.FileMode) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(part.Path), ".source-split-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err := file.Chmod(mode); err == nil {
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
