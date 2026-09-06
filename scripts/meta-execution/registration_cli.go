package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

var registrationMode = flag.String("registration-mode", "", "inspect, plan or worker; empty uses the common executor")
var registrationRequestPath = flag.String("registration-request", "", "explicit typed registration request JSON")
var registrationRoot = flag.String("registration-root", "", "read-only input project root")
var registrationBase = flag.String("registration-base", "", "exact common-plan base commit")
var registrationMetricsPath = flag.String("source-metrics", "", "explicit common source metrics JSON")

func configuredSourceMetricsPath() (string, error) {
	if *registrationMetricsPath != "" {
		return *registrationMetricsPath, nil
	}
	return sourceMetricsPath()
}

func runRegistrationMode() (bool, error) {
	if *registrationMode == "" {
		return false, nil
	}
	if *registrationRoot == "" || *registrationRequestPath == "" {
		return true, fmt.Errorf("registration root and typed request are required")
	}
	raw, err := os.ReadFile(*registrationRequestPath)
	if err != nil {
		return true, err
	}
	request, err := syntaxregistration.DecodeRequest(raw)
	if err != nil {
		return true, err
	}
	value, err := registrationCommand(*registrationMode, request)
	if err != nil {
		if failure, ok := errors.AsType[*syntaxregistration.Failure](err); ok {
			_ = json.NewEncoder(os.Stderr).Encode(failure)
		}
		return true, err
	}
	return true, json.NewEncoder(os.Stdout).Encode(value)
}

func registrationCommand(mode string, request syntaxregistration.Request) (any, error) {
	repository := os.DirFS(*registrationRoot)
	switch mode {
	case "inspect":
		identity, err := syntaxregistration.ObserveExecutionIdentity()
		if err != nil {
			return nil, err
		}
		request.ExecutionIdentity, request.Toolchain = identity, identity.GoVersion
		request.SnapshotDigest, request.SourceDigest, err = syntaxregistration.InspectInputs(repository, request)
		return request, err
	case "worker":
		plan, err := syntaxregistration.Compile(repository, request)
		if err != nil {
			return nil, err
		}
		candidate, err := plan.Generate(repository)
		if err != nil {
			return nil, err
		}
		return candidate, plan.ValidateCandidate(repository, candidate)
	case "plan":
		var metrics linecaps.LineMetricsReport
		if *registrationMetricsPath == "" {
			return nil, fmt.Errorf("registration planning requires explicit source metrics")
		}
		if err := decodeJSON(*registrationMetricsPath, &metrics); err != nil {
			return nil, err
		}
		indicator, err := generation.ObserveRegistrationInput(repository, request)
		if err != nil {
			return nil, err
		}
		metrics.Meta.Indicators = append(metrics.Meta.Indicators, indicator)
		return generation.BuildWithRegistrationInputs(*registrationBase, metrics.CommitSHA, metrics.Meta,
			map[string]syntaxregistration.Request{syntaxregistration.RequestDigest(request): request}), nil
	default:
		return nil, fmt.Errorf("unsupported registration mode %q", mode)
	}
}
