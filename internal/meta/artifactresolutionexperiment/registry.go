package artifactresolutionexperiment

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func registryObserved(input Input) int {
	expectedKinds := []string{artifactemit.OperationInterfaceKind, artifactemit.OperationManifestKind,
		artifactemit.SymbolicInvocationSchemaKind}
	manifest := input.Manifest.Extensions
	public := input.Interface.Extensions
	if manifest.RegisteredEmitters != input.Contract.RegisteredEmitters ||
		public.RegisteredEmitters != input.Contract.RegisteredEmitters ||
		!reflect.DeepEqual(manifest.Kinds, expectedKinds) ||
		!reflect.DeepEqual(public.Kinds, expectedKinds) {
		return -1
	}
	return manifest.RegisteredEmitters
}

func unknownRejected(input Input) bool {
	unknown := input.UnknownEmitter
	return unknown.Decision == "FAIL_CLOSED" && unknown.Resolution == "LOWER_RESOLUTION" &&
		unknown.Reason == "EMITTER_UNKNOWN"
}

func effectsObserved(input Input) int {
	effects := input.Manifest.Effects.RepositoryWrites + input.Interface.Effects.RepositoryWrites
	if input.Manifest.Effects.MutationAuthority || input.Interface.Effects.MutationAuthority {
		effects++
	}
	return effects
}
