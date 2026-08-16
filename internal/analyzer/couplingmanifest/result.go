package couplingmanifest

import (
	"fmt"

	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func unknownError(code ConstructionCode, format string, args ...any) *ConstructionError {
	return &ConstructionError{Code: code, Message: fmt.Sprintf(format, args...), Status: ConstructionUnknown, FullSuiteRequired: true}
}

func failError(code ConstructionCode, format string, args ...any) *ConstructionError {
	return &ConstructionError{Code: code, Message: fmt.Sprintf(format, args...), Status: ConstructionFailClosed, FullSuiteRequired: true}
}

func constructionError(err error) *ConstructionError {
	if typed, ok := err.(*ConstructionError); ok {
		return typed
	}
	return failError(CodeMalformedBinding, "%v", err)
}

func failedOutput(err *ConstructionError) BuildOutput {
	return BuildOutput{
		Manifest: detector.ChangeManifest{Schema: detector.ManifestSchemaV1, Entries: []detector.ManifestEntry{}},
		Metadata: Metadata{Status: err.Status, Reason: err.Code},
	}
}

func acceptStructuralDetectorResult(result detector.Result) *ConstructionError {
	if result.Status == detector.StatusPass {
		return nil
	}
	if result.Status == detector.StatusUnknown && len(result.Reasons) == 1 && result.Reasons[0].Code == detector.ReasonExternalReceiptMissing {
		return nil
	}
	if len(result.Reasons) == 0 {
		return unknownError(CodeAuthorityDrift, "detector rejected manifest without a reason")
	}
	return &ConstructionError{Code: CodeAuthorityDrift, Message: string(result.Reasons[0].Code) + ": " + result.Reasons[0].Detail, Status: constructionStatus(result.Status), FullSuiteRequired: true}
}

func constructionStatus(status detector.Status) ConstructionStatus {
	if status == detector.StatusFailClosed {
		return ConstructionFailClosed
	}
	return ConstructionUnknown
}

func completeMetadata(sourceMapDigest string, surfaces []detector.Surface, before, head map[semantic.ID]resolved) Metadata {
	ids := make([]semantic.ID, 0, len(surfaces))
	for _, surface := range sortedSurfaces(surfaces) {
		ids = append(ids, surface.SurfaceID)
	}
	return Metadata{
		Status: ConstructionComplete, SourceMapDigest: sourceMapDigest,
		ResolvedSurfaceIDs: ids,
		Counts:             ComponentCounts{Registered: len(surfaces), Before: len(before), Head: len(head), Resolved: len(surfaces)},
		Work:               Work{ComponentCount: len(surfaces), WorkUnits: len(surfaces)},
	}
}
