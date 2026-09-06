package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func main() {
	root := flag.String("root", ".", "read-only input project")
	requestPath := flag.String("request", "", "explicit typed registration request JSON")
	output := flag.String("output", "", "new caller-owned directory outside the project")
	inspect := flag.Bool("inspect", false, "observe input digests without generating a candidate")
	flag.Parse()
	started := time.Now()
	if err := run(*root, *requestPath, *output, *inspect, started); err != nil {
		if failure, ok := errors.AsType[*syntaxregistration.Failure](err); ok {
			raw, _ := json.Marshal(failure)
			fmt.Fprintln(os.Stderr, string(raw))
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root, requestPath, output string, inspect bool, started time.Time) error {
	raw, err := os.ReadFile(requestPath)
	if err != nil {
		return err
	}
	request, err := syntaxregistration.DecodeRequest(raw)
	if err != nil {
		return err
	}
	repository := os.DirFS(root)
	if inspect {
		snapshot, source, err := syntaxregistration.InspectInputs(repository, request)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]string{
			"snapshot_digest": snapshot, "source_digest": source, "toolchain": runtime.Version(),
			"semantic_admission": "UNASSESSED"})
	}
	plan, err := syntaxregistration.Compile(repository, request)
	if err != nil {
		return err
	}
	candidate, err := plan.Generate(repository)
	if err != nil {
		return err
	}
	if err := plan.ValidateCandidate(repository, candidate); err != nil {
		return err
	}
	return exportCandidate(root, output, candidate, time.Since(started).Milliseconds())
}
