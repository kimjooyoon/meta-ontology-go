package publicpartialreuse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

const PartitionInputProfile = "gooo/independent-declaration-partitions/v1"

// The generation record binds the complete source and output. Its raw identity
// belongs to the original invocation; only the checked partition projection is
// compared across source revisions.
func validateGenerationOrigin(policy Policy, program, manifest []byte) error {
	if err := completeness.CheckJSON(manifest, 1<<20); err != nil {
		return err
	}
	var value struct {
		Schema    string `json:"schema"`
		Producer  string `json:"producer"`
		Operation string `json:"operation"`
		Status    string `json:"status"`
		Source    string `json:"source_digest"`
		Semantic  string `json:"semantic_digest"`
		Generated string `json:"generated_digest"`
	}
	if err := json.Unmarshal(manifest, &value); err != nil {
		return err
	}
	if value.Schema != "gooo/cli/generate-manifest/v1" || value.Producer != "gooo" || value.Operation != "generate" ||
		value.Status != "pass" || value.Source != policy.SourceDigest || value.Semantic != policy.SemanticDigest ||
		value.Generated != cache.HashBytes(program).String() {
		return errors.New("generation origin does not bind the selected source and complete program")
	}
	return nil
}

func partitionOwners(policy Policy) (map[string]string, error) {
	owners := map[string]string{}
	for _, partition := range policy.Partitions {
		for _, symbol := range partition.Symbols {
			if owners[symbol] != "" {
				return nil, fmt.Errorf("generated symbol %q has ambiguous partition ownership", symbol)
			}
			owners[symbol] = partition.ID
		}
	}
	return owners, nil
}

// This profile conservatively rejects cross-partition references, including
// references from shared declarations into a partition. Name observations may
// overapproximate references; a rejected profile never authorizes reuse.
func checkPartitionReferences(declaration ast.Decl, owner string, owners map[string]string) error {
	var invalid string
	ast.Inspect(declaration, func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok {
			if target := owners[id.Name]; target != "" && target != owner {
				invalid = id.Name
				return false
			}
		}
		return invalid == ""
	})
	if invalid != "" {
		return fmt.Errorf("generated declaration depends on another partition through %q", invalid)
	}
	return nil
}

// Shared context keeps package identity, imports, unowned declarations and all
// comment text (including compiler directives) in source order. Byte offsets
// are not identities, so editing an earlier owned declaration does not move it.
func sharedProgramDigest(policy Policy, program []byte) (string, error) {
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "generated.go", program, parser.ParseComments)
	if err != nil {
		return "", err
	}
	owners, err := partitionOwners(policy)
	if err != nil {
		return "", err
	}
	shared := struct {
		Profile, Package string
		Declarations     []string
		Comments         []string
	}{
		Profile: PartitionInputProfile, Package: file.Name.Name, Declarations: []string{}, Comments: []string{},
	}
	for _, declaration := range file.Decls {
		owner := owners[declarationName(declaration)]
		if err := checkPartitionReferences(declaration, owner, owners); err != nil {
			return "", err
		}
		if owner == "" {
			var output bytes.Buffer
			if err := format.Node(&output, set, declaration); err != nil {
				return "", err
			}
			shared.Declarations = append(shared.Declarations, output.String())
		}
	}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			shared.Comments = append(shared.Comments, comment.Text)
		}
	}
	return hashJSON(shared), nil
}
