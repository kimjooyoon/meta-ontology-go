package embeduse

import (
	"crypto/sha256"
	"fmt"
	"os"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func selectedSources(inputs map[string]string, names []string) ([]sourcebundle.Source, error) {
	if len(names) == 0 || len(names) > sourcebundle.MaxSources {
		return nil, fmt.Errorf("partition scope requires 1 to %d observed source members", sourcebundle.MaxSources)
	}
	result := make([]sourcebundle.Source, 0, len(names))
	for _, name := range names {
		content, ok := inputs[name]
		if !ok {
			return nil, fmt.Errorf("partition source %q has no current consumer input observation", name)
		}
		result = append(result, sourcebundle.Source{Filename: name, Content: content})
	}
	return result, nil
}

func captureInputs(report Report, projections []sourcebundle.Source) (map[string]string, error) {
	o := observer{root: report.Directory}
	proposed := make(map[string]string, len(projections))
	for _, source := range projections {
		proposed[source.Filename] = source.Content
	}
	result := make(map[string]string, len(report.Files))
	for _, file := range report.Files {
		content, projected := proposed[file.Filename]
		if file.Projected != projected {
			return nil, fmt.Errorf("partition projection identity differs from the observation")
		}
		if !projected {
			address, err := o.address(file.Filename)
			if err != nil {
				return nil, err
			}
			raw, err := readBounded(address)
			if err != nil {
				return nil, err
			}
			content = string(raw)
		}
		if len(content) != file.ByteLength || fmt.Sprintf("%x", sha256.Sum256([]byte(content))) != file.SHA256 {
			return nil, fmt.Errorf("partition input changed since observation: %s", file.Filename)
		}
		result[file.Filename] = content
	}
	return result, nil
}

func unusedDestination(root, name string) error {
	o := observer{root: root}
	address, err := o.address(name)
	if err != nil {
		return err
	}
	_, err = os.Lstat(address)
	if !os.IsNotExist(err) {
		return fmt.Errorf("partition destination must not already exist: %s", name)
	}
	return nil
}

func samePartitionContext(before, after Report) error {
	if before.Directory != after.Directory || before.ImportPath != after.ImportPath ||
		!reflect.DeepEqual(before.Build, after.Build) || !reflect.DeepEqual(before.IgnoredGoFiles, after.IgnoredGoFiles) ||
		len(before.Consumers) != len(after.Consumers) {
		return fmt.Errorf("embed build context changed during partition planning")
	}
	prior := make(map[string]sourcebundle.Member, len(before.Files))
	for _, file := range before.Files {
		prior[file.Filename] = file.Member
	}
	for _, file := range after.Files {
		if !file.Projected && prior[file.Filename] != file.Member {
			return fmt.Errorf("unprojected consumer input changed: %s", file.Filename)
		}
		delete(prior, file.Filename)
	}
	if len(prior) != 0 {
		return fmt.Errorf("original consumer input was omitted during partition planning")
	}
	for i, original := range before.Consumers {
		current := after.Consumers[i]
		if original.Name != current.Name || original.Kind != current.Kind || original.DeclaredType != current.DeclaredType ||
			original.Declaration != current.Declaration || before.Files[original.GoSource] != after.Files[current.GoSource] {
			return fmt.Errorf("embed consumer declaration changed during partition planning")
		}
	}
	return nil
}
