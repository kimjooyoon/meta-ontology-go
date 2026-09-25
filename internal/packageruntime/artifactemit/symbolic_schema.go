package artifactemit

func projectSymbolicInvocationSchema(source packageReceipt) Artifact {
	entry := source.Execution.Entry
	prefixItems := make([]ConstSchema, len(entry.Inputs))
	inputIDs := make([]string, len(entry.Inputs))
	for index, input := range entry.Inputs {
		prefixItems[index] = ConstSchema{Const: input.ID}
		inputIDs[index] = input.ID
	}
	return finish(Artifact{
		Schema: SymbolicInvocationArtifact, Decision: "PASS",
		Resolution: SymbolicInvocationResolution,
		Reason:     "SYMBOLIC_INVOCATION_SCHEMA_EMITTED", Kind: SymbolicInvocationSchemaKind,
		SubjectDigest: source.Digest,
		Package:       Package{Path: source.PackagePath, Name: source.Package, Namespace: source.Namespace},
		Operation:     Operation{Activity: entry.Activity, Inputs: entry.Inputs, Output: entry.Output},
		Definitions:   DefinitionSet{Language: "gooo", Files: []Definition{}},
		JSONSchema: &InvocationSchema{
			Dialect: JSONSchemaDraft202012, Title: entry.Activity + " symbolic invocation", Type: "object",
			Properties: InvocationSchemaProperties{
				Activity: ConstSchema{Const: entry.Activity},
				Inputs: TupleSchema{Type: "array", PrefixItems: prefixItems, Items: false,
					MinItems: len(prefixItems), MaxItems: len(prefixItems)},
			},
			Examples: []InvocationExample{{Activity: entry.Activity, Inputs: inputIDs}},
			Required: []string{"activity", "inputs"}, AdditionalProperties: false,
		},
		Extensions: registryReceipt(), Effects: source.Effects,
	})
}
