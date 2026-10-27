// Package sourcepartition preserves exact source bytes across a declaration cut.
// Callers resolve semantic identity and verify their complete consumer image.
package sourcepartition

import (
	"crypto/sha256"
	"fmt"
	"path"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Slice records bytes copied from the original source, including duplicated headers.
type Slice struct {
	InputStart  int `json:"input_start_byte"`
	InputEnd    int `json:"input_end_byte"`
	OutputStart int `json:"output_start_byte"`
	OutputEnd   int `json:"output_end_byte"`
}

type File struct {
	Filename string  `json:"filename"`
	SHA256   string  `json:"sha256"`
	Source   string  `json:"source"`
	Origins  []Slice `json:"origins"`
}

// SplitAt checks that cut is the physical start of a non-first declaration.
// It preserves comments, bodies, bindings and spacing without reformatting them.
// The resulting files require joint semantic validation in their original scope.
func SplitAt(filename, source, into string, cut int) ([]File, error) {
	return SplitAtWithEntityFieldsSupport(filename, source, into, cut, syntax.EntityFieldsV4Support())
}

// SplitAtWithEntityFieldsSupport checks the physical boundary under the exact
// caller-selected profile. Semantic and consumer validation remain separate.
func SplitAtWithEntityFieldsSupport(filename, source, into string, cut int, support syntax.EntityFieldsSupport) ([]File, error) {
	if into == "" || path.Clean(into) != into || path.IsAbs(into) || path.Ext(into) != ".gooo" ||
		strings.ContainsAny(into, "\\\x00:") || strings.HasPrefix(into, "../") ||
		strings.EqualFold(path.Clean(filename), into) {
		return nil, fmt.Errorf("split destination must be a different canonical relative .gooo filename")
	}
	file, header, err := partitionSyntaxWithEntityFieldsSupport(filename, source, support)
	if err != nil {
		return nil, err
	}
	matched := false
	for _, declaration := range file.Decls[1:] {
		matched = matched || declaration.SourceSpan().Start.Offset == cut
	}
	if !matched || header <= 0 || cut <= header || cut >= len(source) {
		return nil, fmt.Errorf("split boundary must precede a non-first declaration")
	}
	left, right := source[:cut], source[:header]+source[cut:]
	return []File{
		{Filename: filename, SHA256: digest(left), Source: left, Origins: []Slice{{0, cut, 0, cut}}},
		{Filename: into, SHA256: digest(right), Source: right,
			Origins: []Slice{{0, header, 0, header}, {cut, len(source), header, len(right)}}},
	}, nil
}

func partitionSyntax(filename, source string) (*syntax.File, int, error) {
	return partitionSyntaxWithEntityFieldsSupport(filename, source, syntax.EntityFieldsV4Support())
}

func partitionSyntaxWithEntityFieldsSupport(filename, source string, support syntax.EntityFieldsSupport) (*syntax.File, int, error) {
	if err := support.Validate(); err != nil {
		return nil, 0, err
	}
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, source, support)
	if file == nil || diagnostics.HasErrors() || len(file.Decls) == 0 {
		return nil, 0, fmt.Errorf("split requires valid source with declarations")
	}
	header := file.Decls[0].SourceSpan().Start.Offset
	if err := validateHeader(file, header); err != nil {
		return nil, 0, err
	}
	return file, header, nil
}

func validateHeader(file *syntax.File, header int) error {
	if file.Package == nil || file.Namespace == nil || file.Package.Span.End.Offset > header || file.Namespace.Span.End.Offset > header {
		return fmt.Errorf("split requires a leading package and namespace header")
	}
	for _, imp := range file.Imports {
		if imp.Span.End.Offset > header {
			return fmt.Errorf("split requires imports before declarations to retain file-local aliases")
		}
	}
	for _, binding := range file.Bindings {
		if binding.Span.Start.Offset < header {
			return fmt.Errorf("split cannot duplicate a binding in the source header")
		}
	}
	return nil
}

func digest(source string) string { return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(source))) }
