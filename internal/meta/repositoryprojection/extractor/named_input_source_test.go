package extractor

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestNamedInputSourcePartitionsFirst(t *testing.T) {
	for _, path := range strings.Fields(namedInputProjectionSubjects) {
		t.Run(strings.ReplaceAll(path, "/", "_"), func(t *testing.T) {
			result, err := ExtractWithResult("../../../..", path)
			if err != nil {
				t.Fatal(err)
			}
			for name, raw := range result.Generated {
				if physicalLines(raw) > 75 {
					t.Fatal("partition over capacity", name)
				}
				if _, err := parser.ParseFile(token.NewFileSet(), name, raw, parser.AllErrors); err != nil {
					t.Fatal(name, err)
				}
			}
		})
	}
}

const namedInputProjectionSubjects = namedInputSubjects0 + "\n" + namedInputSubjects1 + "\n" + namedInputSubjects2

const namedInputSubjects0 = `internal/bidir/bx_contract_hash_part01.go
internal/bidir/equivalence_part01.go
internal/bidir/input_names.go
internal/bidir/input_names_test.go
internal/bidir/lens_helpers_part01.go
internal/bidir/lens_helpers_part03.go
internal/bidir/lens_part02.go
internal/bidir/lower_document_part03.go
internal/bidir/lower_part05.go
internal/bidir/model_part01.go
internal/bidir/model_part02.go
internal/bidir/model_part04.go
internal/bodycodegen/body_condition_ranges.go
internal/bodycodegen/body_source.go
internal/bodycodegen/body_value_origin.go
internal/bodycodegen/body_value_transfer.go
internal/bodycodegen/input_names_test.go
internal/bodycodegen/input_parameters.go
internal/bodycodegen/rename_assembly.go
internal/bodycodegen/rename_body.go
internal/bodycodegen/rename_body_parse.go
internal/bodycodegen/source_activity_reference.go
internal/bodycodegen/source_fill_origin.go
internal/bodycodegen/source_references.go
internal/bodycodegen/source_rename.go`

const namedInputSubjects1 = `internal/bodycodegen/typed_input_declaration.go
internal/bodycodegen/typed_input_declaration_text.go
internal/bodycodegen/typed_value_origins.go
internal/inputbinding/names.go
internal/meta/languagereadiness/languagesyntax/replay/shape.go
internal/packageruntime/declarations.go
internal/packageruntime/interface.go
internal/packageruntime/interface_projection.go
internal/packageruntime/model.go
internal/packageruntime/workspaceexecution/activity_rename_references.go
internal/packageruntime/workspaceexecution/body_call_rewrite.go
internal/packageruntime/workspaceexecution/condition_source.go
internal/packageruntime/workspaceexecution/input_names_record_test.go
internal/packageruntime/workspaceexecution/input_names_test.go
internal/packageruntime/workspaceexecution/record_constructor_scope.go
internal/packageruntime/workspaceexecution/value_transfer_catalog.go
internal/packageruntime/workspaceexecution/value_transfer_query.go
internal/packageruntime/workspaceexecution/value_transfer_walk.go
internal/semantic/canonical_part01.go
internal/semantic/graph_part03.go
internal/semantic/graph_validate_part01.go
internal/semantic/input_names.go
internal/semantic/types_part02.go
internal/semantic/types_part03.go
internal/sourceimpact/compare.go`

const namedInputSubjects2 = `internal/sourceimpact/input_names_test.go
internal/sourceimpact/inspect.go
internal/sourceimpact/references.go
internal/syntax/activity_input_names.go
internal/syntax/activity_input_names_test.go
internal/syntax/activity_parser.go
internal/syntax/activity_type_reference.go
internal/syntax/ast_part03.go
internal/syntax/ast_part04.go
internal/syntax/format_part04.go`
