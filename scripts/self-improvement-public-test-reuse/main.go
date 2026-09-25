package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "", "baseline, reuse, or verify")
	policy := flag.String("policy", "", "canonical .gooo policy and source")
	source := flag.String("source", "", "canonical source .gooo")
	program := flag.String("program", "", "generated program")
	manifest := flag.String("manifest", "", "generated manifest")
	testContract := flag.String("test-contract", "", "generated test contract")
	packageDir := flag.String("package-dir", "", "caller-owned generated package directory")
	out := flag.String("out", "", "caller-owned report directory")
	receipt := flag.String("receipt", "", "caller-owned immutable receipt or receipt directory")
	authorize := flag.Bool("authorize-reuse", false, "explicitly authorize reuse of the receipt")
	verificationManifest := flag.String("verification-manifest", "", "verification input manifest")
	output := flag.String("output", "", "verification report")
	humanOutput := flag.String("human-output", "", "human verification dossier")
	flag.Parse()

	var err error
	switch *mode {
	case "baseline":
		err = runBaseline(executionInput{Policy: *policy, Source: *source, Program: *program, Manifest: *manifest, TestContract: *testContract, PackageDir: *packageDir, OutputDir: *out})
	case "reuse":
		err = runReuse(executionInput{Policy: *policy, Source: *source, Program: *program, Manifest: *manifest, TestContract: *testContract, PackageDir: *packageDir, OutputDir: *out, Receipt: *receipt, Authorize: *authorize})
	case "verify":
		err = runVerification(*verificationManifest, *output, *humanOutput)
	default:
		err = errors.New("usage: self-improvement-public-test-reuse -mode baseline|reuse|verify")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
