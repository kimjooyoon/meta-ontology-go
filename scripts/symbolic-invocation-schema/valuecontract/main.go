package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: valuecontract ARTIFACT OUTPUT SUBJECT_SHA")
		os.Exit(64)
	}
	artifactJSON, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read compiler artifact: %v\n", err)
		os.Exit(1)
	}
	contract, err := artifactemit.CompileSymbolicValueContract(artifactJSON, os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "compile symbolic value contract: %v\n", err)
		os.Exit(1)
	}
	encoded, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode symbolic value contract: %v\n", err)
		os.Exit(1)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(os.Args[2], encoded, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write symbolic value contract: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("compiler symbolic value contract: %s %d/%d rules=%d\n", contract.Decision, contract.Coordinates.Satisfied, contract.Coordinates.Total, len(contract.Rules))
}
