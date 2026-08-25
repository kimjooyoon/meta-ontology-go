package artifactresolutionexperiment

import "reflect"

func summarize(input Input, indicators []Indicator) Summary {
	satisfied := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		}
	}
	return Summary{
		Coordinates: Coordinates{Satisfied: satisfied, Total: len(indicators),
			BasisPoints: basisPoints(satisfied, len(indicators))},
		Artifacts: ArtifactSummary{Manifest: boolInt(manifestEmitted(input)),
			Interface:     boolInt(interfaceEmitted(input)),
			GoldenMatches: boolInt(reflect.DeepEqual(input.Manifest, input.ManifestGolden)) + boolInt(reflect.DeepEqual(input.Interface, input.InterfaceGolden)),
			Replays:       boolInt(reflect.DeepEqual(input.Manifest, input.ManifestReplay)) + boolInt(reflect.DeepEqual(input.Interface, input.InterfaceReplay))},
		Resolution: ResolutionSummary{ManifestDefinitions: len(input.Manifest.Definitions.Files),
			InterfaceDefinitions: len(input.Interface.Definitions.Files),
			RegisteredEmitters:   registryObserved(input), CoherentOperations: boolInt(operationCoherent(input))},
		Counterexamples: CounterexampleSummary{UnknownEmitterRejections: boolInt(unknownRejected(input))},
		Effects: EffectSummary{RepositoryWrites: input.Manifest.Effects.RepositoryWrites + input.Interface.Effects.RepositoryWrites,
			MutationAuthority: input.Manifest.Effects.MutationAuthority || input.Interface.Effects.MutationAuthority},
		NotClaimed: len(input.Contract.NotClaimed), Unknowns: 0,
	}
}

func basisPoints(satisfied, total int) int {
	if total == 0 {
		return 0
	}
	return satisfied * 10000 / total
}
