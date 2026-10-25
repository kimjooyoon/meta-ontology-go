package languagesyntax

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func projectInventorySources(before []replay.FileObservation, directory string,
	proposal sourcebundle.Partition) ([]replay.FileObservation, error) {
	positions := make(map[string]int, len(before))
	created := path.Join(directory, proposal.Files[1].Filename)
	for i, file := range before {
		if strings.EqualFold(file.Path, created) {
			return nil, fmt.Errorf("partition destination already appears in source inventory: %s", file.Path)
		}
		positions[file.Path] = i
	}
	for _, member := range proposal.Before.Provenance.Members {
		name := path.Join(directory, member.Filename)
		index, found := positions[name]
		if !found || before[index].SourceDigest != "sha256:"+member.SHA256 {
			return nil, fmt.Errorf("partition member differs from source inventory: %s", name)
		}
	}
	after := slices.Clone(before)
	for _, member := range proposal.Sources {
		name := path.Join(directory, member.Filename)
		observed := replay.ObserveSource(name, []byte(member.Content))
		if i, exists := positions[name]; exists {
			after[i] = observed
		} else {
			after = append(after, observed)
		}
	}
	slices.SortFunc(after, func(a, b replay.FileObservation) int { return compareWalkPaths(a.Path, b.Path) })
	return after, nil
}

// WalkDir sorts directory entries, then visits each subtree. Full-path sorting
// differs for e.g. a/nested.gooo and a.gooo; preserve the existing observer order.
func compareWalkPaths(a, b string) int {
	for {
		aHead, aTail, aMore := strings.Cut(a, "/")
		bHead, bTail, bMore := strings.Cut(b, "/")
		if comparison := strings.Compare(aHead, bHead); comparison != 0 {
			return comparison
		}
		if !aMore && !bMore {
			return 0
		}
		if !aMore {
			return -1
		}
		if !bMore {
			return 1
		}
		a, b = aTail, bTail
	}
}
