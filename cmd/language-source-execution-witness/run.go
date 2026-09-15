package main

import (
	"flag"
	"fmt"
	"os"

	meta "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagesourceexecution"
)

func run(args []string) int {
	flags := flag.NewFlagSet("language-source-execution-witness", flag.ContinueOnError)
	var head, contractPath, positivePath, replayPath, unknownPath, invalidPath, out string
	flags.StringVar(&head, "head", "", "exact subject commit")
	flags.StringVar(&contractPath, "contract", "", "execution contract")
	flags.StringVar(&positivePath, "positive", "", "positive receipt")
	flags.StringVar(&replayPath, "replay", "", "replay receipt")
	flags.StringVar(&unknownPath, "unknown-entry", "", "unknown entry receipt")
	flags.StringVar(&invalidPath, "invalid-syntax", "", "invalid syntax receipt")
	flags.StringVar(&out, "out", "", "artifact output")
	if flags.Parse(args) != nil || head == "" || contractPath == "" || positivePath == "" ||
		replayPath == "" || unknownPath == "" || invalidPath == "" || out == "" {
		return 2
	}
	contractRaw, err := os.ReadFile(contractPath)
	if err != nil {
		return 2
	}
	contract, err := meta.DecodeContract(contractRaw)
	if err != nil {
		return 2
	}
	read := func(path string) []byte {
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			err = readErr
		}
		return raw
	}
	input := meta.Input{Contract: contract, HeadSHA: head, Positive: read(positivePath),
		Replay: read(replayPath), UnknownEntry: read(unknownPath), InvalidSyntax: read(invalidPath)}
	if err != nil {
		return 2
	}
	artifact := meta.Evaluate(input)
	if err := meta.WriteArtifact(out, artifact); err != nil {
		return 2
	}
	fmt.Printf("source execution: %s %d/%d\n", artifact.Decision,
		artifact.Summary.CasesSatisfied, artifact.Summary.CasesTotal)
	if artifact.Decision != "PASS" {
		return 1
	}
	return 0
}
