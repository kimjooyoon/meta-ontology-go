package languageexampleexperiment

import "reflect"

func summarize(input Input) Summary {
	artifact, profile := input.Artifact, input.Profile
	summary := Summary{NotClaimed: len(input.Contract.NotClaimed)}
	summary.Value.PrimaryArtifacts = profile.PrimaryArtifacts
	if reflect.DeepEqual(artifact.Package, input.Golden.Package) &&
		reflect.DeepEqual(artifact.Operation, input.Golden.Operation) {
		summary.Value.GoldenMatches = 1
	}
	if artifact.Digest != "" && artifact.Digest == input.Replay.Digest {
		summary.Value.DeterministicReplays = 1
	}
	summary.Compiler.SourceFiles = len(artifact.Definitions.Files)
	summary.Compiler.GoooFiles, summary.Compiler.GoFiles = profile.GoooFiles, profile.GoFiles
	totalDefinitions := profile.GoooFiles + profile.GoFiles
	if artifact.Definitions.Language == "gooo" && totalDefinitions > 0 {
		summary.Compiler.GoooDefinitionBPS = profile.GoooFiles * 10000 / totalDefinitions
	}
	summary.Compiler.RegisteredEmitters = artifact.Extensions.RegisteredEmitters
	summary.Resources.Samples, summary.Resources.BinaryBytes = len(profile.Samples), profile.BinaryBytes
	for _, sample := range profile.Samples {
		if sample.WallMS > summary.Resources.MaxWallMS {
			summary.Resources.MaxWallMS = sample.WallMS
		}
		if sample.RSSKiB > summary.Resources.MaxRSSKiB {
			summary.Resources.MaxRSSKiB = sample.RSSKiB
		}
		if sample.WallMS > input.Contract.Limits.WallMS {
			summary.Resources.WallViolations++
		}
		if sample.RSSKiB > input.Contract.Limits.RSSKiB {
			summary.Resources.RSSViolations++
		}
	}
	if profile.BinaryBytes > input.Contract.Limits.BinaryBytes {
		summary.Resources.BinaryViolations = 1
	}
	unknown := input.UnknownEmitter
	if unknown.Decision == "FAIL_CLOSED" && unknown.Resolution == "LOWER_RESOLUTION" && unknown.Reason == "EMITTER_UNKNOWN" {
		summary.Counterexamples.UnknownEmitterRejections = 1
	}
	summary.Effects.RepositoryWrites = artifact.Effects.RepositoryWrites + input.Replay.Effects.RepositoryWrites +
		unknown.Effects.RepositoryWrites + profile.Effects.RepositoryWrites
	summary.Effects.MutationAuthority = artifact.Effects.MutationAuthority || input.Replay.Effects.MutationAuthority ||
		unknown.Effects.MutationAuthority || profile.Effects.MutationAuthority
	return summary
}
