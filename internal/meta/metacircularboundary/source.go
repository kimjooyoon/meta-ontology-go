package metacircularboundary

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func observeSource(path string, source []byte) (SourceObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.HasErrors() {
		return SourceObservation{}, fmt.Errorf("parse %s: %w", path, diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceObservation{}, fmt.Errorf("lower %s: %w", path, err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		return SourceObservation{}, fmt.Errorf("normalize %s: %w", path, err)
	}
	if err := normalized.Validate(); err != nil {
		return SourceObservation{}, fmt.Errorf("validate %s: %w", path, err)
	}
	entities := []string{}
	activities := []string{}
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			entities = append(entities, value.Name)
		case *syntax.ActivityDecl:
			activities = append(activities, value.Name)
		}
	}
	if !containsAll(entities, requiredEntities) || !containsAll(activities, requiredActivities) {
		return SourceObservation{}, fmt.Errorf("meta-circular source does not declare the boundary vocabulary")
	}
	return SourceObservation{
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
