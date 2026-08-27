package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/causalci"
)

func main() {
	inputPath := flag.String("input", "", "causal CI selection input")
	sourcePath := flag.String("source", "", "authoritative .gooo source")
	outputPath := flag.String("output", "", "receipt output")
	verifyPath := flag.String("verify", "", "receipt to independently verify")
	check := flag.Bool("check", false, "run the independent verifier after producing a receipt")
	flag.Parse()
	if *inputPath == "" || *sourcePath == "" {
		fail("-input and -source are required")
	}
	input, err := os.ReadFile(*inputPath)
	if err != nil {
		fail(err.Error())
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	if *verifyPath != "" {
		if *outputPath != "" || *check {
			fail("-verify cannot be combined with -output or -check")
		}
		receiptRaw, err := os.ReadFile(*verifyPath)
		if err != nil {
			fail(err.Error())
		}
		var receipt causalci.Receipt
		if err := json.Unmarshal(receiptRaw, &receipt); err != nil {
			fail(err.Error())
		}
		if err := causalci.Verify(input, *sourcePath, source, receipt); err != nil {
			fail(err.Error())
		}
		fmt.Printf("independent causal CI verifier: PASS cases=%d checks=%d indicators=%d\n", len(receipt.Cases), receipt.Metrics.FixedCheckDenominator, receipt.Metrics.FixedIndicatorDenominator)
		return
	}
	if *outputPath == "" {
		fail("-output is required")
	}
	receipt, err := causalci.Evaluate(input, *sourcePath, source)
	if err != nil {
		fail(err.Error())
	}
	if *check {
		if err := causalci.Verify(input, *sourcePath, source, receipt); err != nil {
			fail(err.Error())
		}
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, data, 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("causal CI selection: decision=%s selected=%d full_fallback=%d rejected=%d checks=%d/%d indicators=%d/%d\n", receipt.Decision, receipt.Metrics.SelectedCheckTotal, receipt.Metrics.FullFallbackCaseTotal, receipt.Metrics.RejectedCaseTotal, receipt.Metrics.FixedCheckDenominator, causalci.FixedCheckDenominator, receipt.Metrics.FixedIndicatorSatisfied, receipt.Metrics.FixedIndicatorDenominator)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
