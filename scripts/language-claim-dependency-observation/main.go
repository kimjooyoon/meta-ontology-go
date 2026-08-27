package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

func main() {
	artifact := flag.String("artifact", "", "target artifact to observe")
	expected := flag.String("expected-bytes-digest", "", "digest expected by the target operation")
	output := flag.String("output", "", "observation receipt output")
	flag.Parse()
	if *artifact == "" || *expected == "" || *output == "" {
		fail("-artifact, -expected-bytes-digest, and -output are required")
	}
	data, err := os.ReadFile(*artifact)
	if err != nil {
		fail(err.Error())
	}
	actual := digestBytes(data)
	result, exitCode := "TARGET_CONTRADICTED", 1
	if actual == *expected {
		result, exitCode = "ACCEPTED", 0
	}
	procedureOutput := fmt.Sprintf("read target path=%s bytes=%d sha256=%s expected=%s result=%s", *artifact, len(data), actual, *expected, result)
	receipt, err := claimdependency.BuildObservationReceipt(*artifact, *expected, procedureOutput, exitCode)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*output, receipt)
	fmt.Printf("target_observation result=%s exit_code=%d target_bytes_digest=%s output_digest=%s\n", receipt.Result, receipt.ExitCode, receipt.TargetBytesDigest, receipt.OutputDigest)
}

func digestBytes(data []byte) string {
	return claimdependency.DigestBytesForObservation(data)
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
