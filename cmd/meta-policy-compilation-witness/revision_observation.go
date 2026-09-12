package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func revisionObservationMode(flags *flag.FlagSet) (bool, error) {
	requested := false
	operationRequested := false
	flags.Visit(func(current *flag.Flag) {
		if current.Name == "observe-revision" || current.Name == "revision-operation" {
			requested = true
		}
		if current.Name == "revision-operation" {
			operationRequested = true
		}
	})
	if !requested {
		return false, nil
	}
	allowed := map[string]bool{
		"observe-revision": true, "policy": true, "revision-operation": true,
		"profile-package": true, "profile-namespace": true,
	}
	var modeError error
	flags.Visit(func(current *flag.Flag) {
		if !allowed[current.Name] {
			modeError = fmt.Errorf("revision observation cannot be combined with -%s", current.Name)
		}
	})
	if modeError != nil {
		return true, modeError
	}
	if operationRequested && flags.Lookup("revision-operation").Value.String() == "" {
		return true, errors.New("Gooo revision operation requires a nonempty contract path")
	}
	if flags.NArg() != 0 || flags.Lookup("observe-revision").Value.String() == "" ||
		flags.Lookup("policy").Value.String() == "" {
		return true, errors.New("revision observation requires -policy and a nonempty -observe-revision request, without positional arguments")
	}
	return true, nil
}

func observeRevision(policyPath, requestPath, profilePackage, profileNamespace string, output io.Writer) error {
	requestBytes, err := os.ReadFile(requestPath)
	if err != nil {
		return fmt.Errorf("read revision observation request: %w", err)
	}
	request, err := policycompilation.DecodePolicyRevisionObservationRequest(requestBytes)
	if err != nil {
		return fmt.Errorf("decode revision observation request: %w", err)
	}
	source, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read revision policy: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if *revisionOperationPath != "" {
		return observeBoundRevision(ctx, policyPath, source, requestBytes, profilePackage, profileNamespace, output)
	}
	report, observationError := policycompilation.ObservePolicyDecisionRevision(
		ctx, policyPath, source, profilePackage, profileNamespace, request,
	)
	if report.Schema == "" {
		return observationError
	}
	report.RequestArtifactDigest = policycompilation.DigestBytes(requestBytes)
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return errors.Join(observationError, fmt.Errorf("write revision observation: %w", err))
	}
	return observationError
}
