package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictrust"
)

func main() {
	mode := flag.String("mode", "verify", "verify or generate")
	sourcePath := flag.String("source", publictrust.CanonicalPolicyPath, "canonical .gooo policy")
	readmePath := flag.String("readme", "README.md", "README to verify")
	outputDir := flag.String("output-dir", "", "directory for generated evidence")
	flag.Parse()
	if err := publictrust.Execute(publictrust.RunOptions{Mode: *mode, SourcePath: *sourcePath, ReadmePath: *readmePath, OutputDir: *outputDir}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
