package languagesyntax

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

// PartitionRegistration projects existing registry membership jointly with an
// exact source partition. It preserves case definitions and their denominators.
// Native registry edits, external consumers and publication are later obligations.
// Inventory remains UNASSESSED: syntax/BX replay is not native program execution.
type PartitionRegistration struct {
	Registry  Registry
	Corpus    []byte
	Inventory PartitionInventory
	Consumers []PartitionConsumerReplay
}

type PartitionConsumerReplay struct {
	Kind, ID string
	Before   replay.Result
	After    replay.Result
}

// PreparePartitionRegistration observes current bytes and returns one complete
// registry projection or none. A physical fragment never becomes a new case.
func PreparePartitionRegistration(repository fs.FS, registryRaw []byte, directory string,
	partition sourcebundle.Partition) (PartitionRegistration, error) {
	inventory, err := InspectPartitionInventory(repository, registryRaw, directory, partition)
	if err != nil {
		return PartitionRegistration{}, err
	}
	registry, err := decodeRegistry(registryRaw)
	if err != nil {
		return PartitionRegistration{}, err
	}
	source, into := path.Join(directory, partition.Files[0].Filename), path.Join(directory, partition.Files[1].Filename)
	if err := partitionRegistrationAddresses(registry, source, into); err != nil {
		return PartitionRegistration{}, err
	}
	projection := map[string][]byte{source: []byte(partition.Files[0].Source), into: []byte(partition.Files[1].Source)}
	planner := partitionRegistrationPlanner{repository: repository, inventory: inventory.Before.Sources,
		projection: projection, inputs: map[string][]byte{}}
	consumers, err := planner.projectMembership(&registry, source, into)
	if err != nil {
		return PartitionRegistration{}, err
	}
	if err := validateCaseSources(registry); err != nil {
		return PartitionRegistration{}, err
	}
	if err := planner.recheckInputs(); err != nil {
		return PartitionRegistration{}, err
	}
	if err := replayInventoryPartition(repository, directory, partition); err != nil {
		return PartitionRegistration{}, err
	}
	projectedRaw, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return PartitionRegistration{}, err
	}
	projectedRaw = append(projectedRaw, '\n')
	inventory.Projected = inventoryFromObservations(projectedRaw, registry, inventory.Projected.Sources)
	inventory.AddedUnregistered = nil
	for _, name := range inventory.Projected.Unregistered {
		if !slices.Contains(inventory.Before.Unregistered, name) {
			inventory.AddedUnregistered = append(inventory.AddedUnregistered, name)
		}
	}
	return PartitionRegistration{Registry: registry, Corpus: projectedRaw, Inventory: inventory, Consumers: consumers}, nil
}

func partitionRegistrationAddresses(registry Registry, source, into string) error {
	found := false
	for _, name := range registryPaths(registry) {
		if strings.EqualFold(name, into) {
			return fmt.Errorf("partition destination already has a registry role: %s", name)
		}
		if strings.EqualFold(name, source) {
			if name != source {
				return fmt.Errorf("partition source differs from its registered address: %s", name)
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("partition requires an already registered source")
	}
	for _, unit := range registry.ProjectionUnits {
		if unit.Path == source {
			return fmt.Errorf("projection consumer %s requires a source-family adapter before partitioning", unit.ID)
		}
	}
	return nil
}
