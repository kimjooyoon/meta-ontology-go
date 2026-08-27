package metacircularboundaryconsumer

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// observeSource is a consumer-owned source replay. It shares language
// parsing/lowering infrastructure only; it does not call the producer.
func observeSource(path string, source []byte) (contract.SourceObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.HasErrors() || file == nil || file.Package == nil || file.Namespace == nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not parse %s", path)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not lower %s: %w", path, err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not normalize %s: %w", path, err)
	}
	if err := normalized.Validate(); err != nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not validate %s: %w", path, err)
	}
	entities := []string{}
	activities := []string{}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			entities = append(entities, value.Name)
		case *syntax.ActivityDecl:
			activities = append(activities, value.Name)
		}
	}
	if !containsAll(entities, requiredEntities) || !containsAll(activities, requiredActivities) {
		return contract.SourceObservation{}, fmt.Errorf("consumer boundary vocabulary is incomplete")
	}
	return contract.SourceObservation{
		Path: path, SourceDigest: digestBytes(source), SemanticDigest: digestBytes([]byte(normalized.SemanticCanonical())),
		Package: file.Package.Name, Namespace: file.Namespace.Name, Entities: entities, Activities: activities,
		DescriptionBound: true, ReadOnly: true, RepositoryWrites: 0, MutationAuthority: false,
	}, nil
}

func containsAll(observed, required []string) bool {
	set := make(map[string]struct{}, len(observed))
	for _, value := range observed {
		set[value] = struct{}{}
	}
	for _, value := range required {
		if _, ok := set[value]; !ok {
			return false
		}
	}
	return true
}
