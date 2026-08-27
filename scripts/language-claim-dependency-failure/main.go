package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

func main() {
	source := flag.String("source", "", "raw .gooo source supplying the edge")
	artifact := flag.String("artifact", "", "raw artifact bound to the failure observation")
	edge := flag.String("edge-id", "", "exact FAILURE_ENTAILMENT edge id")
	outputText := flag.String("output-text-file", "", "captured stdout/stderr from the executed process")
	exitCode := flag.Int("exit-code", 0, "exit code captured from the executed process")
	output := flag.String("output", "", "failure receipt output")
	flag.Parse()
	if *source == "" || *artifact == "" || *edge == "" || *outputText == "" || *output == "" {
		fail("-source, -artifact, -edge-id, -output-text-file, and -output are required")
	}
	sourceBytes, err := os.ReadFile(*source)
	if err != nil {
		fail(err.Error())
	}
	outputBytes, err := os.ReadFile(*outputText)
	if err != nil {
		fail(err.Error())
	}
	receipt, err := claimdependency.BuildFailureReceipt(*source, sourceBytes, *artifact, *edge, string(outputBytes), *exitCode)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*output, receipt)
	fmt.Printf("failure_observation edge=%s exit_code=%s result=%s digest=%s\n", receipt.EdgeID, strconv.Itoa(receipt.ExitCode), receipt.Result, receipt.Digest)
}

func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
