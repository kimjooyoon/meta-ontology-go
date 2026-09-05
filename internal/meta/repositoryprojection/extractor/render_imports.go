package extractor

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"

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
	normalized, eligible, err := generation.NormalizeImportHeaderGroup(file, group, specs, generation.ImportHeaderNormalizationPlain)
	if err != nil {
		return nil, fail("validate-ast-imports", "normalize-imports", "IMPORT_NORMALIZATION_POLICY_UNPROVEN", "DIRECT_MISSING", "restore-operation-input-contract", nil)
	}
	if !eligible {
		return formatImport(fset, file, group, specs)
	}
	var output bytes.Buffer
	for index, declaration := range normalized {
		if index > 0 {
			output.WriteByte('\n')
		}
		data, err := formatImport(fset, file, declaration, declaration.Specs)
		if err != nil {
			return nil, err
		}
		output.Write(data)
	}
	return output.Bytes(), nil
}

func eligiblePlainImportGroup(file *ast.File, group *ast.GenDecl, specs []*ast.ImportSpec) bool {
	_, eligible, err := generation.NormalizeImportHeaderGroup(file, group, specs, generation.ImportHeaderNormalizationPlain)
	return err == nil && eligible
}

func formatImport(fset *token.FileSet, file *ast.File, group *ast.GenDecl, specs []*ast.ImportSpec) ([]byte, error) {
	copyGroup := *group
	copyGroup.Specs = make([]ast.Spec, len(specs))
	for i, spec := range specs {
		copyGroup.Specs[i] = spec
	}
	if len(specs) == 1 {
		copyGroup.Lparen, copyGroup.Rparen = token.NoPos, token.NoPos
	}
	var out bytes.Buffer
	var node any = &copyGroup
	if comments := importComments(file, group); len(comments) != 0 {
		node = &printer.CommentedNode{Node: &copyGroup, Comments: comments}
	}
	if err := format.Node(&out, fset, node); err != nil {
		return nil, fail("rewrite-source", "render-imports", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return out.Bytes(), nil
}

func importComments(file *ast.File, group *ast.GenDecl) []*ast.CommentGroup {
	if file == nil || group == nil {
		return nil
	}
	start := group.Pos()
	if group.Doc != nil && group.Doc.Pos() < start {
		start = group.Doc.Pos()
	}
	for _, raw := range group.Specs {
		spec, ok := raw.(*ast.ImportSpec)
		if ok && spec.Doc != nil && spec.Doc.Pos() < start {
			start = spec.Doc.Pos()
		}
	}
	comments := make([]*ast.CommentGroup, 0)
	for _, current := range file.Comments {
		if current.End() >= start && current.Pos() <= group.End() {
			comments = append(comments, current)
		}
	}
	return comments
}
