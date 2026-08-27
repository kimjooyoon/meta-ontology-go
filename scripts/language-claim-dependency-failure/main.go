package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

type comparatorInput struct {
	EdgeID         string                        `json:"edge_id"`
	EdgeKind       claimdependency.EdgeKind      `json:"edge_kind"`
	FromClaimID    string                        `json:"from_claim_id"`
	ToClaimID      string                        `json:"to_claim_id"`
	FromTarget     claimdependency.TargetAddress `json:"from_target"`
	ToTarget       claimdependency.TargetAddress `json:"to_target"`
	FromValue      string                        `json:"from_value_program"`
	ToValue        string                        `json:"to_value_program"`
	ArtifactPath   string                        `json:"artifact_path"`
	ArtifactDigest string                        `json:"artifact_digest"`
}

func main() {
	comparator := flag.Bool("comparator", false, "run the fixed edge-specific comparator child")
	inputPath := flag.String("input", "", "content-addressed comparator input")
	source := flag.String("source", "", "raw .gooo source supplying the edge")
	artifact := flag.String("artifact", "", "raw target artifact consumed by the edge")
	edge := flag.String("edge-id", "", "exact FAILURE_ENTAILMENT edge id")
	output := flag.String("output", "", "failure receipt output")
	flag.Parse()
	if *comparator {
		runComparator(*inputPath, *edge)
	}
	if *source == "" || *artifact == "" || *edge == "" || *output == "" {
		fail("-source, -artifact, -edge-id, and -output are required")
	}
	sourceBytes, err := os.ReadFile(*source)
	if err != nil {
		fail(err.Error())
	}
	artifactBytes, err := os.ReadFile(*artifact)
	if err != nil {
		fail(err.Error())
	}
	graph, err := claimdependency.GraphFromSource(sourceBytes, *source)
	if err != nil {
		fail(err.Error())
	}
	index := -1
	for i, candidate := range graph.Edges {
		if candidate.EdgeID == *edge {
			index = i
		}
	}
	if index < 0 || graph.Edges[index].Kind != claimdependency.FailureEntailment {
		fail("failure receipt requires the exact FAILURE_ENTAILMENT edge E08")
	}
	ed := graph.Edges[index]
	from := claimByID(graph.Nodes, ed.FromClaimID)
	to := claimByID(graph.Nodes, ed.ToClaimID)
	input := comparatorInput{EdgeID: ed.EdgeID, EdgeKind: ed.Kind, FromClaimID: from.ClaimID, ToClaimID: to.ClaimID, FromTarget: from.Target, ToTarget: to.Target, FromValue: from.ValueProgram, ToValue: to.ValueProgram, ArtifactPath: *artifact, ArtifactDigest: claimdependency.DigestBytesForObservation(artifactBytes)}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		fail(err.Error())
	}
	inputFile := *output + ".input.json"
	if err := os.MkdirAll(filepath.Dir(inputFile), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(inputFile, inputBytes, 0o644); err != nil {
		fail(err.Error())
	}
	executable, err := os.Executable()
	if err != nil {
		fail(err.Error())
	}
	executableBytes, err := os.ReadFile(executable)
	if err != nil {
		fail(err.Error())
	}
	argv := []string{"-comparator", "-input", inputFile, "-edge-id", ed.EdgeID}
	command := exec.Command(executable, argv...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()
	exitCode := 0
	if runErr != nil && command.ProcessState != nil {
		exitCode = command.ProcessState.ExitCode()
	}
	if runErr == nil || exitCode != 1 {
		fail(fmt.Sprintf("edge-specific comparator did not observe the required exit 1: %v", runErr))
	}
	receipt, err := claimdependency.BuildFailureReceipt(*source, sourceBytes, *artifact, ed.EdgeID, "CI_EDGE_SPECIFIC_FAILURE_COMPARATOR", executableBytes, argv, stdout.Bytes(), stderr.Bytes(), exitCode)
	if err != nil {
		fail(err.Error())
	}
	writeJSON(*output, receipt)
	fmt.Printf("failure_observation edge=%s observed_exit=%d result=%s digest=%s\n", receipt.EdgeID, receipt.ObservedExitCode, receipt.Result, receipt.Digest)
}

func runComparator(path, edge string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err.Error())
	}
	var input comparatorInput
	if err := json.Unmarshal(data, &input); err != nil {
		fail(err.Error())
	}
	if edge != "E08" || input.EdgeID != "E08" || input.EdgeKind != claimdependency.FailureEntailment || input.FromClaimID == "" || input.ToClaimID == "" || input.FromValue != "claim.edge:contradicts|contradiction-observation-bad" || input.ToValue != "claim.edge:failure-entailment|failure-observation-bad" || input.ArtifactPath == "" || input.ArtifactDigest == "" {
		fail("edge-specific failure antecedent predicate was not observed")
	}
	fmt.Printf("FAILURE_ANTECEDENT_OBSERVED EDGE_SPECIFIC edge=%s from=%s to=%s target=%s\n", input.EdgeID, input.FromClaimID, input.ToClaimID, input.ArtifactPath)
	os.Exit(1)
}

func claimByID(claims []claimdependency.Claim, id string) claimdependency.Claim {
	for _, claim := range claims {
		if claim.ClaimID == id {
			return claim
		}
	}
	fail("edge claim identity missing")
	return claimdependency.Claim{}
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

func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(2) }
