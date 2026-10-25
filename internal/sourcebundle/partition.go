package sourcebundle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcepartition"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type Partition struct {
	Before, After       Bundle
	BoundaryID          string
	Sources             []Source
	Files               []sourcepartition.File
	EntityFieldsSupport *syntax.EntityFieldsSupport `json:"entity_fields_support,omitempty"`
}

// PreparePartition resolves the boundary in the original source set by stable ID.
// Both outputs use the same lossless partitioner as external workspace proposals.
// Consumer selection, native ABI acceptance and publication are separate obligations.
func PreparePartition(sources []Source, filename, beforeID, into string) (Partition, error) {
	return preparePartition(sources, filename, beforeID, into, syntax.CurrentEntityFieldsSupport())
}

// PreparePartitionWithEntityFieldsSupport records the explicit profile used by
// this partition. Registry consumers must still replay under their own profiles.
func PreparePartitionWithEntityFieldsSupport(sources []Source, filename, beforeID, into string,
	support syntax.EntityFieldsSupport) (Partition, error) {
	result, err := preparePartition(sources, filename, beforeID, into, support)
	if err != nil {
		return Partition{}, err
	}
	result.EntityFieldsSupport = &support
	return result, nil
}

func preparePartition(sources []Source, filename, beforeID, into string, support syntax.EntityFieldsSupport) (Partition, error) {
	before, err := LowerWithEntityFieldsSupport(sources, support)
	if err != nil {
		return Partition{}, err
	}
	index, cut := -1, -1
	for i, source := range sources {
		if strings.EqualFold(source.Filename, into) {
			return Partition{}, fmt.Errorf("partition destination already belongs to the source set")
		}
		if source.Filename == filename {
			index = i
		}
	}
	for _, origin := range before.Provenance.Declarations {
		if origin.ID == beforeID && origin.Span.File == filename {
			cut = origin.Span.Start
		}
	}
	if index < 0 || cut < 0 {
		return Partition{}, fmt.Errorf("partition requires a stable declaration ID from the selected source")
	}
	files, err := sourcepartition.SplitAtWithEntityFieldsSupport(filename, sources[index].Content, into, cut, support)
	if err != nil {
		return Partition{}, err
	}
	afterSources := slices.Clone(sources)
	afterSources[index].Content = files[0].Source
	afterSources = append(afterSources, Source{Filename: into, Content: files[1].Source})
	after, err := LowerWithEntityFieldsSupport(afterSources, support)
	if err != nil {
		return Partition{}, err
	}
	if before.IR.SemanticCanonical() != after.IR.SemanticCanonical() {
		return Partition{}, fmt.Errorf("partition changed the source set's meaning")
	}
	return Partition{Before: before, After: after, BoundaryID: beforeID, Sources: afterSources, Files: files}, nil
}
