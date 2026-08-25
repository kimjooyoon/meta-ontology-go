package languageexampleexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"

func validInput() Input {
	packageValue := artifactemit.Package{Path: "billing-package", Name: "billing", Namespace: "billing"}
	operation := artifactemit.Operation{
		Activity: "PayOrder", Inputs: []artifactemit.Binding{{Name: "Order", ID: "urn:order"}},
		Output: artifactemit.Binding{Name: "Receipt", ID: "urn:receipt"},
	}
	artifact := artifactemit.Artifact{
		Schema: artifactemit.OperationManifestSchema, Decision: "PASS", Resolution: "EXACT",
		Reason: "OPERATION_MANIFEST_EMITTED", Kind: artifactemit.OperationManifestKind,
		Package: packageValue, Operation: operation,
		Definitions: artifactemit.DefinitionSet{Language: "gooo", Files: []artifactemit.Definition{
			{Filename: "activity.gooo", Digest: "sha256:a", DeclarationCount: 1},
			{Filename: "entities.gooo", Digest: "sha256:b", DeclarationCount: 2},
		}},
		Extensions: artifactemit.ExtensionRegistry{RegisteredEmitters: 1, Kinds: []string{"operation-manifest"}},
		Digest: "sha256:artifact",
	}
	contract := Contract{
		Schema: ContractSchema, ID: "fixture", ArtifactSchema: artifactemit.OperationManifestSchema,
		EmitterKind: artifactemit.OperationManifestKind,
		Fixed: Fixed{SourceFiles: 2, PrimaryArtifacts: 1, DeterministicReplays: 1,
			RegisteredEmitters: 1, ResourceSamples: 2, UnknownEmitterRejections: 1,
			Indicators: 13, NotClaimed: 4},
		Limits: Limits{WallMS: 100, RSSKiB: 100, BinaryBytes: 1000},
		NotClaimed: []string{"a", "b", "c", "d"},
	}
	return Input{
		ExpectedHead: "head", Contract: contract, Golden: Golden{Package: packageValue, Operation: operation},
		Artifact: artifact, Replay: artifact,
		UnknownEmitter: artifactemit.Artifact{Schema: artifactemit.OperationManifestSchema,
			Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION", Reason: "EMITTER_UNKNOWN"},
		Profile: Profile{Schema: ProfileSchema, SubjectSHA: "head", ExecutableDigest: "sha256:binary",
			GoooFiles: 2, GoFiles: 0, PrimaryArtifacts: 1, BinaryBytes: 100,
			Samples: []Sample{{Sequence: 1, WallMS: 1, RSSKiB: 1}, {Sequence: 2, WallMS: 2, RSSKiB: 2}}},
	}
}
