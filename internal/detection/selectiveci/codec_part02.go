package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"sort"
)

func (result PlanResult) Canonical() string {
	data, err := result.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(data)
}
func (result PlanResult) StableDigest() string {
	return digestBytes([]byte(result.Canonical()))
}
func sealResult(result PlanResult) PlanResult {
	result = normalizeResult(result)
	result.CanonicalDigest = result.StableDigest()
	result.Digest = result.CanonicalDigest
	return result
}
func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func normalizeInput(input Input) Input {
	input.Base = normalizeManifest(input.Base)
	input.Head = normalizeManifest(input.Head)
	input.Registry = normalizeRegistry(input.Registry)
	input.Receipts = append([]Receipt{}, input.Receipts...)
	input.ProvenancePaths = append([]ProvenancePath{}, input.ProvenancePaths...)
	sort.Slice(input.Receipts, func(i, j int) bool { return input.Receipts[i].CommandID < input.Receipts[j].CommandID })
	sort.Slice(input.ProvenancePaths, func(i, j int) bool { return input.ProvenancePaths[i].CommandID < input.ProvenancePaths[j].CommandID })
	for i := range input.ProvenancePaths {
		if normalized, err := input.ProvenancePaths[i].Path.Normalized(); err == nil {
			input.ProvenancePaths[i].Path = normalized
		}
	}
	return input
}
func normalizeManifest(manifest SnapshotManifest) SnapshotManifest {
	manifest.Files = append([]SnapshotFile{}, manifest.Files...)
	for i := range manifest.Files {
		manifest.Files[i].SemanticIDs = sortedCopy(manifest.Files[i].SemanticIDs)
	}
	sort.Slice(manifest.Files, func(i, j int) bool { return manifest.Files[i].Path < manifest.Files[j].Path })
	return manifest
}
func normalizeRegistry(registry Registry) Registry {
	registry.Nodes = append([]impactgraph.Node{}, registry.Nodes...)
	registry.DependencyEdges = append([]DependencyEdge{}, registry.DependencyEdges...)
	registry.Obligations = append([]ObligationBinding{}, registry.Obligations...)
	registry.Commands = append([]Command{}, registry.Commands...)
	registry.GlobalGuardCommands = append([]Command{}, registry.GlobalGuardCommands...)
	sort.Slice(registry.Nodes, func(i, j int) bool { return registry.Nodes[i].ID < registry.Nodes[j].ID })
	sort.Slice(registry.DependencyEdges, func(i, j int) bool {
		return edgeKey(registry.DependencyEdges[i]) < edgeKey(registry.DependencyEdges[j])
	})
	sort.Slice(registry.Obligations, func(i, j int) bool { return registry.Obligations[i].ID < registry.Obligations[j].ID })
	sort.Slice(registry.Commands, func(i, j int) bool { return registry.Commands[i].ID < registry.Commands[j].ID })
	sort.Slice(registry.GlobalGuardCommands, func(i, j int) bool { return registry.GlobalGuardCommands[i].ID < registry.GlobalGuardCommands[j].ID })
	for i := range registry.Obligations {
		registry.Obligations[i].CommandIDs = sortedCopy(registry.Obligations[i].CommandIDs)
	}
	return registry
}
