package workgraph

func replayCell(observation Observation, gate GateSpec) Cell {
	if observation.GeneratedDigest == "" || observation.ReplayDigest == "" {
		return unknownCell(gate, "OPERATION_CLASS", "GENERATED_REPLAY_NOT_OBSERVED")
	}
	if observation.GeneratedDigest != observation.ReplayDigest {
		return refutedCell(gate, "GENERATED_REPLAY_DIGEST_MISMATCH", observation.ReplayDigest)
	}
	return closedCell(gate, "GENERATED_REPLAY_DIGEST_MATCH", observation.GeneratedDigest)
}

func artifactCell(observation Observation, gate GateSpec) Cell {
	if observation.GeneratedDigest == "" {
		return unknownCell(gate, "OPERATION_CLASS", "GENERATED_ARTIFACT_NOT_OBSERVED")
	}
	if observation.GeneratedBytes <= 0 {
		return refutedCell(gate, "GENERATED_ARTIFACT_EMPTY", observation.GeneratedDigest)
	}
	return closedCell(gate, "GENERATED_ARTIFACT_OBSERVED", observation.GeneratedDigest)
}

func resourceCell(observation Observation, gate GateSpec) Cell {
	sample := observation.Resource
	if sample.Samples != 1 || sample.WallNanoseconds <= 0 || sample.HeapSysBytes == 0 {
		return unknownCell(gate, "OPERATION_CLASS", "RESOURCE_SAMPLE_NOT_OBSERVED")
	}
	return closedCell(gate, "RESOURCE_SAMPLE_OBSERVED", DigestValue(sample))
}
