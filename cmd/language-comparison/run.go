package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagecomparison"
)

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("language-comparison", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var subject, runnerLabel, goooFile, goFile, entry, executable, output string
	var samples int
	flags.StringVar(&subject, "subject", "", "exact source commit")
	flags.StringVar(&runnerLabel, "runner", "", "pinned runner image label")
	flags.StringVar(&goooFile, "gooo", "", "Gooo declaration fixture")
	flags.StringVar(&goFile, "go", "", "Go AST baseline fixture")
	flags.StringVar(&entry, "entry", "", "operation name")
	flags.StringVar(&executable, "executable", "", "built comparison executable")
	flags.StringVar(&output, "out", "", "receipt output path")
	flags.IntVar(&samples, "samples", 5, "paired samples per language")
	if err := flags.Parse(args); err != nil || subject == "" || runnerLabel == "" || goooFile == "" ||
		goFile == "" || entry == "" || executable == "" || output == "" || flags.NArg() != 0 {
		usage := "usage: language-comparison -subject <sha> -runner <label> -gooo <file> " +
			"-go <file> -entry <name> -executable <path> -out <path> [-samples <count>]"
		fmt.Fprintln(stderr, usage)
		return 2
	}
	goooSource, err := os.ReadFile(goooFile)
	if err != nil {
		fmt.Fprintf(stderr, "language-comparison: Gooo fixture: %v\n", err)
		return 2
	}
	goSource, err := os.ReadFile(goFile)
	if err != nil {
		fmt.Fprintf(stderr, "language-comparison: Go fixture: %v\n", err)
		return 2
	}
	executableBytes, err := os.ReadFile(executable)
	if err != nil {
		fmt.Fprintf(stderr, "language-comparison: executable: %v\n", err)
		return 2
	}
	executableDigest := sha256.Sum256(executableBytes)
	receipt := languagecomparison.ObserveRuntime(languagecomparison.Request{
		SubjectSHA: subject, ExecutableDigest: "sha256:" + hex.EncodeToString(executableDigest[:]),
		RunnerLabel: runnerLabel, GoooFilename: goooFile, GoooSource: string(goooSource),
		GoFilename: goFile, GoSource: string(goSource), Entry: entry, Samples: samples,
	})
	payload, err := languagecomparison.Marshal(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "language-comparison: receipt: %v\n", err)
		return 1
	}
	if err := os.WriteFile(output, payload, 0o644); err != nil {
		fmt.Fprintf(stderr, "language-comparison: output: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "comparison: %s samples=%d equivalent=%d/%d go/gooo wall=%dppm allocations=%dppm\n",
		receipt.Decision, receipt.Summary.SamplesObserved, receipt.Summary.EquivalentOutputSamples,
		receipt.Summary.SamplesRequested, receipt.Summary.GoToGoooWallRatioPPM,
		receipt.Summary.GoToGoooAllocRatioPPM)
	return 0
}
