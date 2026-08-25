package artifactemit

const OperationInterfaceKind = "operation-interface"
const OperationInterfaceSchema = "gooo/operation-interface/v1"
const OperationInterfaceResolution = "INTERFACE_ONLY"

func projectOperationInterface(source packageReceipt) Artifact {
	entry := source.Execution.Entry
	return finish(Artifact{
		Schema: OperationInterfaceSchema, Decision: "PASS",
		Resolution: OperationInterfaceResolution,
		Reason: "OPERATION_INTERFACE_EMITTED", Kind: OperationInterfaceKind,
		SubjectDigest: source.Digest,
		Package:       Package{Path: source.PackagePath, Name: source.Package, Namespace: source.Namespace},
		Operation:     Operation{Activity: entry.Activity, Inputs: entry.Inputs, Output: entry.Output},
		Definitions:   DefinitionSet{Language: "gooo", Files: []Definition{}},
		Extensions:    registryReceipt(), Effects: source.Effects,
	})
}

func schemaForKind(kind string) string {
	if kind == OperationInterfaceKind {
		return OperationInterfaceSchema
	}
	return OperationManifestSchema
}
