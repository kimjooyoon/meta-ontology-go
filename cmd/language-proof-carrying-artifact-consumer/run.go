package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	verifier "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageproofartifactverifier"
)

func run(args []string) int {
	flags := flag.NewFlagSet("language-proof-carrying-artifact-consumer", flag.ContinueOnError)
	bundlePath := flags.String("bundle", "", "portable proof bundle")
	reportPath := flags.String("report", "", "independent verifier report")
	target := flags.String("target", "artifact.json", "bundle target permitted by consumer policy")
	out := flags.String("out", "", "consumer receipt output")
	if flags.Parse(args) != nil || *bundlePath == "" || *reportPath == "" || *out == "" {
		return 2
	}
	bundleRaw, err := os.ReadFile(*bundlePath)
	if err != nil {
		return 2
	}
	bundle, err := verifier.DecodeBundle(bundleRaw)
	if err != nil {
		return 1
	}
	report, err := verifier.LoadReport(*reportPath)
	if err != nil {
		return 1
	}
	receipt, err := verifier.ConsumeBundle(bundle, report, *target)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil || os.WriteFile(*out, append(raw, '\n'), 0o644) != nil {
		return 1
	}
	fmt.Printf("read-only consumer: target=%s digest=%s output=%s\n", receipt.TargetPath, receipt.TargetDigest, receipt.OutputDigest)
	return 0
}
