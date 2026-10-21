package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publictrust"
)

func main() {
	mode := flag.String("mode", "verify", "verify, generate, or generate-evaluator")
	sourcePath := flag.String("source", publictrust.CanonicalPolicyPath, "canonical .gooo policy")
	readmePath := flag.String("readme", "README.md", "README to verify")
	outputDir := flag.String("output-dir", "", "directory for generated evidence")
	evaluatorPath := flag.String("evaluator-output", "", "generated evaluator path for generate-evaluator mode")
	flag.Parse()
	if *mode == "generate-evaluator" {
		if *evaluatorPath == "" || *evaluatorPath == *sourcePath {
			fmt.Fprintln(os.Stderr, "generate-evaluator requires a distinct -evaluator-output path")
			os.Exit(1)
		}
		source, err := os.ReadFile(*sourcePath)
		if err == nil {
			var generated []byte
			_, generated, err = publictrust.GenerateNamed(*sourcePath, source)
			if err == nil {
				err = os.WriteFile(*evaluatorPath, generated, 0o644)
			}
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := publictrust.Execute(publictrust.RunOptions{Mode: *mode, SourcePath: *sourcePath, ReadmePath: *readmePath, OutputDir: *outputDir}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
