package selectiveci

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"strings"
)

func (registry Registry) Validate() error {
	if registry.SchemaVersion != RegistrySchemaVersion {
		return failure(ReasonUnsupportedSchema, "unsupported registry schema_version")
	}
	if !validDigest(registry.PolicyDigest) {
		return failure(ReasonMismatchedDigest, "policy_digest is not SHA-256")
	}
	if registry.Nodes == nil || registry.DependencyEdges == nil || registry.Obligations == nil || registry.Commands == nil || registry.GlobalGuardCommands == nil {
		return failure(ReasonInvalidInput, "registry arrays are required")
	}
	if err := validateNodes(registry.Nodes); err != nil {
		return err
	}
	if err := validateCommands(registry.Commands, registry.GlobalGuardCommands); err != nil {
		return err
	}
	if err := validateBindings(registry.Obligations); err != nil {
		return err
	}
	if registry.Digest != registry.ComputedDigest() {
		return failure(ReasonMismatchedDigest, "registry_digest does not match registry")
	}
	return nil
}
func (registry Registry) ComputedDigest() string {
	copy := normalizeRegistry(registry)
	copy.Digest = ""
	data, err := json.Marshal(copy)
	if err != nil {
		return ""
	}
	return digestBytes(data)
}
func validateNodes(nodes []impactgraph.Node) error {
	seen := map[string]struct{}{}
	for _, node := range nodes {
		if node.ID == "" || strings.TrimSpace(node.ID) != node.ID {
			return failure(ReasonInvalidInput, "node ID is invalid")
		}
		if _, exists := seen[node.ID]; exists {
			return failure(ReasonDuplicateID, "duplicate registry node")
		}
		seen[node.ID] = struct{}{}
	}
	return nil
}
