package artifactemit

func projectOperationManifest(source packageReceipt) Artifact {
	definitions := make([]Definition, len(source.Sources))
	for index, definition := range source.Sources {
		definitions[index] = Definition{
			Filename: definition.Filename, Digest: definition.Digest,
			DeclarationCount: definition.DeclarationCount,
		}
	}
	entry := source.Execution.Entry
	return finish(Artifact{
		Schema: OperationManifestSchema, Decision: "PASS", Resolution: "EXACT",
		Reason: "OPERATION_MANIFEST_EMITTED", Kind: OperationManifestKind,
		SubjectDigest: source.Digest,
		Package: Package{Path: source.PackagePath, Name: source.Package, Namespace: source.Namespace},
		Operation: Operation{Activity: entry.Activity, Inputs: entry.Inputs, Output: entry.Output},
		Definitions: DefinitionSet{Language: "gooo", Files: definitions},
		Extensions: registryReceipt(), Effects: source.Effects,
	})
}

func registryReceipt() ExtensionRegistry {
	return ExtensionRegistry{RegisteredEmitters: len(emitterRegistry), Kinds: RegisteredKinds()}
}
