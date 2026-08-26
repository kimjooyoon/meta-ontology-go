package symbolicinvocationusecase

func invocationFacts(input Input) facts {
	receipt, artifact, observation := input.ProducerReceipt, input.ProducerArtifact, input.Observation
	effects := Effects{
		RepositoryWrites:  receipt.Effects.RepositoryWrites + artifact.Effects.RepositoryWrites + observation.Effects.RepositoryWrites,
		MutationAuthority: receipt.Effects.MutationAuthority || artifact.Effects.MutationAuthority || observation.Effects.MutationAuthority,
	}
	return facts{
		UserDecisions:     observation.AcceptedInstances + observation.RejectedInstances,
		AcceptedInstances: observation.AcceptedInstances, RejectedInstances: observation.RejectedInstances,
		GeneratedInstances: observation.GeneratedInstances, GeneratedGoldenMatches: observation.GeneratedGoldenMatches,
		DeterministicReplays: receipt.DeterministicReplays, Source: receipt.Source, Effects: effects,
		Producer: ProducerBinding{
			ReceiptSchema: receipt.Schema, ArtifactSchema: artifact.Schema, ArtifactDigest: artifact.Digest,
			JSONSchemaDigest: receipt.Artifact.JSONSchemaDigest, Validator: receipt.Validation.Tool,
			ValidatorDigest: receipt.Validation.ToolDigest, CompilerBinaryBytes: receipt.Compiler.BinaryBytes,
			CompilerBinaryDigest: receipt.Compiler.BinaryDigest, RegisteredEmitters: receipt.Compiler.RegisteredEmitters,
		},
		Resources: ResourceObservation{
			Mode: "RUNNER_SCOPED_NONDETERMINISTIC", MeasurementReplayAuthority: false,
			Samples: receipt.Resources.SampleCount, MaxWallMS: receipt.Resources.MaxWallMS, MaxRSSKiB: receipt.Resources.MaxRSSKiB,
		},
	}
}
