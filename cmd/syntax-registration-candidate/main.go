package main

import (
	"context"
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
	intentPath := flag.String("intent", "", "compact Gooo-owned case intent JSON")
	output := flag.String("output", "", "new caller-owned directory outside the project")
	inspect := flag.Bool("inspect", false, "observe input digests without generating a candidate")
	inventory := flag.Bool("inventory", false, "read source membership without generating or replaying cases")
	profileHint := flag.String("profile-hint", "", "read-only advisory parser profile hint for one .gooo path")
	flag.Parse()
	if *profileHint != "" {
		if *requestPath != "" || *intentPath != "" || *output != "" || *inspect || *inventory {
			fmt.Fprintln(os.Stderr, "--profile-hint uses only --root")
			os.Exit(1)
		}
		report := syntaxregistration.InspectProfileHint(context.Background(), os.DirFS(*root), *profileHint)
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *inventory {
		if *requestPath != "" || *intentPath != "" || *output != "" || *inspect {
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
	if err := run(*root, *requestPath, *intentPath, *output, *inspect, started); err != nil {
		if failure, ok := errors.AsType[*syntaxregistration.Failure](err); ok {
			raw, _ := json.Marshal(failure)
			fmt.Fprintln(os.Stderr, string(raw))
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root, requestPath, intentPath, output string, inspect bool, started time.Time) error {
	if (requestPath == "") == (intentPath == "") {
		return fmt.Errorf("provide one --request or --intent")
	}
	repository := os.DirFS(root)
	var request syntaxregistration.Request
	if intentPath != "" {
		raw, err := os.ReadFile(intentPath)
		if err != nil {
			return err
		}
		intent, err := syntaxregistration.DecodeCaseIntent(raw)
		if err != nil {
			return err
		}
		request, _, err = syntaxregistration.ComposeCurrentDraft(context.Background(), repository, intent)
		if err != nil {
			return err
		}
	} else {
		raw, err := os.ReadFile(requestPath)
		if err != nil {
			return err
		}
		request, err = syntaxregistration.DecodeRequest(raw)
		if err != nil {
			return err
		}
	}
	if inspect || intentPath != "" {
		if intentPath == "" {
			if err := syntaxregistration.VerifyCaseComposition(context.Background(), request); err != nil {
				return err
			}
		}
		snapshot, source, err := syntaxregistration.InspectInputs(repository, request)
		if err != nil {
			return err
		}
		identity, err := syntaxregistration.ObserveExecutionIdentity()
		if err != nil {
			return err
		}
		response := newInspectionResponse(request, snapshot, source, runtime.Version(), identity)
		if inspect {
			return json.NewEncoder(os.Stdout).Encode(response)
		}
		request = response.PinnedRequest
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
