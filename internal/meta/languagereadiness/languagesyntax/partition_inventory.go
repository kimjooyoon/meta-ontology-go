package languagesyntax

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

// PartitionInventory compares the current source inventory with a proposed
// partition under the same registry. Projected is not installed. Existing
// missing/unregistered paths remain visible; neither view is semantic admission.
type PartitionInventory struct {
	Before, Projected SourceInventory
	AddedUnregistered []string
}

// InspectPartitionInventory replays a prepared partition against current member
// bytes, then applies its file observations to the existing corpus comparison.
// It writes nothing, does not edit registration, and discovers no other consumer.
// The caller owns global snapshot freshness; selected members are rechecked.
func InspectPartitionInventory(repository fs.FS, registryRaw []byte, directory string,
	proposal sourcebundle.Partition) (PartitionInventory, error) {
	registry, err := decodeRegistry(registryRaw)
	if err != nil {
		return PartitionInventory{}, err
	}
	if repository == nil || !fs.ValidPath(directory) || strings.ContainsAny(directory, "\\\x00:") {
		return PartitionInventory{}, fmt.Errorf("partition inventory requires a filesystem and canonical relative directory")
	}
	if err := replayInventoryPartition(repository, directory, proposal); err != nil {
		return PartitionInventory{}, err
	}
	before, err := replay.Observe(repository)
	if err != nil {
		return PartitionInventory{}, err
	}
	after, err := projectInventorySources(before, directory, proposal)
	if err != nil {
		return PartitionInventory{}, err
	}
	// Recheck the selected member bytes and occupied destination after the walk.
	if err := replayInventoryPartition(repository, directory, proposal); err != nil {
		return PartitionInventory{}, err
	}
	result := PartitionInventory{Before: inventoryFromObservations(registryRaw, registry, before),
		Projected: inventoryFromObservations(registryRaw, registry, after)}
	for _, name := range result.Projected.Unregistered {
		if !slices.Contains(result.Before.Unregistered, name) {
			result.AddedUnregistered = append(result.AddedUnregistered, name)
		}
	}
	return result, nil
}

func replayInventoryPartition(repository fs.FS, directory string, proposal sourcebundle.Partition) error {
	if len(proposal.Files) != 2 {
		return fmt.Errorf("partition inventory requires one replayable two-file partition")
	}
	files, err := fs.Sub(repository, directory)
	if err != nil {
		return err
	}
	names := make([]string, len(proposal.Before.Provenance.Members))
	for i, member := range proposal.Before.Provenance.Members {
		names[i] = member.Filename
	}
	sources, err := sourcebundle.Read(files, names)
	if err != nil {
		return err
	}
	if err := sourcebundle.ValidatePartition(sources, proposal); err != nil {
		return err
	}
	if _, err := fs.Stat(repository, path.Join(directory, proposal.Files[1].Filename)); !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("partition inventory destination must be absent")
	}
	return nil
}
