package packageruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Qualified spellings are source references. Lowering reuses an existing entity
// with the resolved ID, or gives that ID a local internal name without its alias.
func appendQualifiedEntityTypes(file *syntax.File, declarations []syntax.Declaration,
	types map[string]string) []syntax.Declaration {
	canonical, used := map[string]string{}, map[string]bool{}
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			used[value.Name] = true
			if previous := canonical[value.ID]; previous == "" || value.Name < previous {
				canonical[value.ID] = value.Name
			}
		case *syntax.ActivityDecl:
			used[value.Name] = true
		case *syntax.PolicyDecl:
			used[value.Name] = true
		}
	}
	var names []string
	for name := range types {
		if strings.Contains(name, ".") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	aliases := map[string]string{}
	for _, name := range names {
		id := types[name]
		lowered := canonical[id]
		if lowered == "" {
			digest := sha256.Sum256([]byte(id))
			lowered = "GoooImportedType" + hex.EncodeToString(digest[:])
			for used[lowered] {
				lowered += "X"
			}
			declarations = append(declarations, &syntax.EntityDecl{Name: lowered, ID: id})
			canonical[id], used[lowered] = lowered, true
		}
		aliases[name] = lowered
	}
	for _, declaration := range declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		for i := range activity.Inputs {
			if lowered := aliases[activity.Inputs[i].Name]; lowered != "" {
				activity.Inputs[i].Name = lowered
			}
		}
		for i := range activity.Parameters {
			if lowered := aliases[activity.Parameters[i].Name]; lowered != "" {
				activity.Parameters[i].Name = lowered
			}
		}
		if lowered := aliases[activity.Output]; lowered != "" {
			activity.Output = lowered
		}
		if lowered := aliases[activity.Result.Name]; lowered != "" {
			activity.Result.Name = lowered
		}
	}
	return declarations
}

// Retain the existing package-wide bare-name environment while qualified aliases
// remain local to the file that declares their import.
func sourceTypeEnvironment(bare, local map[string]string) map[string]string {
	result := make(map[string]string, len(bare)+len(local))
	maps.Copy(result, bare)
	maps.Copy(result, local)
	return result
}
