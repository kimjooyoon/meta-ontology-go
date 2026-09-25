package languageexampleexperiment

func rejectDecisions(input Input) (string, string) {
	for _, decision := range []string{input.Artifact.Decision, input.Replay.Decision} {
		switch decision {
		case "PASS":
		case "FAIL_CLOSED":
			return "ARTIFACT_DECISION_REJECTED", "EXACT"
		default:
			return "ARTIFACT_DECISION_UNKNOWN", "LOWER_RESOLUTION"
		}
	}
	switch input.UnknownEmitter.Decision {
	case "PASS", "FAIL_CLOSED":
		return "", ""
	default:
		return "UNKNOWN_EMITTER_DECISION_UNKNOWN", "LOWER_RESOLUTION"
	}
}

func profileEvidenceValid(profile Profile, fixed Fixed) bool {
	if profile.GoooFiles < 0 || profile.GoFiles < 0 || profile.PrimaryArtifacts < 0 ||
		profile.BinaryBytes <= 0 || len(profile.Samples) != fixed.ResourceSamples ||
		profile.Effects.RepositoryWrites < 0 {
		return false
	}
	for index, sample := range profile.Samples {
		if !sampleValid(sample, index) {
			return false
		}
	}
	return true
}

func sampleValid(sample Sample, index int) bool {
	return sample.Sequence == index+1 && sample.WallMS >= 0 && sample.RSSKiB > 0
}

func artifactEffectsValid(input Input) bool {
	return input.Artifact.Effects.RepositoryWrites >= 0 && input.Replay.Effects.RepositoryWrites >= 0 &&
		input.UnknownEmitter.Effects.RepositoryWrites >= 0
}
