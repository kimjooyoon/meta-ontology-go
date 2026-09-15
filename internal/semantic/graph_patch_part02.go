package semantic

import (
	"fmt"
	"strings"
)

func (e GraphPatchConflict) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("%s: %s", ErrGraphPatch, e.Code)
	}
	return fmt.Sprintf("%s: %s: %s", ErrGraphPatch, e.Code, e.Detail)
}
func (e GraphPatchConflict) Unwrap() error { return ErrGraphPatch }

// NodeFieldHash returns a schema-bound digest for a mutable presentation field.
func NodeFieldHash(node Node, field string) (string, error) {
	normalized, err := node.Normalized()
	if err != nil {
		return "", patchConflict(PatchInvalidRequest, "node: "+err.Error())
	}
	var value strings.Builder
	value.WriteString("gooo-graph-node-field/v1\n")
	value.WriteString(field)
	value.WriteByte('\n')
	switch field {
	case "name":
		writeCanonicalField(&value, normalized.Name)
	case "aliases":
		for _, alias := range normalized.Aliases {
			writeCanonicalField(&value, alias)
		}
	default:
		return "", patchConflict(PatchUnknownField, field)
	}
	return StableHashString(value.String()), nil
}

// ValidatePatchPreconditions rejects stale or malformed edits before mutation.
func (g Graph) ValidatePatchPreconditions(base GraphPatchBase, request GraphPatchRequest) error {
	if request.SchemaVersion != GraphPatchSchemaVersion {
		return patchConflict(PatchInvalidRequest, "unsupported schema version")
	}
	if request.Operation != GraphPatchSetNodeField && request.Operation != GraphPatchAddFact {
		return patchConflict(PatchInvalidRequest, "unsupported operation")
	}
	if err := requireDigest(request.ExpectedGraphHash, "expected graph hash"); err != nil {
		return err
	}
	if request.ExpectedGraphHash != g.StableHash() {
		return patchConflict(PatchStaleGraphHash, "expected graph hash does not match")
	}
	if err := validatePatchBase(base, request); err != nil {
		return err
	}
	if strings.TrimSpace(request.AllowedIntent) == "" {
		return patchConflict(PatchIntentMissing, "allowed intent is required")
	}
	if strings.TrimSpace(request.Locality) == "" {
		return patchConflict(PatchLocalityMissing, "locality is required")
	}
	switch request.Operation {
	case GraphPatchSetNodeField:
		return g.validateNodeFieldPatch(request)
	case GraphPatchAddFact:
		return g.validateFactPatch(request)
	default:
		return patchConflict(PatchInvalidRequest, "unsupported operation")
	}
}
