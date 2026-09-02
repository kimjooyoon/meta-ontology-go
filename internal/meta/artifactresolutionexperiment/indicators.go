package artifactresolutionexperiment

import "reflect"

func buildIndicators(input Input) []Indicator {
	return []Indicator{
		metric("artifact.manifest", "OUTCOME", "FOUNDATION", "project-full-receipt", boolInt(manifestEmitted(input)), 1),
		metric("artifact.interface", "OUTCOME", "FOUNDATION", "project-public-interface", boolInt(interfaceEmitted(input)), 1),
		metric("golden.manifest", "OUTCOME", "COHERENCE", "compare-manifest-golden", boolInt(reflect.DeepEqual(input.Manifest, input.ManifestGolden)), 1),
		metric("golden.interface", "OUTCOME", "COHERENCE", "compare-interface-golden", boolInt(reflect.DeepEqual(input.Interface, input.InterfaceGolden)), 1),
		metric("replay.manifest", "OUTCOME", "REGRESSION", "replay-full-receipt", boolInt(reflect.DeepEqual(input.Manifest, input.ManifestReplay)), 1),
		metric("replay.interface", "OUTCOME", "REGRESSION", "replay-public-interface", boolInt(reflect.DeepEqual(input.Interface, input.InterfaceReplay)), 1),
		metric("coherence.operation", "OUTCOME", "COHERENCE", "compare-operation-semantics", boolInt(operationCoherent(input)), 1),
		metric("resolution.manifest-definitions", "DRIVER", "FOUNDATION", "count-full-definition-receipts", len(input.Manifest.Definitions.Files), input.Contract.ManifestDefinitions),
		metric("resolution.interface-definitions", "DRIVER", "FOUNDATION", "count-public-definition-receipts", len(input.Interface.Definitions.Files), input.Contract.InterfaceDefinitions),
		metric("compiler.emitter-registry", "DRIVER", "FOUNDATION", "count-resolution-projectors", registryObserved(input), input.Contract.RegisteredEmitters),
		metric("counterexample.unknown-emitter", "GUARDRAIL", "REGRESSION", "reject-unknown-projector", boolInt(unknownRejected(input)), 1),
		metric("guardrail.effects", "GUARDRAIL", "REGRESSION", "deny-repository-effects", effectsObserved(input), 0),
		metric("guardrail.non-claims", "GUARDRAIL", "FOUNDATION", "preserve-resolution-non-claims", len(input.Contract.NotClaimed), input.Contract.NotClaimedCount),
	}
}
