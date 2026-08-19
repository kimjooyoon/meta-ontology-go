package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/conformance/adapter"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func buildProjectionManifest(filename, generatedFile string, source []byte, previous []byte, ir semantic.IR, result generator.Result) (projectionManifest, error) {
	observed, err := observedProjection(result)
	if err != nil {
		return projectionManifest{}, err
	}
	previousSlots, err := ownedSlotBodies(previous)
	if err != nil {
		return projectionManifest{}, err
	}
	currentSlots, err := ownedSlotBodies(result.Source)
	if err != nil {
		return projectionManifest{}, err
	}
	protectedEqual := len(previous) == 0 || equalSlotBodies(previousSlots, currentSlots)
	response := adapter.Response{
		Schema:           adapter.ProtocolSchema,
		Fixture:          filename,
		Operation:        adapter.OperationGenerate,
		Status:           adapter.StatusPass,
		Observed:         observed,
		PromotionEligible: false,
	}
	response.Evidence, err = response.ProjectEvidence("gooo", adapter.StageGoooAuthoritative)
	if err != nil {
		return projectionManifest{}, err
	}
	manifest, err := response.Evidence.Manifest()
	if err != nil {
		return projectionManifest{}, err
	}
	responseDigest, err := response.Digest()
	if err != nil {
		return projectionManifest{}, err
	}
	return projectionManifest{
		Schema:            projectionManifestSchema,
		Producer:          "gooo",
		Operation:         string(adapter.OperationGenerate),
		Fixture:           filename,
		Status:            string(adapter.StatusPass),
		PromotionEligible:  false,
		SemanticDigest:     ir.StableHash(),
		SourceDigest:       semantic.StableHash(source),
		GeneratedDigest:    semantic.StableHash(result.Source),
		SourceMapDigest:    sourceMapDigest(result.SourceMap),
		PreviousGoProvided: len(previous) > 0,
		PreviousGoDigest:   previousDigest(previous),
		ProtectedBytesEqual: protectedEqual,
		GeneratedFile:      generatedFile,
		ResponseDigest:     responseDigest,
		EvidenceManifest:   manifest,
	}, nil
}
