package artifactresolutionexperiment

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func validInput() Input {
	extensions := artifactemit.ExtensionRegistry{RegisteredEmitters: 3,
		Kinds: []string{artifactemit.OperationInterfaceKind, artifactemit.OperationManifestKind,
			artifactemit.SymbolicInvocationSchemaKind}}
	manifest := artifactemit.Artifact{Schema: artifactemit.OperationManifestSchema,
		Decision: "PASS", Resolution: "EXACT", Reason: "OPERATION_MANIFEST_EMITTED",
		Kind: artifactemit.OperationManifestKind, SubjectDigest: "sha256:subject",
		Package:   artifactemit.Package{Path: "billing", Name: "billing", Namespace: "billing"},
		Operation: artifactemit.Operation{Activity: "PayOrder"},
		Definitions: artifactemit.DefinitionSet{Language: "gooo", Files: []artifactemit.Definition{
			{Filename: "activity.gooo", Digest: "sha256:a", DeclarationCount: 1},
			{Filename: "entities.gooo", Digest: "sha256:b", DeclarationCount: 2}}},
		Extensions: extensions, Effects: artifactemit.Effects{}, Digest: "sha256:manifest"}
	public := manifest
	public.Schema, public.Resolution = artifactemit.OperationInterfaceSchema, artifactemit.OperationInterfaceResolution
	public.Reason, public.Kind = "OPERATION_INTERFACE_EMITTED", artifactemit.OperationInterfaceKind
	public.Definitions = artifactemit.DefinitionSet{Language: "gooo", Files: []artifactemit.Definition{}}
	public.Digest = "sha256:interface"
	unknown := artifactemit.Artifact{Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		Reason: "EMITTER_UNKNOWN"}
	contract := Contract{Schema: ContractSchema, ID: "fixture",
		ManifestSchema: artifactemit.OperationManifestSchema, InterfaceSchema: artifactemit.OperationInterfaceSchema,
		ManifestDefinitions: 2, InterfaceDefinitions: 0, RegisteredEmitters: 3,
		Indicators: ExpectedIndicators, NotClaimedCount: ExpectedNonClaims,
		NotClaimed: []string{"a", "b", "c", "d"}}
	return Input{SubjectSHA: strings.Repeat("a", 40), Contract: contract,
		Manifest: manifest, ManifestReplay: manifest, ManifestGolden: manifest,
		Interface: public, InterfaceReplay: public, InterfaceGolden: public, UnknownEmitter: unknown}
}
