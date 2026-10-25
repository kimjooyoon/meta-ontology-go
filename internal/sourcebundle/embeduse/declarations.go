package embeduse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

type declaration struct {
	consumer Consumer
	source   []byte
	ranges   [][2]int
}

func (o *observer) declarations(selected listedPackage) ([]declaration, map[string]string, error) {
	var result []declaration
	masked := maps.Clone(o.projections)
	groups := []struct {
		kind  string
		files []string
	}{
		{"production", append(slices.Clone(selected.GoFiles), selected.CgoFiles...)},
		{"internal-test", selected.TestGoFiles}, {"external-test", selected.XTestGoFiles},
	}
	seen := map[string]bool{}
	for _, group := range groups {
		for _, name := range group.files {
			if seen[name] {
				return nil, nil, fmt.Errorf("Go file selected in multiple contexts")
			}
			seen[name] = true
			raw, index, err := o.read(name)
			if err != nil {
				return nil, nil, err
			}
			items, blanked, err := parseDeclarations(name, raw, index, group.kind)
			if err != nil {
				return nil, nil, err
			}
			result = append(result, items...)
			if len(result) > maxConsumers {
				return nil, nil, fmt.Errorf("Go embed consumer count exceeds limit")
			}
			address, err := o.address(name)
			if err != nil {
				return nil, nil, err
			}
			masked[address], err = o.backing(blanked)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return result, masked, nil
}

func parseDeclarations(name string, raw []byte, index int, kind string) ([]declaration, []byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, raw, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	blanked := slices.Clone(raw)
	all, used := 0, 0
	for _, group := range file.Comments {
		for _, comment := range embedComments(group) {
			all++
			for at := fset.PositionFor(comment.Pos(), false).Offset; at < fset.PositionFor(comment.End(), false).Offset; at++ {
				if blanked[at] != '\n' && blanked[at] != '\r' {
					blanked[at] = ' '
				}
			}
		}
	}
	var result []declaration
	for _, node := range file.Decls {
		group, ok := node.(*ast.GenDecl)
		if !ok || group.Tok != token.VAR {
			continue
		}
		for _, spec := range group.Specs {
			value := spec.(*ast.ValueSpec)
			comments := embedComments(value.Doc)
			if len(group.Specs) == 1 {
				comments = append(embedComments(group.Doc), comments...)
			}
			if len(comments) == 0 {
				continue
			}
			if len(value.Names) != 1 || value.Type == nil {
				return nil, nil, fmt.Errorf("embed directive has no exact variable declaration")
			}
			item := declaration{source: raw, consumer: Consumer{Name: value.Names[0].Name, Kind: kind, GoSource: index}}
			item.consumer.Declaration = physicalSpan(fset, comments[0].Pos(), value.End())
			typeSpan := physicalSpan(fset, value.Type.Pos(), value.Type.End())
			item.consumer.DeclaredType = string(raw[typeSpan.Start:typeSpan.End])
			for _, comment := range comments {
				span := physicalSpan(fset, comment.Pos(), comment.End())
				item.ranges = append(item.ranges, [2]int{span.Start, span.End})
			}
			used += len(comments)
			result = append(result, item)
		}
	}
	if all != used {
		return nil, nil, fmt.Errorf("embed directive could not be bound to one variable")
	}
	return result, blanked, nil
}

func embedComments(group *ast.CommentGroup) []*ast.Comment {
	var result []*ast.Comment
	if group == nil {
		return result
	}
	for _, comment := range group.List {
		text := comment.Text
		if text == "//go:embed" || strings.HasPrefix(text, "//go:embed ") || strings.HasPrefix(text, "//go:embed\t") {
			result = append(result, comment)
		}
	}
	return result
}

func physicalSpan(fset *token.FileSet, start, end token.Pos) bidir.SourceSpan {
	first, last := fset.PositionFor(start, false), fset.PositionFor(end, false)
	return bidir.SourceSpan{File: first.Filename, Start: first.Offset, End: last.Offset,
		StartLine: first.Line, StartColumn: first.Column, EndLine: last.Line, EndColumn: last.Column}
}

func (o *observer) observeDeclaration(original listedPackage, item declaration, masked map[string]string) error {
	replace := maps.Clone(masked)
	name := item.consumer.Declaration.File
	address, err := o.address(name)
	if err != nil {
		return err
	}
	raw, err := readBounded(masked[address])
	if err != nil {
		return err
	}
	for _, span := range item.ranges {
		copy(raw[span[0]:span[1]], item.source[span[0]:span[1]])
	}
	replace[address], err = o.backing(raw)
	if err != nil {
		return err
	}
	selected, err := o.list(replace)
	if err != nil {
		return err
	}
	if !sameGoSelection(original, selected) {
		return fmt.Errorf("isolated embed query changed selected Go inputs")
	}
	members := selected.EmbedFiles
	if item.consumer.Kind == "internal-test" {
		members = selected.TestEmbedFiles
	}
	if item.consumer.Kind == "external-test" {
		members = selected.XTestEmbedFiles
	}
	item.consumer.AllMemberCount = len(members)
	for _, member := range members {
		if !strings.HasSuffix(member, ".gooo") && !strings.HasSuffix(member, ".gooo.fixture") {
			continue
		}
		_, index, err := o.read(member)
		if err != nil {
			return err
		}
		item.consumer.Members = append(item.consumer.Members, index)
	}
	o.report.Consumers = append(o.report.Consumers, item.consumer)
	return nil
}

func sameGoSelection(a, b listedPackage) bool {
	return a.Dir == b.Dir && a.ImportPath == b.ImportPath && slices.Equal(a.GoFiles, b.GoFiles) &&
		slices.Equal(a.CgoFiles, b.CgoFiles) && slices.Equal(a.TestGoFiles, b.TestGoFiles) &&
		slices.Equal(a.XTestGoFiles, b.XTestGoFiles) && slices.Equal(a.IgnoredGoFiles, b.IgnoredGoFiles)
}
