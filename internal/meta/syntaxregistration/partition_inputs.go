package syntaxregistration

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func readPartitionNativeInputs(repository fs.FS) (map[string][]byte, error) {
	if repository == nil {
		return nil, fmt.Errorf("partition registry requires a filesystem")
	}
	names, err := fs.Glob(repository, syntaxRoot+"*.go")
	if err != nil {
		return nil, err
	}
	selected := []string{corpusPath}
	for _, name := range names {
		if !strings.HasSuffix(name, "_test.go") {
			selected = append(selected, name)
		}
	}
	return readPartitionNamedInputs(repository, selected)
}

func readPartitionNamedInputs(repository fs.FS, names []string) (map[string][]byte, error) {
	if repository == nil || len(names) == 0 || len(names) > 1024 {
		return nil, fmt.Errorf("partition input scope is missing or exceeds 1024 files")
	}
	inputs, seen, remaining := map[string][]byte{}, map[string]bool{}, sourcebundle.MaxTotalBytes
	for _, name := range names {
		key := strings.ToLower(name)
		if !fs.ValidPath(name) || strings.ContainsAny(name, "\\\x00:") || seen[key] {
			return nil, fmt.Errorf("partition input address is invalid or aliased")
		}
		seen[key] = true
		file, err := repository.Open(name)
		if err != nil {
			return nil, err
		}
		limit := min(remaining, sourcebundle.MaxSourceBytes)
		raw, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
		closeErr := file.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(raw) > limit {
			return nil, fmt.Errorf("partition inputs exceed the source byte limits")
		}
		inputs[name], remaining = raw, remaining-len(raw)
	}
	return inputs, nil
}

func capturePartitionSources(repository fs.FS, directory string, partition sourcebundle.Partition,
	registry languagesyntax.Registry, projection languagesyntax.PartitionRegistration) (map[string][]byte, error) {
	names := map[string]bool{}
	for _, member := range partition.Before.Provenance.Members {
		names[path.Join(directory, member.Filename)] = true
	}
	for _, consumer := range projection.Consumers {
		for _, member := range consumer.Before.SourceMembers {
			names[member.Path] = true
		}
		if consumer.Kind == "case" {
			for _, definition := range registry.Cases {
				if definition.ID == consumer.ID {
					names[definition.Path] = true
				}
			}
		} else if consumer.Kind == "package-syntax" {
			for _, unit := range registry.PackageUnits {
				if unit.ID == consumer.ID {
					for _, member := range unit.Members {
						names[member] = true
					}
				}
			}
		}
	}
	inputs, err := readPartitionNamedInputs(repository, sortedPaths(names))
	if err != nil {
		return nil, err
	}
	observed := map[string]replay.FileObservation{}
	for _, item := range projection.Inventory.Before.Sources {
		observed[item.Path] = item
	}
	for name, raw := range inputs {
		if replay.ObserveSource(name, raw) != observed[name] {
			return nil, fmt.Errorf("source consumer changed after partition observation: %s", name)
		}
	}
	return inputs, nil
}
