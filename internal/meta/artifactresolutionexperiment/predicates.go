package artifactresolutionexperiment

import (
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func manifestEmitted(input Input) bool {
	return input.Manifest.Schema == input.Contract.ManifestSchema &&
		input.Manifest.Decision == "PASS" && input.Manifest.Resolution == "EXACT" &&
		input.Manifest.Kind == artifactemit.OperationManifestKind &&
		strings.HasPrefix(input.Manifest.Digest, "sha256:")
}

func interfaceEmitted(input Input) bool {
	return input.Interface.Schema == input.Contract.InterfaceSchema &&
		input.Interface.Decision == "PASS" &&
		input.Interface.Resolution == artifactemit.OperationInterfaceResolution &&
		input.Interface.Kind == artifactemit.OperationInterfaceKind &&
		strings.HasPrefix(input.Interface.Digest, "sha256:")
}

func operationCoherent(input Input) bool {
	return reflect.DeepEqual(input.Manifest.Package, input.Interface.Package) &&
		reflect.DeepEqual(input.Manifest.Operation, input.Interface.Operation) &&
		input.Manifest.SubjectDigest == input.Interface.SubjectDigest
}
