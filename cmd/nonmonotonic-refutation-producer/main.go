package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/nonmonotonicrefutation"
)

func main() {
	sourcePath := flag.String("source", "", "canonical .gooo source")
	outputPath := flag.String("output", "", "producer receipt output")
	repositoryRoot := flag.String("repository-root", ".", "repository root for before/after write observation")
	flag.Parse()
	if *sourcePath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "-source and -output are required")
		os.Exit(2)
	}
	before, err := repositorySnapshot(*repositoryRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	after, err := repositorySnapshot(*repositoryRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	writes := 0
	if !bytes.Equal(before, after) {
		writes = 1
	}
	report, err := nonmonotonicrefutation.Produce(*sourcePath, source, writes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	file, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		_ = file.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := file.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("producer receipt: %s\n", *outputPath)
}

func repositorySnapshot(root string) ([]byte, error) {
	return exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all").Output()
}
