package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	candidate "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
	v25 "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutioncontract"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

func runCanonicalFixture(program grant.PolicyProgram, settings options) error {
	if settings.outputDir == "" || settings.v24RequestPath == "" || settings.v24ResolutionPath == "" || settings.v24VerificationPath == "" || settings.v25ContractPath == "" || settings.sourceProvenancePath == "" {
		return errors.New("-output-dir, -v24-request, -v24-resolution, -v24-verification, -v25-contract, and -source-provenance are required for canonical-fixture mode")
	}
	var v24Request candidate.AuthorizationRequest
	var v24Resolution candidate.AuthorizationResolution
	var v24Verification candidate.AuthorizationVerification
	var v25Report v25.LiveReport
	var source grant.CanonicalExecutorSourceArtifact
	for _, item := range []struct {
		path   string
		target any
	}{
		{settings.v24RequestPath, &v24Request}, {settings.v24ResolutionPath, &v24Resolution}, {settings.v24VerificationPath, &v24Verification}, {settings.v25ContractPath, &v25Report}, {settings.sourceProvenancePath, &source},
	} {
		if err := readJSON(item.path, item.target); err != nil {
			return fmt.Errorf("read canonical fixture input %s: %w", item.path, err)
		}
	}
	contractProgram, err := v25.CompilePolicy(os.DirFS("."), v25.PolicyPath)
	if err != nil {
		return fmt.Errorf("compile exact v25 contract: %w", err)
	}
	fixture, err := grant.BuildCanonicalExecutorGrantFixture(program, contractProgram, v24Request, v24Resolution, v24Verification, v25Report, source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(settings.outputDir, 0o755); err != nil {
		return err
	}
	materializedBefore, bindingsBefore, placeholdersBefore, err := priorCanonicalFixtureObservations(settings.outputDir)
	if err != nil {
		return err
	}
	artifactNames := grant.CanonicalExecutorArtifactNames()
	for _, name := range artifactNames[:4] {
		if err := writeCanonicalFixtureArtifact(settings.outputDir, name, fixture); err != nil {
			return err
		}
	}
	actualNames, err := canonicalFixtureNames(settings.outputDir)
	if err != nil {
		return err
	}
	if len(actualNames) != 4 {
		return fmt.Errorf("canonical fixture writer observed %d pre-manifest artifacts, want 4", len(actualNames))
	}
	actualNames = append(actualNames, artifactNames[4])
	fixture, err = grant.FinalizeCanonicalExecutorBindingManifest(program, fixture, actualNames, 5, materializedBefore, bindingsBefore, placeholdersBefore)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(settings.outputDir, artifactNames[4]), fixture.Manifest); err != nil {
		return err
	}
	if settings.check {
		if err := checkPersistedCanonicalFixture(program, settings.outputDir, fixture); err != nil {
			return err
		}
	}
	return nil
}

func writeCanonicalFixtureArtifact(dir, name string, fixture grant.CanonicalExecutorGrantFixture) error {
	var value any
	switch name {
	case "canonical-executor-grant-request.json":
		value = fixture.Request
	case "canonical-executor-grant-decision.json":
		value = fixture.Decision
	case "canonical-executor-grant-receipt.json":
		value = fixture.Receipt
	case "canonical-executor-grant-verification.json":
		value = fixture.Verification
	default:
		return fmt.Errorf("unknown canonical fixture artifact %q", name)
	}
	return writeJSON(filepath.Join(dir, name), value)
}

func canonicalFixtureNames(dir string) ([]string, error) {
	known := map[string]bool{}
	for _, name := range grant.CanonicalExecutorArtifactNames()[:4] {
		known[name] = true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && known[entry.Name()] {
			result = append(result, entry.Name())
		}
	}
	sort.Strings(result)
	return result, nil
}

func priorCanonicalFixtureObservations(dir string) (int, int, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, 0, nil
		}
		return 0, 0, 0, err
	}
	manifestPath := filepath.Join(dir, "canonical-executor-grant-binding-manifest.json")
	manifestPresent := false
	known := map[string]bool{}
	for _, name := range grant.CanonicalExecutorArtifactNames() {
		known[name] = true
	}
	for _, entry := range entries {
		if !entry.IsDir() && known[entry.Name()] {
			manifestPresent = manifestPresent || entry.Name() == "canonical-executor-grant-binding-manifest.json"
		}
	}
	if !manifestPresent {
		for _, entry := range entries {
			if !entry.IsDir() && known[entry.Name()] {
				return 0, 0, 0, errors.New("partial prior canonical fixture artifacts are not replaceable")
			}
		}
		return 0, 0, 0, nil
	}
	var manifest grant.CanonicalExecutorBindingManifest
	if err := readJSON(manifestPath, &manifest); err != nil {
		return 0, 0, 0, fmt.Errorf("read prior canonical fixture manifest: %w", err)
	}
	return manifest.MaterializedFixtureCountAfter, manifest.ExactGrantIdentityBindingsAfter, manifest.PlaceholderFieldsAfter, nil
}

func checkPersistedCanonicalFixture(program grant.PolicyProgram, dir string, expected grant.CanonicalExecutorGrantFixture) error {
	var actual grant.CanonicalExecutorGrantFixture
	if err := readJSON(filepath.Join(dir, "canonical-executor-grant-request.json"), &actual.Request); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(dir, "canonical-executor-grant-decision.json"), &actual.Decision); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(dir, "canonical-executor-grant-receipt.json"), &actual.Receipt); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(dir, "canonical-executor-grant-verification.json"), &actual.Verification); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(dir, "canonical-executor-grant-binding-manifest.json"), &actual.Manifest); err != nil {
		return err
	}
	if !reflect.DeepEqual(actual, expected) {
		return errors.New("persisted canonical fixture differs from the materialized fixture")
	}
	if err := grant.ValidateCanonicalExecutorFixture(program, actual); err != nil {
		return err
	}
	return nil
}
