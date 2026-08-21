package semantic

func validatePatchBase(base GraphPatchBase, request GraphPatchRequest) error {
	if err := requireDigest(request.ExpectedSourceDigest, "expected source digest"); err != nil {
		return err
	}
	if err := requireDigest(request.ExpectedIRDigest, "expected IR digest"); err != nil {
		return err
	}
	if err := requireDigest(base.SourceDigest, "source digest"); err != nil {
		return patchConflict(PatchBaseTupleMismatch, err.Error())
	}
	if err := requireDigest(base.IRDigest, "IR digest"); err != nil {
		return patchConflict(PatchBaseTupleMismatch, err.Error())
	}
	if base.SourceDigest != request.ExpectedSourceDigest || base.IRDigest != request.ExpectedIRDigest {
		return patchConflict(PatchBaseTupleMismatch, "source or IR digest does not match")
	}
	return nil
}
func (g Graph) validateNodeFieldPatch(request GraphPatchRequest) error {
	id, err := ParseIdentity(request.NodeID.String())
	if err != nil {
		return patchConflict(PatchInvalidRequest, "node ID: "+err.Error())
	}
	node, ok := g.Node(id)
	if !ok {
		return patchConflict(PatchUnknownNode, id.String())
	}
	if err := requireDigest(request.ExpectedNodeHash, "expected node hash"); err != nil {
		return err
	}
	if request.ExpectedNodeHash != node.StableHash() {
		return patchConflict(PatchNodeHashMismatch, id.String())
	}
	if request.Field != "name" && request.Field != "aliases" {
		if request.Field == "id" || request.Field == "kind" || request.Field == "namespace" {
			return patchConflict(PatchImmutableField, request.Field)
		}
		return patchConflict(PatchUnknownField, request.Field)
	}
	actual, err := NodeFieldHash(node, request.Field)
	if err != nil {
		return err
	}
	if err := requireDigest(request.ExpectedFieldHash, "expected field hash"); err != nil {
		return err
	}
	if request.ExpectedFieldHash != actual {
		return patchConflict(PatchFieldHashMismatch, request.Field)
	}
	return nil
}
func (g Graph) validateFactPatch(request GraphPatchRequest) error {
	subject, err := g.patchNode(request.Subject, "subject")
	if err != nil {
		return err
	}
	object, err := g.patchNode(request.Object, "object")
	if err != nil {
		return err
	}
	if !request.Predicate.Valid() {
		return patchConflict(PatchInvalidRelation, request.Predicate.String())
	}
	if err := request.Predicate.ValidateKinds(subject.Kind, object.Kind); err != nil {
		return patchConflict(PatchEndpointKindMismatch, err.Error())
	}
	return nil
}
