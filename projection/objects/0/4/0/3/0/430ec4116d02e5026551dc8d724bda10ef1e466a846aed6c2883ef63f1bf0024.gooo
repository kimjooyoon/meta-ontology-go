package semantic

import (
	"encoding/hex"
	"strings"
)

func (g Graph) patchNode(raw ID, role string) (Node, error) {
	id, err := ParseIdentity(raw.String())
	if err != nil {
		return Node{}, patchConflict(PatchInvalidRequest, role+" ID: "+err.Error())
	}
	node, ok := g.Node(id)
	if !ok {
		return Node{}, patchConflict(PatchUnknownEndpoint, id.String())
	}
	return node, nil
}
func requireDigest(value, label string) error {
	if len(value) != 64 || value != strings.ToLower(value) {
		return patchConflict(PatchInvalidRequest, label+" must be lowercase SHA-256")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return patchConflict(PatchInvalidRequest, label+" must be lowercase SHA-256")
	}
	return nil
}
func patchConflict(code PatchConflictCode, detail string) error {
	return GraphPatchConflict{Code: code, Detail: detail}
}
