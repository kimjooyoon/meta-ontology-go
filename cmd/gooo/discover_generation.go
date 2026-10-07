package main

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type discoveryGeneration struct {
	Path             string `json:"path"`
	Digest           string `json:"artifact_digest"`
	ActivityID       string `json:"activity_id"`
	GeneratedDigest  string `json:"generated_digest"`
	ProjectionReplay bool   `json:"projection_replayed"`
	prior            bodycodegen.Result
	parent           []byte
}

func loadDiscoveryGeneration(reader SourceReader, filename string, source []byte, path string) (*discoveryGeneration, error) {
	if path == "" {
		return nil, nil
	}
	raw, err := reader.ReadFile(path)
	if err != nil {
		return nil, err
	}
	prior, parent, err := bodyexecution.DecodeGeneration(raw)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	assembly, err := bodycodegen.SourceAssembly(ctx, filename, source, prior.Report.Activity)
	if err != nil || assembly == nil {
		return nil, fmt.Errorf("generation replay requires a source-owned assembling declaration: %v", err)
	}
	switch {
	case prior.Report.BodySearch != nil:
		err = bodycodegen.VerifyIRBodySearchProjection(ctx, filename, source, prior)
	case prior.Report.BodyPaths != nil:
		document, decodeErr := bodycodegen.DecodeSourcePathDocument(ctx, filename, source, prior.Report.Activity, nil)
		if decodeErr != nil {
			return nil, decodeErr
		}
		err = bodycodegen.VerifyTypedPathProjection(ctx, filename, source, document, prior)
	default:
		return nil, fmt.Errorf("generation replay supports source-owned IR search and typed path results")
	}
	if err != nil {
		return nil, err
	}
	return &discoveryGeneration{Path: path, Digest: "sha256:" + sha256Hex(raw), ActivityID: prior.Report.ActivityID,
		GeneratedDigest: prior.Report.GeneratedDigest, ProjectionReplay: true, prior: prior, parent: parent}, nil
}

func discoveryGenerationCoverage(actual semantic.IR, inputs map[semantic.ID][]semantic.ID,
	contract *discoveryDomainContract, generation *discoveryGeneration) completeness.CompletenessDimension {
	if generation == nil {
		return unknownCompletenessDimension("generation_coverage", "generated artifacts bound to the discovered declaration", "Capability discovery does not generate code.")
	}
	dimension := completeness.CompletenessDimension{
		ID: "generation_coverage", Status: "UNKNOWN",
		Unit:   "expected domain activity signatures with a source-replayed generated projection",
		Reason: "The saved projection replays; a separate domain contract with expected activities is needed for a coverage denominator.",
		Evidence: []string{"generation_artifact_digest:" + generation.Digest, "generated_digest:" + generation.GeneratedDigest,
			"generated_activity_id:" + generation.ActivityID, "projection_replayed:true", "projection_replay_native_execution:false"},
	}
	if contract == nil {
		return dimension
	}
	actualByID := make(map[semantic.ID]semantic.Node)
	for _, node := range domainDeclarations(actual) {
		actualByID[node.ID] = node
	}
	for _, expected := range domainDeclarations(contract.IR) {
		if expected.Kind != semantic.Activity {
			continue
		}
		dimension.Denominator++
		observed, ok := actualByID[expected.ID]
		if ok && expected.ID.String() == generation.ActivityID &&
			discoveryDeclarationFingerprint(contract.IR, expected, contract.InputSequences) == discoveryDeclarationFingerprint(actual, observed, inputs) {
			dimension.Numerator++
		} else {
			dimension.Evidence = append(dimension.Evidence, "generation_unobserved_for:"+expected.ID.String())
		}
	}
	if dimension.Denominator > 0 {
		dimension.Status = "PROGRESS"
		dimension.Reason = "Some expected domain activities still need a matching source-replayed projection; this axis measures construction coverage."
		if dimension.Numerator == dimension.Denominator {
			dimension.Status = "PASS"
			dimension.Reason = "Every expected domain activity has a matching source-replayed projection; behavior is measured separately."
		}
	}
	return dimension
}
