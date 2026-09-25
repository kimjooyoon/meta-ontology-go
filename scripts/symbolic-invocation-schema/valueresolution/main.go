package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: valueresolution INPUT SUBJECT_SHA OUTPUT")
		os.Exit(2)
	}
	payload, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result, err := artifactemit.CompileSymbolicValueReaderProjection(payload, os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(os.Args[3], encoded, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf(
		"compiler symbolic reader projection: %s %d/%d USER=%d TOOL_AUTHOR=%d GOVERNOR=%d\n",
		result.Decision, result.Coordinates.Satisfied, result.Coordinates.Total,
		result.Readers[0].Coordinates.Total, result.Readers[1].Coordinates.Total,
		result.Readers[2].Coordinates.Total,
	)
	if result.Decision != "PASS" {
		os.Exit(1)
	}
}
