package couplingmanifest

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
)

const absentPath = "<absent>"

type observed struct {
	Role          string
	Path          string
	BlobDigest    string
	BindingDigest string
}
type resolved struct {
	Observed observed
}

// BuildOutput keeps detector output and adapter metadata separate. The
// detector result is produced by detector.Evaluate and is never synthesized by
// this package.
type BuildOutput struct {
	Manifest       detector.ChangeManifest
	DetectorResult detector.Result
	Metadata       Metadata
}

// Build constructs the exact detector ChangeManifest. Construction metadata is
// available through BuildDetailed; detector semantic decisions remain outside
// this method.
func Build(input Input) (Manifest, error) {
	output, err := BuildDetailed(input)
	return output.Manifest, err
}

// BuildDetailed constructs a manifest and runs the detector's own structural
// validation path against it. A missing external resource receipt is the
// expected result here because this adapter does not make the detector's final
// coupling decision.
func BuildDetailed(input Input) (BuildOutput, error) {
	if err := validateSnapshots(input); err != nil {
		return failedOutput(err), err
	}
	if err := validateSourceMapContext(input); err != nil {
		return failedOutput(err), err
	}
	beforeDigest, err := rawDigest(input.Before.Digest)
	if err != nil {
		constructionErr := unknownError(CodeInvalidSnapshot, "before snapshot digest is malformed")
		return failedOutput(constructionErr), constructionErr
	}
	headDigest, err := rawDigest(input.Head.Digest)
	if err != nil {
		constructionErr := unknownError(CodeInvalidSnapshot, "head snapshot digest is malformed")
		return failedOutput(constructionErr), constructionErr
	}
	if err := matchSnapshotAuthority(input, beforeDigest, headDigest); err != nil {
		return failedOutput(err), err
	}
	before, head, resolveErr := resolveSnapshots(input)
	if resolveErr != nil {
		return failedOutput(resolveErr), resolveErr
	}
	manifest := makeManifest(input.Authority, beforeDigest, headDigest, before, head)
	detectorResult := ValidateManifest(manifest, input.Authority)
	if err := acceptStructuralDetectorResult(detectorResult); err != nil {
		return BuildOutput{Manifest: manifest, DetectorResult: detectorResult}, err
	}
	metadata := completeMetadata(input.SourceMap.Digest, input.Authority.Registry.Surfaces, before, head)
	return BuildOutput{Manifest: manifest, DetectorResult: detectorResult, Metadata: metadata}, nil
}
