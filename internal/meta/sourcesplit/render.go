package sourcesplit

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"sort"
)

func renderPart(fset *token.FileSet, file *ast.File, declarations []ast.Decl) ([]byte, error) {
	importDecls, imports, err := importsFor(file, declarations)
	if err != nil {
		return nil, err
	}
	allDecls := make([]ast.Decl, 0, len(importDecls)+len(declarations))
	allDecls = append(allDecls, importDecls...)
	allDecls = append(allDecls, declarations...)
	part := *file
	part.Decls = allDecls
	part.Imports = imports
	part.Comments = commentsForPart(fset, file, &part)
	var output bytes.Buffer
	if err := format.Node(&output, fset, &part); err != nil {
		return nil, err
	}
	if output.Len() != 0 && output.Bytes()[output.Len()-1] != '\n' {
		output.WriteByte('\n')
	}
	return output.Bytes(), nil
}

func PhysicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func parseSource(target, subject string) (os.FileInfo, *token.FileSet, *ast.File, error) {
	source, err := os.ReadFile(target)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read %s: %w", subject, err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("stat %s: %w", subject, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, subject, source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse %s: %w", subject, err)
	}
	return info, fset, file, nil
}

func sourceDeclarations(file *ast.File) []ast.Decl {
	declarations := make([]ast.Decl, 0, len(file.Decls))
	for _, declaration := range file.Decls {
		general, isImport := declaration.(*ast.GenDecl)
		if !isImport || general.Tok != token.IMPORT {
			declarations = append(declarations, declaration)
		}
	}
	return declarations
}

func positionLine(fset *token.FileSet, declaration ast.Decl) int {
	return fset.Position(declaration.Pos()).Line
}

func commentsForPart(fset *token.FileSet, file, part *ast.File) []*ast.CommentGroup {
	comments := ast.NewCommentMap(fset, file, file.Comments).Filter(part).Comments()
	seen := make(map[*ast.CommentGroup]bool, len(comments))
	appendComment := func(group *ast.CommentGroup) {
		if group != nil && !seen[group] {
			comments = append(comments, group)
			seen[group] = true
		}
	}
	for _, group := range comments {
		seen[group] = true
	}
	for _, declaration := range part.Decls {
		partGroup, ok := declaration.(*ast.GenDecl)
		if !ok || partGroup.Tok != token.IMPORT {
			continue
		}
		appendComment(partGroup.Doc)
		for _, raw := range partGroup.Specs {
			spec, ok := raw.(*ast.ImportSpec)
			if !ok {
				continue
			}
			appendComment(spec.Doc)
			appendComment(spec.Comment)
		}
		for _, declaration := range file.Decls {
			sourceGroup, ok := declaration.(*ast.GenDecl)
			if !ok || sourceGroup.Tok != token.IMPORT || sourceGroup.Pos() != partGroup.Pos() {
				continue
			}
			if sourceGroup.Doc != nil {
				appendComment(sourceGroup.Doc)
			}
		}
	}
	for _, group := range file.Comments {
		if group.End() < file.Package && !seen[group] {
			comments = append(comments, group)
		}
	}
	sort.Slice(comments, func(i, j int) bool { return comments[i].Pos() < comments[j].Pos() })
	return comments
}
