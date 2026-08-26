package symbolicinvocationusecase

import "slices"

func contractFailure(input Input) string {
	receipt, artifact, observation, contract := input.ProducerReceipt, input.ProducerArtifact, input.Observation, input.Contract
	if receipt.Compiler.GoVersion != contract.ExpectedGoVersion ||
		receipt.Compiler.RegisteredEmitters != contract.ExpectedRegisteredEmitters ||
		artifact.Extensions.RegisteredEmitters != contract.ExpectedRegisteredEmitters ||
		!slices.Equal(artifact.Extensions.Kinds, []string{"operation-interface", "operation-manifest", "symbolic-invocation-schema"}) ||
		receipt.Compiler.BinaryBytes < 1 {
		return reasonEvidenceInvalid
	}
	wantSource := SourceCoordinate{
		GoooFiles: contract.ExpectedGoooFiles, GoFiles: contract.ExpectedGoFiles,
		GoooLines: contract.ExpectedGoooLines, Files: contract.ExpectedFiles, Directories: contract.ExpectedDirectories,
	}
	if receipt.Source != wantSource || receipt.Artifact.Kind != artifact.Kind ||
		receipt.Artifact.JSONSchemaDialect != "https://json-schema.org/draft/2020-12/schema" ||
		receipt.Validation.Tool != contract.ExpectedValidator ||
		receipt.Validation.AcceptedInstances != contract.ExpectedAcceptedInstances ||
		receipt.Validation.RejectedInstances != contract.ExpectedRejectedInstances ||
		receipt.DeterministicReplays != contract.ExpectedDeterministicReplays ||
		!canonicalNonClaims(receipt.NotClaimed) || len(receipt.NotClaimed) != contract.ExpectedNonClaims {
		return reasonEvidenceInvalid
	}
	if observation.AcceptedInstances != contract.ExpectedAcceptedInstances ||
		observation.RejectedInstances != contract.ExpectedRejectedInstances ||
		receipt.Validation.GeneratedInstances != contract.ExpectedGeneratedInstances ||
		receipt.Validation.GeneratedGoldenMatches != contract.ExpectedGeneratedGoldenMatches ||
		observation.GeneratedInstances != contract.ExpectedGeneratedInstances ||
		observation.GeneratedGoldenMatches != contract.ExpectedGeneratedGoldenMatches ||
		!validResources(receipt.Resources, contract.ExpectedResourceSamples) {
		return reasonEvidenceInvalid
	}
	return ""
}

func validResources(value ResourceEvidence, expected int) bool {
	if value.SampleCount != expected || len(value.Samples) != expected || value.MaxWallMS < 1 || value.MaxRSSKiB < 1 {
		return false
	}
	maxWall, maxRSS := 0, 0
	for index, sample := range value.Samples {
		if sample.Sequence != index+1 || sample.WallMS < 1 || sample.RSSKiB < 1 {
			return false
		}
		maxWall = max(maxWall, sample.WallMS)
		maxRSS = max(maxRSS, sample.RSSKiB)
	}
	return maxWall == value.MaxWallMS && maxRSS == value.MaxRSSKiB
}
