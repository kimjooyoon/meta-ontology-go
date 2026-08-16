package selectiveci

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type validationError struct {
	reason string
	err    error
}

func (e *validationError) Error() string { return e.reason + ": " + e.err.Error() }

func failure(reason, message string) error {
	return &validationError{reason: reason, err: fmt.Errorf("%s", message)}
}

func reasonFor(err error) string {
	if typed, ok := err.(*validationError); ok {
		return typed.reason
	}
	return ReasonInvalidInput
}

func (input Input) Validate() error {
	if input.SchemaVersion != SchemaVersion {
		return failure(ReasonUnsupportedSchema, "unsupported schema_version")
	}
	if err := input.Base.Validate(); err != nil {
		return err
	}
	if err := input.Head.Validate(); err != nil {
		return err
	}
	if input.CPUCapacity == 0 {
		return failure(ReasonInvalidInput, "cpu_capacity must be positive")
	}
	if input.Receipts == nil || input.ProvenancePaths == nil {
		return failure(ReasonInvalidInput, "resource_receipts and provenance_paths are required")
	}
	if err := input.Registry.Validate(); err != nil {
		return err
	}
	if err := validateReceipts(input.Receipts); err != nil {
		return err
	}
	return validateProvenancePaths(input.ProvenancePaths)
}

func (manifest SnapshotManifest) Validate() error {
	if manifest.SchemaVersion != ManifestSchemaVersion {
		return failure(ReasonUnsupportedSchema, "unsupported manifest schema_version")
	}
	if manifest.Files == nil {
		return failure(ReasonInvalidInput, "manifest files are required")
	}
	seen := map[string]struct{}{}
	for _, file := range manifest.Files {
		if !validRepoPath(file.Path) {
			return failure(ReasonUnknownPath, "manifest path is not normalized")
		}
		if _, exists := seen[file.Path]; exists {
			return failure(ReasonDuplicateID, "duplicate manifest path")
		}
		seen[file.Path] = struct{}{}
		if !validDigest(file.BlobDigest) {
			return failure(ReasonMismatchedDigest, "blob_digest is not SHA-256")
		}
		if err := validateIDs(file.SemanticIDs); err != nil {
			return err
		}
		if len(file.SemanticIDs) != len(sortedUnique(file.SemanticIDs)) {
			return failure(ReasonDuplicateID, "duplicate semantic ID in manifest file")
		}
	}
	if manifest.Digest != manifest.ComputedDigest() {
		return failure(ReasonMismatchedDigest, "snapshot_digest does not match manifest")
	}
	return nil
}

func (manifest SnapshotManifest) ComputedDigest() string {
	canonical, err := json.Marshal(normalizeManifest(manifest))
	if err != nil {
		return ""
	}
	var value struct {
		SchemaVersion string         `json:"schema_version"`
		Files         []SnapshotFile `json:"files"`
	}
	value.SchemaVersion, value.Files = manifest.SchemaVersion, normalizeManifest(manifest).Files
	canonical, err = json.Marshal(value)
	if err != nil {
		return ""
	}
	return digestBytes(canonical)
}

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

func validateCommands(commands, guards []Command) error {
	seen := map[string]struct{}{}
	for _, command := range append(append([]Command(nil), commands...), guards...) {
		if !validStableID(command.ID) {
			return failure(ReasonInvalidInput, "command ID is invalid")
		}
		if _, exists := seen[command.ID]; exists {
			return failure(ReasonDuplicateID, "duplicate command ID")
		}
		seen[command.ID] = struct{}{}
		if len(command.Argv) == 0 || command.Argv[0] == "" {
			return failure(ReasonInvalidArgv, "argv must contain an executable")
		}
		for _, arg := range command.Argv {
			if strings.IndexByte(arg, 0) >= 0 {
				return failure(ReasonInvalidArgv, "argv contains NUL")
			}
		}
		if !validRepoDir(command.WorkingDir) {
			return failure(ReasonUnknownPath, "working directory is not normalized")
		}
		if command.CPUWorkUnits == 0 || command.MemoryBytes == 0 {
			return failure(ReasonInvalidInput, "command resource ceiling is zero")
		}
	}
	return nil
}

func validateBindings(bindings []ObligationBinding) error {
	seen := map[string]struct{}{}
	for _, binding := range bindings {
		if !validStableID(binding.ID) || !validStableID(binding.Subject) {
			return failure(ReasonMissingBinding, "obligation binding identity is invalid")
		}
		if _, exists := seen[binding.ID]; exists {
			return failure(ReasonDuplicateID, "duplicate obligation ID")
		}
		seen[binding.ID] = struct{}{}
		if len(binding.CommandIDs) == 0 {
			return failure(ReasonMissingBinding, "obligation has no commands")
		}
		if len(binding.CommandIDs) != len(sortedUnique(binding.CommandIDs)) {
			return failure(ReasonDuplicateID, "duplicate command reference")
		}
		if err := validateIDs(binding.CommandIDs); err != nil {
			return failure(ReasonDanglingReference, err.Error())
		}
	}
	return nil
}

func validateReceipts(receipts []Receipt) error {
	seen := map[string]struct{}{}
	for _, receipt := range receipts {
		if !validStableID(receipt.CommandID) {
			return failure(ReasonDanglingReference, "resource receipt command ID is invalid")
		}
		if _, exists := seen[receipt.CommandID]; exists {
			return failure(ReasonDuplicateID, "duplicate resource receipt")
		}
		seen[receipt.CommandID] = struct{}{}
		if !validDigest(receipt.SnapshotDigest) {
			return failure(ReasonMismatchedDigest, "resource receipt snapshot digest is invalid")
		}
	}
	return nil
}

func validateProvenancePaths(paths []ProvenancePath) error {
	seenCommands := map[string]struct{}{}
	seenPaths := map[string]struct{}{}
	for _, path := range paths {
		if !validStableID(path.CommandID) || !validStableID(path.Requirement.PathID) {
			return failure(ReasonAmbiguousPath, "provenance identity is invalid")
		}
		if _, exists := seenCommands[path.CommandID]; exists {
			return failure(ReasonAmbiguousPath, "multiple provenance paths for command")
		}
		seenCommands[path.CommandID] = struct{}{}
		if _, exists := seenPaths[path.Requirement.PathID]; exists {
			return failure(ReasonAmbiguousPath, "duplicate provenance path ID")
		}
		seenPaths[path.Requirement.PathID] = struct{}{}
		if len(path.Requirement.RecordIDs) == 0 || len(path.Requirement.RecordIDs) != len(path.Requirement.ExpectedKinds) {
			return failure(ReasonAmbiguousPath, "provenance requirement is not finite")
		}
		if err := path.Path.Validate(); err != nil {
			return failure(ReasonAmbiguousPath, err.Error())
		}
		if err := validateIDs(append(append([]string{}, path.Requirement.RecordIDs...), path.Requirement.StartID, path.Requirement.EndID)); err != nil {
			return failure(ReasonAmbiguousPath, err.Error())
		}
	}
	return nil
}

func validateIDs(ids []string) error {
	for _, id := range ids {
		if !validStableID(id) {
			return fmt.Errorf("invalid stable ID %q", id)
		}
	}
	return nil
}

func validStableID(value string) bool {
	parsed, err := semantic.ParseIdentity(value)
	return err == nil && parsed.String() == value
}

func validDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func validRepoPath(value string) bool {
	return value != "" && value != "." && !strings.HasPrefix(value, "/") && !strings.Contains(value, "\\") && path.Clean(value) == value && !strings.HasPrefix(value, "../") && value != ".." && !strings.ContainsRune(value, '\x00')
}

func validRepoDir(value string) bool {
	return value == "." || validRepoPath(value)
}
