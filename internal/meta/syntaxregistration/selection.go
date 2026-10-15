package syntaxregistration

import (
	"fmt"
	"go/ast"
)

func generateSelection(source *goSource, version, capability int) error {
	for _, item := range []struct{ function, result, kind, empty string }{
		{"activeDenominator", evidenceName(version), "[]byte", "nil"},
		{"activeDenominatorDigest", digestName(version), "string", `""`},
	} {
		if err := extendSelection(source, item.function, item.result, item.kind, item.empty, capability); err != nil {
			return err
		}
	}
	var anchor ast.Node
	for _, declaration := range source.file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if ok && len(value.Names) == 1 && value.Names[0].Name == evidenceName(version-1) {
				if anchor != nil {
					return fmt.Errorf("baseline embedded evidence is ambiguous")
				}
				anchor = value
			}
		}
	}
	if anchor == nil {
		return fmt.Errorf("baseline embedded evidence is missing")
	}
	if err := source.requireImport(anchor, "embed", "_"); err != nil {
		return err
	}
	source.appendAt(anchor, fmt.Sprintf("\n//go:embed evidence/denominator-v%d.json\nvar %s []byte\n", version, evidenceName(version)))
	return nil
}

func extendSelection(source *goSource, name, result, kind, empty string, capability int) error {
	root, err := source.function(name)
	if err != nil {
		return err
	}
	entries, helpers, err := selectionEntries(source, root, kind, empty)
	if err != nil {
		return err
	}
	baseline := 0
	seen := map[int]bool{}
	for _, entry := range entries {
		if seen[entry.capability] {
			return fmt.Errorf("denominator selection %s repeats capability", name)
		}
		seen[entry.capability] = true
		if entry.capability == capability+1 {
			return fmt.Errorf("denominator selection %s already contains new capability", name)
		}
		if entry.capability == capability {
			baseline++
		}
	}
	if baseline != 1 {
		return fmt.Errorf("denominator selection %s lacks one exact baseline", name)
	}
	entries = append(entries, selectionEntry{capability + 1, result})
	wrapper, chunks := renderSelection(name, kind, empty, entries)
	source.replace(root, wrapper)
	for _, helper := range helpers {
		source.replace(helper, "")
	}
	source.appendAt(root, chunks)
	return nil
}
