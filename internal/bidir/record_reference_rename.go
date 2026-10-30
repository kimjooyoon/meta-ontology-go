package bidir

// Only unchanged, source-bound references follow a display rename. Explicit
// type edits retain their own presentation and pass through normal validation.
// Original spans locate the source being edited; they are not output offsets.
func updateRenamedRecordReferences(source Model, updated *Model) {
	before, after := nodeMap(source.Nodes), nodeMap(updated.Nodes)
	fields := make(map[ID]Field)
	for _, node := range source.Nodes {
		for _, field := range node.Fields {
			fields[field.ID] = field
		}
	}
	for i := range updated.Nodes {
		for j := range updated.Nodes[i].Fields {
			field := &updated.Nodes[i].Fields[j]
			previous, exists := fields[field.ID]
			if !exists || field.TypeRefUse.Form != TypeRefFormLookup ||
				field.TypeRef != previous.TypeRef || field.TypeRefUse != previous.TypeRefUse {
				continue
			}
			targetID := previous.TypeRefUse.ResolvedID
			oldTarget, newTarget := before[targetID], after[targetID]
			if oldTarget.Kind != EntityKind || newTarget.Kind != EntityKind ||
				len(oldTarget.Fields) == 0 || len(newTarget.Fields) == 0 || oldTarget.Name == newTarget.Name {
				continue
			}
			presentation, err := parseLookupTypeRef(previous.TypeRefUse.Spelling)
			if err != nil || presentation.Name != oldTarget.Name {
				continue
			}
			presentation.Name = newTarget.Name
			field.TypeRef.Name = newTarget.Name
			field.TypeRefUse.Spelling = lookupTypeRefSpelling(presentation)
		}
	}
}
