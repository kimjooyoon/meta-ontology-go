package main

import (
	"errors"
	"sort"
	"strings"

	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func plannerManifestFromAnalyzerSnapshot(snapshot analyzersci.Snapshot) (plannersci.SnapshotManifest, error) {
	files := make([]plannersci.SnapshotFile, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		ids := make([]string, 0, len(source.Bindings))
		seen := map[string]struct{}{}
		for _, binding := range source.Bindings {
			if binding.Status != analyzersci.StatusBound || binding.ID == "" {
				return plannersci.SnapshotManifest{}, errors.New("source binding is not BOUND")
			}
			if _, exists := seen[binding.ID]; exists {
				return plannersci.SnapshotManifest{}, errors.New("duplicate source binding")
			}
			seen[binding.ID] = struct{}{}
			ids = append(ids, binding.ID)
		}
		sort.Strings(ids)
		files = append(files, plannersci.SnapshotFile{Path: source.Path, BlobDigest: rawDigest(source.BlobDigest), SemanticIDs: ids})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	manifest := plannersci.SnapshotManifest{SchemaVersion: plannersci.ManifestSchemaVersion, Files: files}
	manifest.Digest = manifest.ComputedDigest()
	if err := manifest.Validate(); err != nil {
		return plannersci.SnapshotManifest{}, err
	}
	return manifest, nil
}

// rawDigest bridges the analyzer's labelled SHA-256 spelling to the raw-hex
// spelling used by the planner, proof, and lane contracts. It does not hash or
// otherwise alter the digest identity.
func rawDigest(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}

func selectedShadowCommands(plan plannersci.PlanResult, registry plannersci.Registry) ([]shadowCommandSpec, []shadowCommandSpec, []shadowResourceReceipt, error) {
	commands := make(map[string]plannersci.Command, len(registry.Commands))
	for _, command := range registry.Commands {
		commands[command.ID] = command
	}
	guards := make(map[string]plannersci.Command, len(registry.GlobalGuardCommands))
	for _, command := range registry.GlobalGuardCommands {
		guards[command.ID] = command
	}
	makeSpecs := func(ids []string, source map[string]plannersci.Command) ([]shadowCommandSpec, []shadowResourceReceipt, error) {
		specs := make([]shadowCommandSpec, 0, len(ids))
		receipts := make([]shadowResourceReceipt, 0, len(ids))
		for _, id := range ids {
			command, ok := source[id]
			if !ok {
				return nil, nil, errors.New("selected command is not registered")
			}
			specs = append(specs, shadowCommandSpec{ID: command.ID, Argv: append([]string{}, command.Argv...)})
			receipts = append(receipts, shadowResourceReceipt{CommandID: command.ID, CPUWorkUnits: command.CPUWorkUnits, MemoryBytes: command.MemoryBytes})
		}
		sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })
		sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
		return specs, receipts, nil
	}
	commandSpecs, commandReceipts, err := makeSpecs(plan.SelectedCommandIDs, commands)
	if err != nil {
		return nil, nil, nil, err
	}
	guardSpecs, guardReceipts, err := makeSpecs(plan.SelectedGuardCommandIDs, guards)
	if err != nil {
		return nil, nil, nil, err
	}
	receipts := append(commandReceipts, guardReceipts...)
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
	return commandSpecs, guardSpecs, receipts, nil
}

func sortedUnion(left, right []string) []string {
	values := append(append([]string{}, left...), right...)
	sort.Strings(values)
	return uniqueStrings(values)
}

func sortedSemanticIDs(values []semantic.ID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	sort.Strings(result)
	return result
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}
