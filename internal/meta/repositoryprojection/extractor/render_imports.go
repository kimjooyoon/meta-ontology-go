package extractor

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"sort"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func renderImports(fset *token.FileSet, file *ast.File, decls []ast.Decl, list []importSpec, includeBlank bool) (map[*ast.GenDecl][]byte, []byte, error) {
	if err := importNormalizationPolicy(); err != nil {
		return nil, nil, err
	}
	selected, err := selectedImports(decls, list, includeBlank)
	if err != nil {
		return nil, nil, err
	}
	replacements := map[*ast.GenDecl][]byte{}
	var helper bytes.Buffer
	for _, node := range file.Decls {
		group, ok := node.(*ast.GenDecl)
		if !ok || group.Tok != token.IMPORT || len(selected[group]) == 0 {
			continue
		}
		data, err := formatSelectedImports(fset, file, group, selected[group])
		if err != nil {
			return nil, nil, err
		}
		replacements[group] = data
		helper.Write(data)
		helper.WriteByte('\n')
	}
	return replacements, helper.Bytes(), nil
}

func importNormalizationPolicy() error {
	policy, err := generation.ImportNormalizationPolicyEvidence()
	if err != nil || policy.InputSubjectKind != sourcepolicy.SubjectKindFile || policy.SourceDigest == "" || policy.SemanticDigest == "" || !policy.UsedInputFact || !policy.GeneratedOutputFact {
		return fail("validate-ast-imports", "normalize-imports", "IMPORT_NORMALIZATION_POLICY_UNPROVEN", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil || policy.SourceDigest != contract.SourceDigest || policy.SemanticDigest != contract.SemanticDigest {
		return failWithDiagnostics("validate-ast-imports", "normalize-imports", "IMPORT_NORMALIZATION_POLICY_UNPROVEN", "KNOWN_CONTRADICTION", "report-counterexample", []string{"policy=eligible-plain-import-group"})
	}
	return nil
}

func formatSelectedImports(fset *token.FileSet, file *ast.File, group *ast.GenDecl, specs []*ast.ImportSpec) ([]byte, error) {
	if !eligiblePlainImportGroup(file, group, specs) {
		return formatImport(fset, group, specs)
	}
	ordered := append([]*ast.ImportSpec{}, specs...)
	sort.SliceStable(ordered, func(left, right int) bool { return ordered[left].Pos() < ordered[right].Pos() })
	var output bytes.Buffer
	for index, spec := range ordered {
		if index > 0 {
			output.WriteByte('\n')
		}
		data, err := formatImport(fset, group, []*ast.ImportSpec{spec})
		if err != nil {
			return nil, err
		}
		output.Write(data)
	}
	return output.Bytes(), nil
}

func eligiblePlainImportGroup(file *ast.File, group *ast.GenDecl, specs []*ast.ImportSpec) bool {
	if file == nil || group == nil || group.Doc != nil || len(group.Specs) < 2 || len(specs) < 2 {
		return false
	}
	for _, comments := range file.Comments {
		if comments.Pos() >= group.Pos() && comments.Pos() <= group.End() {
			return false
		}
	}
	for _, raw := range group.Specs {
		spec, ok := raw.(*ast.ImportSpec)
		if !ok || spec.Name != nil || spec.Doc != nil || spec.Comment != nil || spec.Path == nil {
			return false
		}
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path == "C" {
			return false
		}
	}
	return true
}

func formatImport(fset *token.FileSet, group *ast.GenDecl, specs []*ast.ImportSpec) ([]byte, error) {
	copyGroup := *group
	copyGroup.Specs = make([]ast.Spec, len(specs))
	for i, spec := range specs {
		copyGroup.Specs[i] = spec
	}
	if len(specs) == 1 {
		copyGroup.Lparen, copyGroup.Rparen = token.NoPos, token.NoPos
	}
	var out bytes.Buffer
	if err := format.Node(&out, fset, &copyGroup); err != nil {
		return nil, fail("rewrite-source", "render-imports", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return out.Bytes(), nil
}
