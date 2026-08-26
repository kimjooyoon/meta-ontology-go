package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "usage: valuereachability ARTIFACT CONTRACT OUTPUT SUBJECT_SHA")
		os.Exit(64)
	}
	artifactJSON, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read compiler artifact: %v\n", err)
		os.Exit(1)
	}
	contractJSON, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read value contract: %v\n", err)
		os.Exit(1)
	}
	reachability, err := artifactemit.CompileSymbolicValueReachability(artifactJSON, contractJSON, os.Args[4])
	if err != nil {
		fmt.Fprintf(os.Stderr, "compile symbolic value reachability: %v\n", err)
		os.Exit(1)
	}
	encoded, err := json.MarshalIndent(reachability, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode symbolic value reachability: %v\n", err)
		os.Exit(1)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(os.Args[3], encoded, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write symbolic value reachability: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("compiler symbolic value reachability: %s %d/%d reachable=%d defense=%d unknown=%d\n",
		reachability.Decision, reachability.Coordinates.Satisfied, reachability.Coordinates.Total,
		reachability.Summary.ReachableRules, reachability.Summary.DefenseOnlyRules+reachability.Summary.DefenseOnlyDefaults,
		reachability.Summary.UnknownPolicyBranches)
}
