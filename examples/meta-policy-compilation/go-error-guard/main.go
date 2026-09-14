package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("go-error-guard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	programPath := flags.String("program", "", "caller-owned Gooo guard program")
	sourcePath := flags.String("source", "", "pinned original Go source")
	pipeline := flags.Bool("pipeline", false, "explicit two-activity canonical Go guard pipeline")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *programPath == "" || *sourcePath == "" || flags.NArg() != 0 {
		flags.Usage()
		return 2
	}
	program, err := os.ReadFile(*programPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var report any
	var proposalErr error
	if *pipeline {
		report, proposalErr = policycompilation.ProposeGoErrorGuardPipeline(*programPath, program, source)
	} else {
		report, proposalErr = policycompilation.ProposeGoErrorGuard(*programPath, program, source)
	}
	if err := json.NewEncoder(stdout).Encode(report); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if proposalErr != nil {
		fmt.Fprintln(stderr, proposalErr)
		return 1
	}
	return 0
}
