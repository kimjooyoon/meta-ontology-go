package sourcebundle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcepartition"
)

type Partition struct {
	Before, After Bundle
	BoundaryID    string
	Sources       []Source
	Files         []sourcepartition.File
}

// PreparePartition resolves the boundary in the original source set by stable ID.
// Both outputs use the same lossless partitioner as external workspace proposals.
// Consumer selection, native ABI acceptance and publication are separate obligations.
func PreparePartition(sources []Source, filename, beforeID, into string) (Partition, error) {
	before, err := Lower(sources)
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
	files, err := sourcepartition.SplitAt(filename, sources[index].Content, into, cut)
	if err != nil {
		return Partition{}, err
	}
	afterSources := slices.Clone(sources)
	afterSources[index].Content = files[0].Source
	afterSources = append(afterSources, Source{Filename: into, Content: files[1].Source})
	after, err := Lower(afterSources)
	if err != nil {
		return Partition{}, err
	}
	if before.IR.SemanticCanonical() != after.IR.SemanticCanonical() {
		return Partition{}, fmt.Errorf("partition changed the source set's meaning")
	}
	return Partition{Before: before, After: after, BoundaryID: beforeID, Sources: afterSources, Files: files}, nil
}
