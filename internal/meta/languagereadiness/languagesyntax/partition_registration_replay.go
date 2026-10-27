package languagesyntax

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

type partitionRegistrationPlanner struct {
	repository fs.FS
	inventory  []replay.FileObservation
	projection map[string][]byte
	inputs     map[string][]byte
}

func (planner *partitionRegistrationPlanner) replayConsumer(kind string, definition CaseDefinition,
	beforeNames, afterNames []string) (PartitionConsumerReplay, error) {
	sources, err := sourcebundle.Read(planner.repository, beforeNames)
	if err != nil {
		return PartitionConsumerReplay{}, err
	}
	files := make(map[string][]byte, len(sources))
	for _, source := range sources {
		raw := []byte(source.Content)
		if !observedCaseMember(planner.inventory, replay.ObserveSource(source.Filename, raw)) {
			return PartitionConsumerReplay{}, fmt.Errorf("consumer input changed after inventory: %s", source.Filename)
		}
		files[source.Filename], planner.inputs[source.Filename] = raw, raw
	}
	pinned := sourcebundle.Overlay{Files: files}
	before := beforeNames
	if len(beforeNames) == 1 {
		before = nil
	}
	result := PartitionConsumerReplay{Kind: kind, ID: definition.ID,
		Before: ReplayCaseSources(pinned, definition, before),
		After:  ReplayCaseSources(sourcebundle.Overlay{FS: pinned, Files: planner.projection}, definition, afterNames)}
	if result.Before.ObservedDecision != DecisionPass || result.After.ObservedDecision != DecisionPass ||
		result.Before.SemanticDigest != result.After.SemanticDigest {
		return PartitionConsumerReplay{}, fmt.Errorf("partition does not preserve %s %s under its declared profile", kind, definition.ID)
	}
	return result, nil
}

func (planner *partitionRegistrationPlanner) recheckInputs() error {
	for _, name := range slices.Sorted(maps.Keys(planner.inputs)) {
		raw := planner.inputs[name]
		sources, err := sourcebundle.Read(planner.repository, []string{name})
		if err != nil {
			return err
		}
		if !bytes.Equal(raw, []byte(sources[0].Content)) {
			return fmt.Errorf("consumer source changed during registration planning: %s", name)
		}
	}
	return nil
}
