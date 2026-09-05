package generation

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type ImportHeaderNormalizationMode string

const (
	ImportHeaderNormalizationPlain      ImportHeaderNormalizationMode = "plain-source-rewrite"
	ImportHeaderNormalizationNamedAlias ImportHeaderNormalizationMode = "named-alias-header"
	ImportHeaderPlainCapability                                       = "plain-source-rewrite"
	ImportHeaderNamedAliasCapability                                  = "named-alias-single-spec-header"
)

// NormalizeImportHeaderGroup returns one-spec import declarations when the
// contract-authorized header mode can preserve the group's source semantics.
// It returns normalized=false for groups that must remain unchanged.
func NormalizeImportHeaderGroup(file *ast.File, group *ast.GenDecl, specs []*ast.ImportSpec, mode ImportHeaderNormalizationMode) ([]*ast.GenDecl, bool, error) {
	if mode != ImportHeaderNormalizationPlain && mode != ImportHeaderNormalizationNamedAlias {
		return nil, false, fmt.Errorf("unknown import header normalization mode %q", mode)
	}
	policy, err := ImportNormalizationPolicyEvidence()
	capability := ImportHeaderPlainCapability
	if mode == ImportHeaderNormalizationNamedAlias {
		policy, err = ImportHeaderNormalizationPolicyEvidence()
		capability = ImportHeaderNamedAliasCapability
	}
	if err != nil {
		return nil, false, err
	}
	if policy.InputSubjectKind != sourcepolicy.SubjectKindFile || policy.SourceDigest == "" || policy.SemanticDigest == "" ||
		!policy.UsedInputFact || !policy.GeneratedOutputFact || policy.HeaderCapability != capability {
		return nil, false, fmt.Errorf("import normalization header capability is not proven")
	}
	if file == nil || group == nil || group.Tok != token.IMPORT || group.Doc != nil || len(group.Specs) < 2 || len(specs) < 2 {
		return nil, false, nil
	}
	if importHeaderHasComments(file, group) {
		return nil, false, nil
	}
	groupSpecs := make(map[*ast.ImportSpec]bool, len(group.Specs))
	for _, raw := range group.Specs {
		spec, ok := raw.(*ast.ImportSpec)
		if !ok || spec.Path == nil {
			return nil, false, nil
		}
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path == "" || path == "C" || strings.HasPrefix(path, "/") || strings.Contains(path, "\\") || strings.Contains(path, "..") {
			return nil, false, nil
		}
		if mode == ImportHeaderNormalizationPlain && spec.Name != nil {
			return nil, false, nil
		}
		if mode == ImportHeaderNormalizationNamedAlias && spec.Name != nil && (spec.Name.Name == "." || spec.Name.Name == "_") {
			return nil, false, nil
		}
		groupSpecs[spec] = true
	}
	ordered := append([]*ast.ImportSpec{}, specs...)
	for _, spec := range ordered {
		if spec == nil || !groupSpecs[spec] {
			return nil, false, nil
		}
	}
	sort.SliceStable(ordered, func(left, right int) bool { return ordered[left].Pos() < ordered[right].Pos() })
	normalized := make([]*ast.GenDecl, 0, len(ordered))
	for _, spec := range ordered {
		copyGroup := *group
		copyGroup.Doc = nil
		copyGroup.Specs = []ast.Spec{spec}
		copyGroup.Lparen, copyGroup.Rparen = token.NoPos, token.NoPos
		normalized = append(normalized, &copyGroup)
	}
	return normalized, true, nil
}

func importHeaderHasComments(file *ast.File, group *ast.GenDecl) bool {
	if group.Doc != nil {
		return true
	}
	for _, raw := range group.Specs {
		spec, ok := raw.(*ast.ImportSpec)
		if !ok || spec.Doc != nil || spec.Comment != nil {
			return true
		}
	}
	for _, comments := range file.Comments {
		if comments.End() >= group.Pos() && comments.Pos() <= group.End() {
			return true
		}
		if comments.End() < group.Pos() && group.Pos()-comments.End() <= 2 {
			return true
		}
	}
	return false
}
