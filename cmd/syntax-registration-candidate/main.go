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

type inspectionResponse struct {
	SnapshotDigest    string                               `json:"snapshot_digest"`
	SourceDigest      string                               `json:"source_digest"`
	Toolchain         string                               `json:"toolchain"`
	ExecutionIdentity syntaxregistration.ExecutionIdentity `json:"execution_identity"`
	SemanticAdmission string                               `json:"semantic_admission"`
	PinnedRequest     syntaxregistration.Request           `json:"pinned_request"`
}

func main() {
	root := flag.String("root", ".", "read-only input project")
	requestPath := flag.String("request", "", "explicit typed registration request JSON")
	output := flag.String("output", "", "new caller-owned directory outside the project")
	inspect := flag.Bool("inspect", false, "observe input digests without generating a candidate")
	inventory := flag.Bool("inventory", false, "read source membership without generating or replaying cases")
	flag.Parse()
	if *inventory {
		if *requestPath != "" || *output != "" || *inspect {
			fmt.Fprintln(os.Stderr, "--inventory uses only --root")
			os.Exit(1)
		}
		if err := writeRegistrationInventory(*root, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
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
		identity, err := syntaxregistration.ObserveExecutionIdentity()
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(newInspectionResponse(request, snapshot, source, runtime.Version(), identity))
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

func newInspectionResponse(request syntaxregistration.Request, snapshot, source, toolchain string, identity syntaxregistration.ExecutionIdentity) inspectionResponse {
	request.SnapshotDigest = snapshot
	request.SourceDigest = source
	request.Toolchain = toolchain
	request.ExecutionIdentity = identity
	return inspectionResponse{SnapshotDigest: snapshot, SourceDigest: source, Toolchain: toolchain,
		ExecutionIdentity: identity, SemanticAdmission: "UNASSESSED", PinnedRequest: request}
}
